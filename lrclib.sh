#!/usr/bin/env bash

source ./lib.sh

url="https://lrclib.net/api/search?q=$(get_title_artists)"

res=$(fetch_with_cache "$url")

synced=$(echo "$res" | jq ".[0].syncedLyrics")

printf "$synced"
