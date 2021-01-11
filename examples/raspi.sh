#!/bin/bash
# simple YouTube TV for Raspberry Pi
# kills any currently playing video as soon as a new video is queued
# needs youtube-dl (or ytdl) and omxplayer
[ ! -e ".screen_id" ] && wget -O ".screen_id" "https://www.youtube.com/api/lounge/pairing/generate_screen_id"
export SCREEN_ID=$(cat ".screen_id")
export SCREEN_NAME="Raspberry Pi"
export SCREEN_APP="pitubecast-v1"
export OMX_OPTS="-o both"
# export ANNOTATE="+0+245"
export VOL="1.0"

[ ! -z `type -p apt` ] && {
    [ -z `type -p bc` ] && sudo apt install bc -y
    [ -z `type -p omxplayer` ] && sudo apt install omxplayer -y
    [ -z `type -p youtube-dl` ] && [ -z `type -p ytdl` ] && [ -z `type -p jq` ] && sudo apt install jq -y
    [ ! -z "$ANNOTATE" ] && [ ! -z `type -p nitrogen` ] && [ -e "background.png" ] && [ -z `type -p convert` ] && sudo apt install imagemagick -y
}

function omxdbus {
    OMXPLAYER_DBUS_ADDR="/tmp/omxplayerdbus.${USER:-root}"
    OMXPLAYER_DBUS_PID="/tmp/omxplayerdbus.${USER:-root}.pid"
    export DBUS_SESSION_BUS_ADDRESS=`cat $OMXPLAYER_DBUS_ADDR`
    export DBUS_SESSION_BUS_PID=`cat $OMXPLAYER_DBUS_PID`
    dbus-send --print-reply=literal --session --reply-timeout=100 --dest=org.mpris.MediaPlayer2.omxplayer /org/mpris/MediaPlayer2 $*
}
function urldecode {
    echo -e "$(sed 's/+/ /g;s/%\(..\)/\\x\1/g;')"
}

gotubecast -s "$SCREEN_ID" -n "$SCREEN_NAME" -i "$SCREEN_APP" | while read line
do
    cmd="`cut -d ' ' -f1 <<< "$line"`"
    arg="`cut -d ' ' -f2 <<< "$line"`"
    case "$cmd" in
        pairing_code)
            echo "Your pairing code: $arg"
            [ ! -z "$ANNOTATE" ] && [ ! -z `type -p nitrogen` ] && [ -e "background.png" ] && {
                convert "background.png" -gravity North -pointsize 30 -fill white -annotate "$ANNOTATE" "$arg" "/tmp/code.png"
                sudo nitrogen --set-centered "/tmp/code.png"
            }
            ;;
        remote_join)
            cut -d ' ' -f3- <<< "$line connected"
            ;;
        video_id)
            killall omxplayer.bin
            [ ! -z `type -p youtube-dl` ] && { # http://rg3.github.io/youtube-dl/
                YTURL="`youtube-dl -g -f mp4 https://www.youtube.com/watch?v=$arg`"
            } || [ ! -z `type -p ytdl` ] && { # https://github.com/rylio/ytdl
                YTURL="`ytdl -u $arg`"
            } || {
                GET=$(wget -qO- "https://www.youtube.com/get_video_info?html5=1&video_id=$arg")
                for s in $(echo $GET | tr "&" "\n")
                do
                    if [ "${s%%=*}" = "player_response" ]; then
                        YTURL=$(echo ${s:16} | urldecode | jq '.streamingData.formats[-1].url' | tail -c +2 | head -c -2)
                    fi
                done
            }
            vol=(`log=$(echo "l($VOL)/l(10)" | bc -l); val=$(echo "$log * 2000" | bc); echo ${val%.*}`)
            omxplayer $OMX_OPTS --vol $vol "$YTURL" </dev/null &
            ;;
        play | pause)
            omxdbus org.mpris.MediaPlayer2.Player.PlayPause >/dev/null
            ;;
        stop)
            omxdbus org.mpris.MediaPlayer2.Player.Stop >/dev/null
            ;;
        seek_to)
            omxdbus org.mpris.MediaPlayer2.Player.SetPosition objpath:/not/used int64:${arg}000000 >/dev/null
            ;;
        set_volume)
            if [ $arg -lt 100 ]; then
                VOL=`echo $arg / 100 | bc -l | awk '{printf "%0.2f\n", $1}'`
            fi
            omxdbus org.freedesktop.DBus.Properties.Set string:"org.mpris.MediaPlayer2.Player" string:"Volume" double:$VOL
            ;;
    esac
done
