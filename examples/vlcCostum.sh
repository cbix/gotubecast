#!/bin/bash
# simple YouTube TV for Raspberry Pi
# kills any currently playing video as soon as a new video is queued
# ONLY NEEDS cvlc no need for youtube-dl
# Made this script out of the already existing raspi.sh 
# by bearkillerPT :)
# 
# Everything working kinda fine now... Specs: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html
# Finally managed to get it working though I'm having some issues with vlc not keeping up? Might be a connection issue:
#	Log:
#		[61cc0ad8] main decoder error: buffer deadlock prevented
#		[00a50b70] main audio output warning: buffer too late (-27547070 us): dropped
#		[00a50b70] pulse audio output warning: starting late (-30195 us)
#		[00a50b70] main audio output warning: playback too late (125949): up-sampling
# Omxplayer wassn't working on my rasp and so this way you can also drop the youtube-dl dependecy

export SCREEN_ID=""
export SCREEN_NAME="Raspberry Pi"
export SCREEN_APP="pitubecast-v1"
export OMX_OPTS="-o hdmi"
#export POS="1"


function omxdbus {  	
		dbus-send --type=method_call --reply-timeout=12000 --print-reply --dest=org.mpris.MediaPlayer2.vlc /org/mpris/MediaPlayer2 $* 
}

function mult1000 () {
		local floor=${1%.*}
      	[[ $floor = "0" ]] && floor=''
		local frac='0000'
        [[ $floor != $1 ]] && frac=${1#*.}$frac
		POS=$(( ${floor}${frac:0:3} ))
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
			killall -9 vlc 
			cvlc -v "http://youtu.be/$arg" </dev/null &
			;;
        play | pause)
            omxdbus org.mpris.MediaPlayer2.Player.PlayPause >/dev/null
            ;;
        stop)
            omxdbus org.mpris.MediaPlayer2.Player.Stop >/dev/null
            ;;
        seek_to)
			echo "Set Position: "$arg
			mult1000 $arg
			(( POS=$POS*1000 ))
	    	STARTAGAIN64BIT=$(( -9223372036854775808 ))
			omxdbus org.mpris.MediaPlayer2.Player.Seek int64:$STARTAGAIN64BIT >/dev/null 
			omxdbus org.mpris.MediaPlayer2.Player.Seek int64:$POS >/dev/null
			#Here had to abuse seek function....
			;;
        set_volume)
			if [ $arg -lt 100 ]; then
                VOL=`echo $arg / 100 | bc -l | awk '{printf "%0.2f\n", $1}'`
            fi
			echo "Set Volume: " $VOL
			omxdbus org.freedesktop.DBus.Properties.Set string:org.mpris.MediaPlayer2.Player string:Volume variant:double:$VOL >/dev/null
			;;
    esac
done
