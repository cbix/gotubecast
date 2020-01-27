#!/bin/bash
# simple YouTube TV for Raspberry Pi
# kills any currently playing video as soon as a new video is queued
# ONLY NEEDS cvlc no need for youtube-dl
# Made this script out of the already existing raspi.sh 
# by bearkillerPT :)
# 
# Everything working but the Seek functionality... Specs: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html
# I'll keep trying to implement it but keep in mind that the Property in Mediaplayer2.Player.Postion is read only...
# Omxplayer wassn't working on my rasp and so this way you can also drop the youtube-dl dependecy

export SCREEN_ID=""
export SCREEN_NAME="Raspberry Pi"
export SCREEN_APP="pitubecast-v1"
export OMX_OPTS="-o hdmi"
export VOL="1.0"


function omxdbus {  	
   dbus-send --type=method_call --dest=org.mpris.MediaPlayer2.vlc /org/mpris/MediaPlayer2 $*
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
            echo "you/$arg"
	    sleep 5
	    killall -9 vlc 
            cvlc "http://youtu.be/$arg" </dev/null &
	    ;;
        play | pause)
            omxdbus org.mpris.MediaPlayer2.Player.PlayPause >/dev/null
            ;;
        stop)
            omxdbus org.mpris.MediaPlayer2.Player.Stop >/dev/null
            ;;
        seek_to)
	    pos= omxdbus org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player string:Position >/dev/null
	    offset=$(($arg-$pos))
	    offset=$(($offset*1000000))
	    echo "$offset"
	    omxdbus org.mpris.MediaPlayer2.Player.Seek int64:$offset >/dev/null

            ;;
        set_volume)
	    VOL= omxdbus org.mpris.MediaPlayer2.Player.Volume >/dev/null
            #echo $VOL
	    if [ $arg -lt 100 ]; then
                VOL=`echo $arg / 100 | bc -l | awk '{printf "%0.2f\n", $1}'`
            fi
	    echo "try: " $VOL
	    omxdbus org.freedesktop.DBus.Properties.Set string:org.mpris.MediaPlayer2.Player string:Volume variant:double:$VOL >/dev/null
	    ;;
    esac
done
