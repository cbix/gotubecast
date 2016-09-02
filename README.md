# gotubecast
gotubecast is a small Go program which you can use to make your own YouTube TV player.

It connects to the YouTube Leanback API and generates a text stream providing pairing codes, video IDs,
play/pause/seek/volume change commands etc. It doesn't have any dependencies and runs on any of the platforms supported by golang.
For example, use it on a Raspberry Pi in combination with youtube-dl and omxplayer for a DIY Chromecast clone or make a YouTube TV
extension for your favorite media center software.

## Build + Install
Provided you have golang correctly set up:

    go get github.com/CBiX/gotubecast

## Run
With default options:

    gotubecast

Give it a name:

    gotubecast -n "Dumb TV" -i dumb-v1 

## Text stream
The following keys are being written to stdout:
### Essential methods
* **pairing_code \<aaa-bbb-ccc-ddd\>**: the device pairing code formatted with separating dashes
* **video\_id \<id\>**
* **play**
* **pause**
* **seek\_to \<seconds\>**
* **set\_volume \<percent\>**

### Other
* **noop**: do nothing
* **generic\_cmd \<cmd\> \<params\>**: all non-implemented commands
* **remote\_join \<id\> \<name\>**: client connects
* **remote\_leave \<id\>**: client disconnects
* **next**
* **previous**
* **screen\_id, lounge\_token, option\_sid, option\_gsessionid**: these are needed by the API

## Roadmap / TODO
* testing
* provide sample bash script for Raspberry Pi
* video duration
* autoplay
* subtitles
