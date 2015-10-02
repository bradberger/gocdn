package gocdn

import (
	"fmt"
	"log"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"
)

// Query holds data related to a file lookup.
// This includes where it came from, it's size, contents, etc.
type Query struct {
	StausCode int
	Cached    bool
	URL       string
	Path      string
	Bytes     []byte
	Size      int
	Duration  time.Duration
}

// CDN does stuff
type CDN struct {
    Prefix string
    CacheDuration int
    Cors bool
    CacheDir string
    UseFileCache bool
}

func init() {
    log.SetPrefix("gocdn")
}

// Handler implements a standard interface for http.HandleFunc
func (c *CDN) Handler(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Cache-Control:public", fmt.Sprintf("max-age=%v", c.CacheDuration))

    if c.Cors {
        if origin := r.Header.Get("Origin"); origin != "" {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Timing-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Methods", r.Method)
            w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
        }
    }

    url := strings.Trim(strings.TrimPrefix(r.URL.Path, c.Prefix), "/")
    ext := path.Ext(url)

    parts := strings.Split(url, "/")
    if len(parts) < 2 {
        http.Error(w, http.StatusText(400), 400)
        return
    }

    f := CDNFile{
        Package:      parts[0],
        Version:      parts[1],
        Path:         strings.Join(parts[2:], "/"),
        Extension:    ext,
        Mime:         mime.TypeByExtension(ext),
        CacheDir:     c.CacheDir,
        UseFileCache: c.UseFileCache,
    }

    incoming := f.Query()
    timeout := make(chan bool, 1)
    go func() {
        time.Sleep(5 * time.Second)
        timeout <- true
    }()

    w.Header().Set("Content-Type", f.Mime)

    select {
    case res := <-incoming:
        w.Write(res.Bytes)
        return
    case <-timeout:
        http.Error(w, http.StatusText(504), 504)
        return
    }

}

func getURL(url string) <-chan Query {

	out := make(chan Query, 1)
	go func() {

		if url == "" {
			return
		}

		log.Printf("GET %s", url)

		start := time.Now()

		// @todo Keep a list of 404's/errors and don't look them up.
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != 200 {
			log.Printf("ERROR %v %s", resp.StatusCode, url)
			return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ERROR %s %s", err.Error(), url)
			return
		}

		elapsed := time.Since(start)
		out <- Query{
			Cached:   false,
			URL:      url,
			Bytes:    body,
			Size:     len(body),
			Duration: elapsed,
		}

		log.Printf("DONE %s %s", elapsed, url)

	}()

	return out

}
