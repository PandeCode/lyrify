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

	"sync"

	"github.com/godbus/dbus/v5"
)

const (
	sleepTime = 1 * time.Second
)

var (
	cacheDir string

	loadingText = "⸜(｡˃ ᵕ ˂ )⸝♡"

	addr       string = ":8888"
	line       string = loadingText
	currentURL string = ""

	prevSong      string
	currentLyrics Lyrics

	connOnce sync.Once
	conn     *dbus.Conn
	connErr  error
	obj      dbus.BusObject
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
	_, err := exec.LookPath("playerctl")
	if err != nil {
		log.Fatalf("playerctl is not installed: %v", err)
	}

	_addr := os.Getenv("LYRIFY_ADDR")
	if _addr != "" {
		addr = _addr
	}

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

	fmt.Printf("Starting server on localhost%s ...", addr)
	for {
		time.Sleep(sleepTime)

		var name, artists string
		var progress, duration int
		var ok bool

		name, artists, progress, duration, ok = getMprisInfo()

		if !ok {
			continue
		}

		n_a := (name + artists)
		if n_a != prevSong {
			line = loadingText
		} else {
			prevSong = n_a
		}

		lyrics := getLyrics(n_a)

		if len(lyrics) == 0 {
			line = "🎼"
			continue
		}

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

func getLyrics(song string) Lyrics {
	params := url.Values{}
	params.Add("q", song)
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

		go os.WriteFile(file, body, 0644)
	} else {
		body = b
	}

	json.Unmarshal(body, &currentLyrics)
	return currentLyrics
}

func getMprisInfo() (string, string, int, int, bool) {
	connOnce.Do(func() {
		conn, connErr = dbus.SessionBus()
	})
	if connErr != nil {
		return "", "", 0, 0, false
	}

	if obj == nil {
		obj = conn.Object("org.mpris.MediaPlayer2.spotify", "/org/mpris/MediaPlayer2")
	}

	var playbackStatus string
	err := obj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.mpris.MediaPlayer2.Player", "PlaybackStatus").Store(&playbackStatus)
	if err != nil {
		return "", "", 0, 0, false
	}

	var metadata map[string]dbus.Variant
	err = obj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.mpris.MediaPlayer2.Player", "Metadata").Store(&metadata)
	if err != nil {
		return "", "", 0, 0, false
	}

	var title, artists string
	if titleVar, ok := metadata["xesam:title"]; ok {
		title, _ = titleVar.Value().(string)
	}
	if artistVar, ok := metadata["xesam:artist"]; ok {
		if artistsSlice, ok := artistVar.Value().([]string); ok {
			artists = strings.Join(artistsSlice, ", ")
		}
	}

	var length int64
	if lengthVar, ok := metadata["mpris:length"]; ok {
		length, _ = lengthVar.Value().(int64)
	}

	var position int64
	err = obj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.mpris.MediaPlayer2.Player", "Position").Store(&position)
	if err != nil {
		return "", "", 0, 0, false
	}

	progress := int(float32(position) * 0.001)
	duration := int(float32(length) * 0.001)

	return title, artists, progress, duration, true
}
