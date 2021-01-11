#!/bin/bash
# script dependencies raspi.sh

[ $(id -u) -ne 0 ] && {
	echo "Please run as root"
	exit 1
} || {
	[ ! -z `type -p apt` ] && {
		[ -z `type -p bc` ] && apt install bc -y
		[ -z `type -p omxplayer` ] && apt install omxplayer -y
		[ -z `type -p youtube-dl` ] && [ -z `type -p ytdl` ] && [ -z `type -p jq` ] && apt install jq -y
		[ ! -z "$ANNOTATE" ] && [ ! -z `type -p nitrogen` ] && [ -e "background.png" ] && [ -z `type -p convert` ] && apt install imagemagick -y
	}
}
