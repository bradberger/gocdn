package gocdn

import (
    "io/ioutil"
    "fmt"
    "path"
    "strings"
)

// CDNFile is a struct which contains necessary information regarding a static file.
// The related methods allow for fetching and serving that file via various CDN's.
type CDNFile struct {
	Package   string
	Version   string
	Path      string
	Extension string
	Mime      string
    CacheDir  string
    UseFileCache bool
}

// CacheToDisk writes the content to disk.
func (f *CDNFile) CacheToDisk(content []byte) {
    if f.UseFileCache {
        go cacheFile(f.getCachePath(), content)
    }
}

// Query returns the file contents in []byte for each file.
// It loops through the possible file URL's, and returns the first result.
// URL's which have errors, don't exist, etc., block forever.
func (f *CDNFile) Query() <-chan Query {

	urls := f.GetUrls()
	out := make(chan Query, len(urls))

	go func() {

        if f.UseFileCache {
            contents, err := ioutil.ReadFile(path.Clean(f.getCachePath()))
            if err == nil {
                out <- Query{
                    Cached: true,
                    Path:   f.getCachePath(),
                    Bytes:  contents,
                    Size:   len(contents),
                }
                close(out)
                return
            }
        }

		haveResult := false
		for _, url := range urls {
			go func(uri string) {
				r := <-getURL(uri)
				if !haveResult {
					// First hit, cache it.
					haveResult = true
					f.CacheToDisk(r.Bytes)
				}
				out <- r
			}(url)
		}

	}()

	return out

}

func (f *CDNFile) getCdnjsURL() (url string) {
	if f.Package == "" || f.Version == "" || f.Path == "" {
		url = ""
		return
	}
	url = fmt.Sprintf("https://cdnjs.cloudflare.com/ajax/libs/%s/%s/%s", f.Package, f.Version, f.Path)
	return
}

func (f *CDNFile) getJsDelivrURL() (url string) {
	if f.Package == "" || f.Version == "" || f.Path == "" {
		url = ""
		return
	}
	url = fmt.Sprintf("https://cdn.jsdelivr.net/%s/%s/%s", f.Package, f.Version, f.Path)
	return
}

func (f *CDNFile) getGoogleApisURL() (url string) {
	if f.Package == "" || f.Version == "" || f.Path == "" {
		url = ""
		return
	}
	url = fmt.Sprintf("https://ajax.googleapis.com/ajax/libs/%s/%s/%s", f.Package, f.Version, f.Path)
	return
}

func (f *CDNFile) getAspNetURL() (url string) {
	if f.Package == "" || f.Version == "" || f.Path == "" {
		url = ""
		return
	}
	path := strings.TrimRight(f.Path, f.Extension)
	min := strings.HasSuffix(path, ".min")
	minStr := ""
	if min {
		path = strings.TrimRight(path, ".min")
		minStr = ".min"
	}

	url = fmt.Sprintf("https://ajax.aspnetcdn.com/ajax/%s/%s-%s%s%s", f.Package, path, f.Version, minStr, f.Extension)
	return
}

func (f *CDNFile) getYandexURL() (url string) {
	if f.Package == "" || f.Version == "" || f.Path == "" {
		url = ""
		return
	}
	url = fmt.Sprintf("https://yastatic.net/%s/%s/%s", f.Package, f.Version, f.Path)
	return
}

func (f *CDNFile) getOssCdnURL() (url string) {
	if f.Package == "" || f.Version == "" || f.Path == "" {
		url = ""
		return
	}
	url = fmt.Sprintf("https://oss.maxcdn.com/%s/%s/%s", f.Package, f.Version, f.Path)
	return
}

// GetUrls returns a set of possible locations for each file based on available CDNs.
func (f *CDNFile) GetUrls() []string {
	return []string{
		f.getCdnjsURL(),
		f.getJsDelivrURL(),
		f.getGoogleApisURL(),
		f.getOssCdnURL(),
		f.getYandexURL(),
		f.getAspNetURL(),
	}
}

func (f *CDNFile) getCachePath() string {
	return path.Clean(fmt.Sprintf("%s/%s/%s/%s", f.CacheDir, f.Package, f.Version, f.Path))
}
