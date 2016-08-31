package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "net/url"
    "encoding/json"
)

type LoungeTokenScreenList struct {
    Screens []LoungeTokenScreenItem
}

type LoungeTokenScreenItem struct {
    ScreenId string
    LoungeToken string
    Expiration uint64
}

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
        "app": {"golang-test-838"},
        "lounge_token": {tokenScreenItem.LoungeToken},
        "screen_id": {tokenScreenItem.ScreenId},
        "screen_name": {"Golang Test TV"},
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
}
