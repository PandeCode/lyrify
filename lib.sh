#!/usr/bin/env bash

cache_dir=$HOME/.cache/spotine
mkdir -p "$cache_dir"

function hash() {
	echo -n "$1" | md5sum | awk '{print $1}'
}

function fetch_with_cache() {
	uri="$1"
	file_path="$cache_dir/$(hash "$uri")"

	if [ -f "$file_path" ]; then
		res=$(cat "$file_path")
		echo "$res"
	else
		res=$(curl -s "$1")
		echo "$res" >"$file_path"
		echo "$res"
	fi
}

function get_title_artists() {
	playerctl -p spotify metadata --format "{{title}} {{artist}}" | jq -sRr @uri
}
