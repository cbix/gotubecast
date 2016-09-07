#!/bin/bash
# simple YouTube TV for Raspberry Pi
# kills any currently playing video as soon as a new video is queued
# needs youtube-dl and omxplayer
export SCREEN_ID=""
export SCREEN_NAME="Raspberry Pi"
export SCREEN_APP="pitubecast-v1"
export OMX_OPTS="-o hdmi"
export YTDL_OPTS="-f mp4"
./gotubecast -s "$SCREEN_ID" -n "$SCREEN_NAME" -i "$SCREEN_APP" | while read line
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
			YTURL="`youtube-dl -g $YTDL_OPTS https://youtube.com?v=$arg`"
			killall omxplayer.bin
			omxplayer $OMX_OPTS "$YTURL" &
			;;
	esac
done
