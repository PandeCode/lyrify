package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	redirectURI = "http://localhost:8888/callback"
	sleepSecs   = 1

	addr = ":8888"
)

var (
	cacheDir string

	line          string = "Loading..."
	currentURL    string = ""
	currentLyrics Lyrics
)

type Lyrics []struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  *string `json:"plainLyrics"`
	SyncedLyrics *string `json:"syncedLyrics"`
}

func main() {
	cacheDir = filepath.Join(os.Getenv("HOME"), ".cache/lyrify")
	startServer()
}

func computePlainLines(s string, progress, duration int) {
	lines := strings.Split(s, "\n")
	n := len(lines)

	line = lines[int(progress*n/duration)]
}

func computeSyncedLines(s string, progress int) {
	lines := strings.Split(s, "\n")
	len_lines := len(lines)

	var index int
	for i := 0; i < len_lines; i++ {
		ms := strToMs(lines[i][:10])
		if ms < progress {
			index = i
		}
	}

	line = lines[index][10:]
}

func strToMs(s string) int {
	_min_, err := strconv.Atoi(s[1:3])
	hperr(err)
	secs, err := strconv.Atoi(s[4:6])
	hperr(err)
	ms, err := strconv.Atoi(s[7:9])
	hperr(err)

	return int(1000*(_min_*60+secs) + ms)
}

func startServer() {

	http.HandleFunc("/line", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(line))
	})

	go func() {
		err := http.ListenAndServe(addr, nil)
		hperr(err)
	}()

	fmt.Print("Starting server on 8888...")
	for {
		time.Sleep(sleepSecs * time.Second)

		var name, artists string
		var progress, duration int
		var ok bool

		name, artists, progress, duration, ok = getMprisInfo()

		if !ok {
			continue
		}

		lyrics := getLyrics(name, artists)

		synced := lyrics[0].SyncedLyrics
		plain := lyrics[0].PlainLyrics

		if synced != nil {
			computeSyncedLines(*synced, progress)
		} else if plain != nil {
			computePlainLines(*plain, progress, duration)
		} else {
			line = "🎼"
		}
	}
}

func hperr(err any) {
	if err != nil {
		log.Fatalln(err)
		panic("Forced Panic on Error")
	}
}

func herr(err any) {
	if err != nil {
		log.Fatalln(err)
	}
}

func hashStr(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprint(h.Sum32())
}

func getLyrics(name string, artists string) Lyrics {
	params := url.Values{}
	params.Add("q", name+" "+artists)
	url := "https://lrclib.net/api/search?" + params.Encode()

	if currentURL == url {
		return currentLyrics
	}
	currentURL = url

	file := filepath.Join(cacheDir, hashStr(url))

	var body []byte
	b, err := os.ReadFile(file)
	if err != nil {
		resp, err := http.Get(url)
		herr(err)
		body, err = io.ReadAll(resp.Body)
		herr(err)

		os.WriteFile(file, body, 0644)
	} else {
		body = b
	}

	json.Unmarshal(body, &currentLyrics)
	return currentLyrics
}

func getMprisInfo() (string, string, int, int, bool) {
	res, err := exec.Command("playerctl", "-p", "spotify", "metadata", "--format", "{{title}}\n{{artist}}\n{{position}}\n{{mpris:length}}").Output()

	_data := string(res)
	if err != nil || _data == "No players found" {
		return "", "", 0, 0, false
	}
	data := strings.Split(_data, "\n")

	name := data[0]
	artists := data[1]
	progress, err := strconv.Atoi(data[2])
	duration, err := strconv.Atoi(data[3])

	return name, artists, int(float32(progress) * 0.001), int(float32(duration) * 0.001), true
}
