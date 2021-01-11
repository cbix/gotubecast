ID="Dgxz0kZ2dp4"
TIMEFORMAT='%3R'

[ ! -e "/tmp/gotubecast" ] && {
	mkdir "/tmp/gotubecast"
	cd "/tmp/gotubecast"
	[ ! -e "./youtube-dl" ] && {
		wget "https://yt-dl.org/downloads/latest/youtube-dl" -O ./youtube-dl
		chmod +x ./youtube-dl
	}
	[ ! -e "./youtubedr" ] && {
		wget "https://github.com/kkdai/youtube/releases/download/v2.3.0/youtube_2.3.0_linux_armv6.tar.gz" -O ./youtubedr.tar.gz
		tar xf ./youtubedr.tar.gz
	}
} || cd "/tmp/gotubecast"

time { # rpi1: 3.956
	function urldecode {
		echo -e "$(sed 's/+/ /g;s/%\(..\)/\\x\1/g;')"
	}
	GET=$(wget -qO- "https://www.youtube.com/get_video_info?html5=1&video_id=$ID")
	for s in $(echo $GET | tr "&" "\n")
	do
		if [ "${s%%=*}" = "player_response" ]; then
			YTURL=$(echo ${s:16} | urldecode | jq '.streamingData.formats[-1].url' | tail -c +2 | head -c -2)
		fi
	done
	echo $YTURL;
}
time { # rpi1: 5.069
	YTURL=$(./youtubedr url -q hd1080 "https://www.youtube.com/watch?v=$ID")
	echo $YTURL;
}
time { # rpi1: 67.614
	YTURL=$(./youtube-dl -g -f mp4 "https://www.youtube.com/watch?v=$ID")
	echo $YTURL;
}
