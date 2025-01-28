# **Lyrify**

Lyrify is a lightweight server that fetches and serves real-time song lyrics for tracks playing on Spotify using `playerctl`.

Additional scripts to fetch lyrics.

## **How It works**

- `playerctl` to get current song info.
- Fetches song lyrics (plain or synchronized) from the [Lrclib](https://lrclib.net/) API.
- Displays the current line of lyrics via a local HTTP endpoint.
- Caches API results to reduce redundant requests (~/.cache/lyrify).

## **Requirements**

- **Go** (1.17+)
- **playerctl** (configured to work with Spotify)
- Spotify desktop client running

## **Installation**

1. Clone this repository:

   ```bash
   git clone https://github.com/yourusername/lyrify
   cd lyrify
   ```

2. Build the binary:

   ```bash
   go build -o lyrify
   ```

3. Run the program:
   ```bash
   ./lyrify
   ```

## **Usage**

1. Start Lyrify:

   ```bash
   ./lyrify
   ```

   The server starts on port `8888` by default.

2. Access the current line of lyrics via:

   ```bash
   curl -s http://localhost:8888/line
   ```

3. Ensure Spotify is playing a track. The `/line` endpoint will return:
   - **Plain lyrics**: Current line based on playback duration.
   - **Synchronized lyrics**: Accurate line based on timestamps.
   - `🎼` if no lyrics are found or the track is instrumental or podcast.

## **Configuration**

- The server listens on port `8888` by default. You can change this by setting the `LYRIFY_ADDR` environment variable:
  ```bash
  export LYRIFY_ADDR=":8080"
  ./lyrify
  ```
