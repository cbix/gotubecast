# gotubecast
gotubecast is a small Go program which you can use to make your own YouTube TV player.

It connects to the YouTube Leanback API and generates a text stream providing pairing codes, video IDs,
play/pause/seek/volume change commands etc. It doesn't have any dependencies and runs on any of the platforms supported by golang.
For example, use it on a Raspberry Pi in combination with youtube-dl and omxplayer for a DIY Chromecast clone or make a YouTube TV
extension for your favorite media center software.

## Build + Install
Provided you have [golang correctly set up](https://golang.org/doc/install):

    go get github.com/CBiX/gotubecast

## Run
With default options:

    gotubecast

Minimal dumb YouTube TV example (opens every video in a new browser window, no control possible):

    gotubecast -n "Dumb TV" -i dumb-v1 | while read line
    do
        cmd="`cut -d ' ' -f1 <<< "$line"`"
        arg="`cut -d ' ' -f2 <<< "$line"`"
        case "$cmd" in
            pairing_code)
                echo "Your pairing code: $arg"
                ;;
            remote_join)
                cut -d ' ' -f3- <<< "$line connected"
                ;;
            video_id)
                xdg-open "https://www.youtube.com/watch?v=$arg" &
                ;;
        esac
    done

Usage help:

	$ gotubecast -h
	Usage of ./gotubecast:
	  -d int
			Debug information level. 0 = off; 1 = full cmd info; 2 = timestamp prefix, this changes the output format!
	  -i string
			Display App (default "golang-test-838")
	  -n string
			Display Name (default "Golang Test TV")
	  -s string
			Screen ID (will be generated if empty)

More in the examples folder.

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
* **screen\_id**: The screen ID will be generated if not passed by -s flag. If you want to keep connected devices over restarts, generate it first and pass it from then on.
* **lounge\_token, option\_sid, option\_gsessionid**: API internals

## Roadmap / TODO
* testing
* provide sample bash script for Raspberry Pi
* video duration
* autoplay
* subtitles
