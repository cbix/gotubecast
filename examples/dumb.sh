#!/bin/bash
# This TV player opens every started video in a new browser window.
# It doesn't handle any controls (play/pause/...).
# Use this script as a starting point for your own bash-based YouTube TV project!
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
