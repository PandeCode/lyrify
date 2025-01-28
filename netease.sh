#!/usr/bin/env bash

source ./lib.sh

keywords=$(get_title_artists)
song_id=$(fetch_with_cache "https://music.xianqiao.wang/neteaseapiv2/search?limit=10&type=1&keywords=$keywords" | jq ".result.songs[0].id")

lyrics=$(fetch_with_cache "https://music.xianqiao.wang/neteaseapiv2/lyric?id=$song_id" | jq ".lrc.lyric")

printf "$lyrics"
