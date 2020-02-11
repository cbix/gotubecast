#!/bin/bash
# simple YouTube TV for Raspberry Pi
# kills any currently playing video as soon as a new video is queued
# needs youtube-dl (or ytdl) and omxplayer
export SCREEN_ID=""
export SCREEN_NAME="Raspberry Pi"
export SCREEN_APP="pitubecast-v1"
export OMX_OPTS="-o hdmi"
# http://rg3.github.io/youtube-dl/
export YOUTUBEDL="youtube-dl -g -f mp4 https://www.youtube.com/watch?v="
# https://github.com/rylio/ytdl
export YTDL="ytdl -u "
export EXTRACTOR="$YOUTUBEDL"
export VOL="1.0"

function omxdbus {
    OMXPLAYER_DBUS_ADDR="/tmp/omxplayerdbus.${USER:-root}"
    OMXPLAYER_DBUS_PID="/tmp/omxplayerdbus.${USER:-root}.pid"
    export DBUS_SESSION_BUS_ADDRESS=`cat $OMXPLAYER_DBUS_ADDR`
    export DBUS_SESSION_BUS_PID=`cat $OMXPLAYER_DBUS_PID`
    dbus-send --print-reply=literal --session --reply-timeout=100 --dest=org.mpris.MediaPlayer2.omxplayer /org/mpris/MediaPlayer2 $*
}

gotubecast -s "$SCREEN_ID" -n "$SCREEN_NAME" -i "$SCREEN_APP" | while read line
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
            YTURL="`$EXTRACTOR$arg`"
            killall omxplayer.bin
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
