# gotubecast
gotubecast is a small Go program which you can use to make your own YouTube TV player.

It connects to the YouTube Leanback API and generates a text stream providing pairing codes, video IDs,
play/pause/seek/volume change commands etc. It doesn't have any dependencies and runs on any of the platforms supported by golang.
For example, use it on a Raspberry Pi in combination with youtube-dl and omxplayer for a DIY Chromecast clone or make a YouTube TV
extension for your favorite media center software.

Work in progress, still contains many hardcoded values. The essential commands are not implemented yet,
but you can already connect your phone and watch the generic command stream.
