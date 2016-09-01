package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type LoungeTokenScreenList struct {
	Screens []LoungeTokenScreenItem
}

type LoungeTokenScreenItem struct {
	ScreenId    string
	LoungeToken string
	Expiration  uint64
}

const (
	screenName string = "Golang Test TV"
	screenApp  string = "golang-test-838"
	screenUid  string = "2a026ce9-4429-4c5e-8ef5-0101eddf5671"
)

var (
	sid        string
	gsessionid string
)

func main() {
	// screen id:
	resp, err := http.Get("https://www.youtube.com/api/lounge/pairing/generate_screen_id")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	screenId := string(body)
	fmt.Println("screen_id", screenId)

	// lounge token:
	resp, err = http.PostForm("https://www.youtube.com/api/lounge/pairing/get_lounge_token_batch", url.Values{"screen_ids": {screenId}})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	tokenObj := new(LoungeTokenScreenList)
	err = json.Unmarshal(body, &tokenObj)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	tokenScreenItem := tokenObj.Screens[0]
	fmt.Println("lounge_token", tokenScreenItem.LoungeToken, tokenScreenItem.Expiration/1000)

	// pairing code:
	resp, err = http.PostForm("https://www.youtube.com/api/lounge/pairing/get_pairing_code?ctx=pair", url.Values{
		"access_type":  {"permanent"},
		"app":          {screenApp},
		"lounge_token": {tokenScreenItem.LoungeToken},
		"screen_id":    {tokenScreenItem.ScreenId},
		"screen_name":  {screenName},
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	pairCode := string(body)
	fmt.Printf("pairing_code %s-%s-%s-%s\n", pairCode[0:3], pairCode[3:6], pairCode[6:9], pairCode[9:12])

	// bind 1
	bindVals := url.Values{
		"device":        {"LOUNGE_SCREEN"},
		"id":            {screenUid},
		"name":          {screenName},
		"app":           {screenApp},
		"theme":         {"cl"},
		"capabilities":  {},
		"mdx-version":   {"2"},
		"loungeIdToken": {tokenScreenItem.LoungeToken},
		"VER":           {"13"},
		"v":             {"2"},
		"RID":           {"1337"},
		"CI":            {"0"},
		"AID":           {"42"},
		"TYPE":          {"xmlhttp"},
		"zx":            {"xxxxxxxxxxxx"},
		"t":             {"1"},
	}
	resp, err = http.PostForm("https://www.youtube.com/api/lounge/bc/bind?"+bindVals.Encode(), url.Values{"count": {"0"}})
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
	decodeBindStream(resp.Body)

	// bind:
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
		"RID":           {"rpc"},
		"SID":           {sid},
		"CI":            {"0"},
		"AID":           {"42"},
		"gsessionid":    {gsessionid},
		"TYPE":          {"xmlhttp"},
		"zx":            {"xxxxxxxxxxxx"},
		"t":             {"1"},
	}
	resp, err = http.Get("https://www.youtube.com/api/lounge/bc/bind?" + bindVals.Encode())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	decodeBindStream(resp.Body)
}

func decodeBindStream(r io.Reader) {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("error:", err.Error())
			return
		}
		//fmt.Printf("%T: %v", t, t)
		switch t.(type) {
		case json.Number:
			// length, ignore
		case json.Delim:
			for dec.More() {
				var indexedCmd []interface{}
				err = dec.Decode(&indexedCmd)
				if err != nil {
					fmt.Println("error:", err.Error())
					return
				}
				cmdArray := indexedCmd[1].([]interface{})
				genericCmd(cmdArray[0].(string), cmdArray[1:])
			}
			// closing ]:
			dec.Token()
		}
	}
}

func genericCmd(cmd string, paramsList []interface{}) {
	switch cmd {
	case "noop":
		fmt.Println("noop")
	case "c":
		sid = paramsList[0].(string)
		fmt.Println("option_sid", sid)
	case "S":
		gsessionid = paramsList[0].(string)
		fmt.Println("option_gsessionid", gsessionid)
	default:
		fmt.Printf("generic_cmd %s %+v\n", cmd, paramsList)
	}
}
