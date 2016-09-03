package main

// TODO documentation, modularity

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
)

var (
	debugEnabled  bool
	screenName    string
	screenApp     string
	bindVals      url.Values
	currentVolume string = "100"
	ofs           uint64 = 0
	playState     string = "3"

	// these two vars are used to determine the current playing time
	startTime time.Time
	curTime   time.Duration

	curVideoId    string
	curVideo      Video
	curListId     string
	curList       []string // Array of video IDs
	curIndex      int
	curListVideos []Video
)

func init() {
	flag.BoolVar(&debugEnabled, "d", false, "Enable debug information (including full cmd info)")
	flag.StringVar(&screenName, "n", defaultScreenName, "Display Name")
	flag.StringVar(&screenApp, "i", defaultScreenApp, "Display App")
}

func main() {
	flag.Parse()
	// screen id:
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
	screenId := string(body)
	fmt.Println("screen_id", screenId)

	// lounge token:
	resp, err = http.PostForm("https://www.youtube.com/api/lounge/pairing/get_lounge_token_batch", url.Values{"screen_ids": {screenId}})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	tokenObj := new(LoungeTokenScreenList)
	err = json.Unmarshal(body, &tokenObj)
	if err != nil {
		panic(err)
	}
	tokenScreenItem := tokenObj.Screens[0]
	fmt.Println("lounge_token", tokenScreenItem.LoungeToken, tokenScreenItem.Expiration/1000)

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

	// pairing code:
	resp, err = http.PostForm("https://www.youtube.com/api/lounge/pairing/get_pairing_code?ctx=pair", url.Values{
		"access_type":  {"permanent"},
		"app":          {screenApp},
		"lounge_token": {tokenScreenItem.LoungeToken},
		"screen_id":    {tokenScreenItem.ScreenId},
		"screen_name":  {screenName},
	})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	pairCode := string(body)
	fmt.Printf("pairing_code %s-%s-%s-%s\n", pairCode[0:3], pairCode[3:6], pairCode[6:9], pairCode[9:12])

	// bind:
	for {
		ofs++
		bindValsGet := bindVals
		bindValsGet["RID"] = []string{"rpc"}
		bindValsGet["CI"] = []string{"0"}
		resp, err = http.Get("https://www.youtube.com/api/lounge/bc/bind?" + bindValsGet.Encode())
		if err != nil {
			panic(err)
		}
		decodeBindStream(resp.Body)
		resp.Body.Close()
	}
}

// decodeBindStream takes an io.Reader (e.g. bind response body) and parses the command stream
func decodeBindStream(r io.Reader) {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		switch t.(type) {
		case json.Number:
			// length, ignore
		case json.Delim:
			for dec.More() {
				var indexedCmd []interface{}
				err = dec.Decode(&indexedCmd)
				if err != nil {
					panic(err)
				}
				cmdArray := indexedCmd[1].([]interface{})
				genericCmd(cmdArray[0].(string), cmdArray[1:])
			}
			// closing ]:
			dec.Token()
		}
	}
}

// genericCmd interpretes and executes commands from the bind stream
func genericCmd(cmd string, paramsList []interface{}) {
	if debugEnabled {
		debugInfo()
		fmt.Printf("dbg_raw_cmd %v %#v\n", cmd, paramsList)
	}

	switch cmd {
	case "noop":
		fmt.Println("noop")
	case "c":
		sid := paramsList[0].(string)
		bindVals["SID"] = []string{sid}
		fmt.Println("option_sid", sid)
	case "S":
		gsessionid := paramsList[0].(string)
		bindVals["gsessionid"] = []string{gsessionid}
		fmt.Println("option_gsessionid", gsessionid)
	case "remoteConnected":
		data := paramsList[0].(map[string]interface{})
		id := data["id"].(string)
		name := data["name"].(string)
		fmt.Println("remote_join", id, name)
	case "remoteDisconnected":
		data := paramsList[0].(map[string]interface{})
		id := data["id"].(string)
		fmt.Println("remote_leave", id)
	case "getNowPlaying":
	case "setPlaylist":
		data := paramsList[0].(map[string]interface{})
		curVideoId = data["videoId"].(string)
		curListId = data["listId"].(string)
		info := getListInfo(curListId)
		curListVideos = info.Video
		currentTime := data["currentTime"].(string)
		videoIds := data["videoIds"].(string)
		curList = strings.Split(videoIds, ",")
		curIndex, err := strconv.Atoi(data["currentIndex"].(string))
		if err != nil {
			curIndex = 0
		}
		curVideo = curListVideos[curIndex]

		// set startTime:
		currentTimeDuration, err := time.ParseDuration(currentTime + "s")
		if err != nil {
			currentTimeDuration = 0
		}
		curTime = currentTimeDuration
		startTime = time.Now().Add(-curTime)

		fmt.Println("video_id", curVideoId)
		postBind("nowPlaying", map[string]string{
			"videoId":      curVideoId,
			"currentTime":  currentTime,
			"ctt":          data["ctt"].(string),
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
		fmt.Println("play")
		playState = "1"
		startTime = time.Now().Add(-curTime)
		postBind("onStateChange", map[string]string{
			"currentTime": fmt.Sprintf("%.3f", curTime.Seconds()),
			"state":       "1",
			"duration":    strconv.Itoa(curVideo.Length),
			"cpn":         "foo",
		})
	case "pause":
		fmt.Println("pause")
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
		fmt.Println("set_volume", currentVolume)
		postBind("onVolumeChanged", map[string]string{"volume": currentVolume, "muted": "false"})
	case "seekTo":
		data := paramsList[0].(map[string]interface{})
		newTime := data["newTime"].(string)
		fmt.Println("seek_to", newTime)
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
		fmt.Println("stop")
		postBind("nowPlaying", map[string]string{})
	case "onUserActivity":
		fmt.Println("user_action")
	case "next":
		fmt.Println("next")
		if curIndex+1 < len(curList) {
			curIndex++
			curTime = 0
			startTime = time.Now()
			curVideoId = curList[curIndex]
			curVideo = curListVideos[curIndex]
			fmt.Println("video_id", curVideoId)
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
		fmt.Println("previous")
		if curIndex > 0 {
			curIndex--
			curTime = 0
			startTime = time.Now()
			curVideoId = curList[curIndex]
			curVideo = curListVideos[curIndex]
			fmt.Println("video_id", curVideoId)
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
		fmt.Printf("generic_cmd %s %v\n", cmd, paramsList)
	}
	if debugEnabled {
		debugInfo()
	}
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
	fmt.Printf(
		"dbg_info curVideoId=%v curListId=%v curList=%v curIndex=%v curTime=%.3f curVolume=%v curState=%v curListVideos=%v curVideo=%v\n",
		curVideoId,
		curListId,
		curList,
		curIndex, time.Now().Sub(startTime).Seconds(),
		currentVolume,
		playState,
		curListVideos,
		curVideo,
	)
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
