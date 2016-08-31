package main

import (
    "fmt"
    "net/http"
    "io"
    "io/ioutil"
    "net/url"
    "encoding/json"
    "os"
)

type LoungeTokenScreenList struct {
    Screens []LoungeTokenScreenItem
}

type LoungeTokenScreenItem struct {
    ScreenId string
    LoungeToken string
    Expiration uint64
}

const (
    screenName string = "Golang Test TV"
    screenApp string = "golang-test-838"
    screenUid string = "2a026ce9-4429-4c5e-8ef5-0101eddf5671"
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
    fmt.Println("lounge_token", tokenScreenItem.LoungeToken, tokenScreenItem.Expiration / 1000)

    // pairing code:
    resp, err = http.PostForm("https://www.youtube.com/api/lounge/pairing/get_pairing_code?ctx=pair", url.Values{
        "access_type": {"permanent"},
        "app": {screenApp},
        "lounge_token": {tokenScreenItem.LoungeToken},
        "screen_id": {tokenScreenItem.ScreenId},
        "screen_name": {screenName},
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

    // bind:
    for {
        optSID := "foo"
        optAID := 42
        optGsessionid := "bar"
        optZX := "xxxxxxxxxxxx"
        resp, err = http.Get("https://www.youtube.com/api/lounge/bc/bind?device=LOUNGE_SCREEN&id=" + screenUid + "&name=" + url.QueryEscape(screenName) + "&app=" + screenApp + "&theme=cl&capabilities&mdx-version=2&loungeIdToken=" + tokenScreenItem.LoungeToken + "&VER=8&v=2&RID=rpc&SID=" + optSID + "&CI=0&AID=" + string(optAID) + "&gsessionid=" + optGsessionid + "&TYPE=xmlhttp&zx=" + optZX + "&t=1")
        if err != nil {
            fmt.Println(err.Error())
        }
        fmt.Println("==response==")
        io.Copy(os.Stdout, resp.Body)
        resp.Body.Close()
    }
}
