package main

// TODO documentation, modularity (library + main),
// handle commands by chan selects,
// "offline" message sent if no command comes in within x seconds (x = 40?),
// run forever, even if offline,
// keep player state, devices etc.

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LoungeTokenScreenList struct {
	Screens []LoungeTokenScreenItem
}

type LoungeTokenScreenItem struct {
	ScreenId    string
	LoungeToken string
	Expiration  uint64
}

type Video struct {
	Id        string `json:"encrypted_id"`
	Length    int    `json:"length_seconds"`
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
}

type PlaylistInfo struct {
	Video []Video
}

const (
	defaultScreenName string = "Golang Test TV"
	defaultScreenApp  string = "golang-test-838"
	screenUid         string = "2a026ce9-4429-4c5e-8ef5-0101eddf5671"
	timefmt           string = "[2006-01-02 15:04:05]"
	errCountThresh    int    = 5
)

var (
	debugLevel    int
	screenId      string
	screenName    string
	screenApp     string
	bindVals      url.Values
	currentVolume string = "100"
	ofs           uint64 = 0
	playState     string = "3"
	ctt           string
	//playTimer     *time.Timer
	currentCmdIndex int64

	// these two vars are used to determine the current playing time
	startTime time.Time
	curTime   time.Duration

	curVideoId    string
	curVideo      Video
	curListId     string
	curList       []string // Array of video IDs
	curIndex      int
	curListVideos []Video
	printLock     sync.Mutex
)

func init() {
	flag.IntVar(&debugLevel, "d", 0, "Debug information level. 0 = off; 1 = full cmd info; 2 = timestamp prefix")
	flag.StringVar(&screenName, "n", defaultScreenName, "Display Name")
	flag.StringVar(&screenApp, "i", defaultScreenApp, "Display App")
	flag.StringVar(&screenId, "s", "", "Screen ID (will be generated if empty)")
}

func main() {
	flag.Parse()
	// screen id:
	if screenId == "" {
		resp, err := http.Get("https://www.youtube.com/api/lounge/pairing/generate_screen_id")
		if err != nil {
			panic(err)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		screenId = string(body)
	}
	msgPrintln(fmt.Sprint("screen_id ", screenId))

	// lounge token:
	resp, err := http.PostForm("https://www.youtube.com/api/lounge/pairing/get_lounge_token_batch", url.Values{"screen_ids": {screenId}})
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	tokenObj := new(LoungeTokenScreenList)
	err = json.Unmarshal(body, &tokenObj)
	if err != nil {
		panic(err)
	}
	tokenScreenItem := tokenObj.Screens[0]
	msgPrintln(fmt.Sprint("lounge_token ", tokenScreenItem.LoungeToken, " ", tokenScreenItem.Expiration/1000))

	bindVals = url.Values{
		"device":        {"LOUNGE_SCREEN"},
		"id":            {screenUid},
		"name":          {screenName},
		"app":           {screenApp},
		"theme":         {"cl"},
		"capabilities":  {},
		"mdx-version":   {"2"},
		"loungeIdToken": {tokenScreenItem.LoungeToken},
		"VER":           {"8"},
		"v":             {"2"},
		"RID":           {"1337"},
		"AID":           {"42"},
		"zx":            {"xxxxxxxxxxxx"},
		"t":             {"1"},
	}

	// bind 1
	resp, err = http.PostForm("https://www.youtube.com/api/lounge/bc/bind?"+bindVals.Encode(), url.Values{"count": {"0"}})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	decodeBindStream(resp.Body)

	// pairing code every 5 minutes:
	go func() {
		for {
			resp, err = http.PostForm("https://www.youtube.com/api/lounge/pairing/get_pairing_code?ctx=pair", url.Values{
				"access_type":  {"permanent"},
				"app":          {screenApp},
				"lounge_token": {tokenScreenItem.LoungeToken},
				"screen_id":    {tokenScreenItem.ScreenId},
				"screen_name":  {screenName},
			})
			if err != nil {
				msgPrintln(fmt.Sprint("error ", err.Error()))
				time.Sleep(10 * time.Second)
				continue
			}
			body, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				msgPrintln(fmt.Sprint("error ", err.Error()))
				time.Sleep(10 * time.Second)
				continue
			}
			pairCode := string(body)
			msgPrintln(fmt.Sprintf("pairing_code %s-%s-%s-%s", pairCode[0:3], pairCode[3:6], pairCode[6:9], pairCode[9:12]))
			time.Sleep(5 * time.Minute)
		}
	}()

	// bind:
	errCount := 0
	for {
		if errCount >= errCountThresh {
			msgPrintln(fmt.Sprintf("error reached %d errors, terminating...", errCount))
			return
		}
		ofs++
		bindValsGet := bindVals
		bindValsGet["RID"] = []string{"rpc"}
		bindValsGet["CI"] = []string{"0"}
		resp, err = http.Get("https://www.youtube.com/api/lounge/bc/bind?" + bindValsGet.Encode())
		if err != nil {
			errCount++
			msgPrintln(fmt.Sprint("error ", err.Error()))
			continue
		}
		err = decodeBindStream(resp.Body)
		resp.Body.Close()
		if err != nil {
			errCount++
			msgPrintln(fmt.Sprint("error ", err.Error()))
			continue
		}
		errCount = 0
	}
}

// decodeBindStream takes an io.Reader (e.g. bind response body) and parses the command stream
func decodeBindStream(r io.Reader) (err error) {
	err = nil
	dec := json.NewDecoder(r)
	dec.UseNumber()
	for {
		var t json.Token
		t, err = dec.Token()
		if err == io.EOF {
			err = nil
			return
		}
		if err != nil {
			return
		}
		switch t.(type) {
		case json.Number:
			// length, ignore
		case json.Delim:
			for dec.More() {
				var indexedCmd []interface{}
				err = dec.Decode(&indexedCmd)
				if err != nil {
					return
				}
				var index int64
				index, err = indexedCmd[0].(json.Number).Int64()
				if err != nil {
					return
				}
				cmdArray := indexedCmd[1].([]interface{})
				genericCmd(index, cmdArray[0].(string), cmdArray[1:])
			}
			// closing ]:
			dec.Token()
		}
	}
	return
}

// genericCmd interpretes and executes commands from the bind stream
func genericCmd(index int64, cmd string, paramsList []interface{}) {
	//debugInfo()
	dbgPrintln(fmt.Sprintf("raw_cmd %v %v %#v", index, cmd, paramsList))
	if currentCmdIndex > 0 && index <= currentCmdIndex {
		dbgPrintln(fmt.Sprintf("skipping already seen cmd %d", index))
		return
	}
	currentCmdIndex = index
	switch cmd {
	case "noop":
		//msgPrintln("noop")
	case "c":
		sid := paramsList[0].(string)
		bindVals["SID"] = []string{sid}
		msgPrintln(fmt.Sprint("option_sid ", sid))
	case "S":
		gsessionid := paramsList[0].(string)
		bindVals["gsessionid"] = []string{gsessionid}
		msgPrintln(fmt.Sprint("option_gsessionid ", gsessionid))
	case "remoteConnected":
		data := paramsList[0].(map[string]interface{})
		id := data["id"].(string)
		name := data["name"].(string)
		msgPrintln(fmt.Sprint("remote_join ", id, " ", name))
	case "remoteDisconnected":
		data := paramsList[0].(map[string]interface{})
		id := data["id"].(string)
		msgPrintln(fmt.Sprint("remote_leave ", id))
	case "getNowPlaying":
		curTime = time.Now().Sub(startTime)
		if curVideoId == "" {
			postBind("nowPlaying", map[string]string{})
		} else {
			postBind("nowPlaying", map[string]string{
				"videoId":      curVideoId,
				"currentTime":  fmt.Sprintf("%.3f", curTime.Seconds()),
				"ctt":          ctt,
				"listId":       curListId,
				"currentIndex": strconv.Itoa(curIndex),
				"state":        playState,
			})
		}
	case "setPlaylist":
		data := paramsList[0].(map[string]interface{})
		curVideoId = data["videoId"].(string)
		curListId = data["listId"].(string)
		info := getListInfo(curListId)
		curListVideos = info.Video
		currentTime := ""
		if data["currentTime"] != nil {
			currentTime = data["currentTime"].(string)
		}
		videoIds := data["videoIds"].(string)
		curList = strings.Split(videoIds, ",")
		curVideo = curListVideos[0]
		if data["currentIndex"] != nil {
			curIndex, err := strconv.Atoi(data["currentIndex"].(string))
			if err == nil {
				curVideo = curListVideos[curIndex]
			}
		}

		// set startTime:
		currentTimeDuration, err := time.ParseDuration(currentTime + "s")
		if err != nil {
			currentTimeDuration = 0
		}
		curTime = currentTimeDuration
		startTime = time.Now().Add(-curTime)
		var ok bool
		ctt, ok = data["ctt"].(string)
		if !ok {
			ctt = ""
		}

		msgPrintln(fmt.Sprint("video_id ", curVideoId))
		postBind("nowPlaying", map[string]string{
			"videoId":      curVideoId,
			"currentTime":  currentTime,
			"ctt":          ctt,
			"listId":       curListId,
			"currentIndex": strconv.Itoa(curIndex),
			"state":        "3",
		})
		playState = "1"
		postBind("onStateChange", map[string]string{
			"currentTime": currentTime,
			"state":       "1",
			"duration":    strconv.Itoa(curVideo.Length),
			"cpn":         "foo",
		})
	case "updatePlaylist":
		data := paramsList[0].(map[string]interface{})
		curListId = data["listId"].(string)
		if data["videoIds"] != nil {
			videoIds := data["videoIds"].(string)
			curList = strings.Split(videoIds, ",")
			if curIndex >= len(curList) {
				curIndex = len(curList) - 1
			}
		} else {
			//empty list
			curList = []string{}
			curIndex = 0
		}
		info := getListInfo(curListId)
		curListVideos = info.Video
	case "play":
		msgPrintln("play")
		playState = "1"
		startTime = time.Now().Add(-curTime)
		postBind("onStateChange", map[string]string{
			"currentTime": fmt.Sprintf("%.3f", curTime.Seconds()),
			"state":       "1",
			"duration":    strconv.Itoa(curVideo.Length),
			"cpn":         "foo",
		})
	case "pause":
		msgPrintln("pause")
		playState = "2"
		curTime = time.Now().Sub(startTime)
		postBind("onStateChange", map[string]string{
			"currentTime": fmt.Sprintf("%.3f", curTime.Seconds()),
			"state":       "2",
			"duration":    strconv.Itoa(curVideo.Length),
			"cpn":         "foo",
		})
	case "getVolume":
		postBind("onVolumeChanged", map[string]string{"volume": currentVolume, "muted": "false"})
	case "setVolume":
		data := paramsList[0].(map[string]interface{})
		currentVolume = data["volume"].(string)
		msgPrintln(fmt.Sprint("set_volume ", currentVolume))
		postBind("onVolumeChanged", map[string]string{"volume": currentVolume, "muted": "false"})
	case "seekTo":
		data := paramsList[0].(map[string]interface{})
		newTime := data["newTime"].(string)
		msgPrintln(fmt.Sprint("seek_to ", newTime))
		// update startTime:
		currentTimeDuration, err := time.ParseDuration(newTime + "s")
		if err != nil {
			currentTimeDuration = 0
		}
		curTime = currentTimeDuration
		startTime = time.Now().Add(-curTime)

		postBind("onStateChange", map[string]string{
			"currentTime": newTime,
			"state":       playState,
			"duration":    strconv.Itoa(curVideo.Length),
			"cpn":         "foo",
		})
	case "stopVideo":
		msgPrintln("stop")
		postBind("nowPlaying", map[string]string{})
	case "onUserActivity":
		msgPrintln("user_action")
	case "next":
		msgPrintln("next")
		if curIndex+1 < len(curList) {
			curIndex++
			curTime = 0
			startTime = time.Now()
			curVideoId = curList[curIndex]
			curVideo = curListVideos[curIndex]
			msgPrintln(fmt.Sprint("video_id ", curVideoId))
			postBind("nowPlaying", map[string]string{
				"videoId":      curVideoId,
				"currentTime":  "0",
				"listId":       curListId,
				"currentIndex": strconv.Itoa(curIndex),
				"state":        "3",
			})
			playState = "1"
			postBind("onStateChange", map[string]string{
				"currentTime": "0",
				"state":       "1",
				"duration":    strconv.Itoa(curVideo.Length),
				"cpn":         "foo",
			})
		}
	case "previous":
		msgPrintln("previous")
		if curIndex > 0 {
			curIndex--
			curTime = 0
			startTime = time.Now()
			curVideoId = curList[curIndex]
			curVideo = curListVideos[curIndex]
			msgPrintln(fmt.Sprint("video_id ", curVideoId))
			postBind("nowPlaying", map[string]string{
				"videoId":      curVideoId,
				"currentTime":  "0",
				"listId":       curListId,
				"currentIndex": strconv.Itoa(curIndex),
				"state":        "3",
			})
			playState = "1"
			postBind("onStateChange", map[string]string{
				"currentTime": "0",
				"state":       "1",
				"duration":    strconv.Itoa(curVideo.Length),
				"cpn":         "foo",
			})
		}
	default:
		msgPrintln(fmt.Sprintf("generic_cmd %s %v", cmd, paramsList))
	}
	/*
		if debugLevel >= 1 {
			debugInfo()
		}
	*/
}

func postBind(sc string, params map[string]string) {
	ofs++
	postVals := url.Values{"count": {"1"}, "ofs": {fmt.Sprintf("%v", ofs)}}
	postVals["req0__sc"] = []string{sc}
	for k, v := range params {
		postVals["req0_"+k] = []string{v}
	}
	bindVals["RID"] = []string{"1337"}
	resp, err := http.PostForm("https://www.youtube.com/api/lounge/bc/bind?"+bindVals.Encode(), postVals)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
}

func debugInfo() {
	dbgPrintln(fmt.Sprintf(
		"info curVideoId=%v curListId=%v curList=%v curIndex=%v curTime=%.3f curVolume=%v curState=%v curListVideos=%v curVideo=%v",
		curVideoId,
		curListId,
		curList,
		curIndex,
		curTime.Seconds(),
		currentVolume,
		playState,
		curListVideos,
		curVideo,
	))
}

func getListInfo(listId string) (ret *PlaylistInfo) {
	resp, err := http.Get("https://www.youtube.com/list_ajax?style=json&action_get_list=1&list=" + listId)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	ret = new(PlaylistInfo)
	err = json.Unmarshal(body, &ret)
	if err != nil {
		panic(err)
	}
	return
}

func msgPrint(line string) {
	printLock.Lock()
	if debugLevel >= 2 {
		fmt.Print(time.Now().Format(timefmt), " ", line)
	} else {
		fmt.Print(line)
	}
	printLock.Unlock()
}

func msgPrintln(line string) {
	printLock.Lock()
	fmt.Println(line)
	printLock.Unlock()
}

func dbgPrintln(line string) {
	if debugLevel >= 1 {
		if debugLevel >= 2 {
			msgPrintln(fmt.Sprint("dbg ", time.Now().Format(timefmt), " ", line))
		} else {
			msgPrintln(fmt.Sprint("dbg ", line))
		}
	}
}
