package origin

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cbsinteractive/bakery/pkg/config"
)

//Origin interface is implemented on Manifest and Propeller struct
type Origin interface {
	GetPlaybackURL() string
	FetchManifest(c config.Config) (string, error)
}

//Manifest struct holds Origin and Path of Manifest
//Variant level manifests will be base64 encoded absolute path
type Manifest struct {
	Origin   string
	Path     string
	Absolute bool
}

//Configure will return proper Origin interface
func Configure(c config.Config, path string) (Origin, error) {
	var renditionURL string

	if strings.Contains(path, "propeller") {
		parts := strings.Split(path, "/") //["", "propeller", "orgID", "channelID.m3u8"]
		if len(parts) != 4 {
			if len(parts) != 5 {
				return &Propeller{}, fmt.Errorf("url path does not follow `/propeller/orgID/channelID.m3u8`")
			}
			renditionURL = parts[4] //base64.m3u8 is rendition level manifest
		}

		orgID := parts[2]
		channelID := strings.Split(parts[3], ".")[0] // split off .m3u8

		o, err := NewPropeller(c, orgID, channelID, renditionURL)
		if err != nil {
			return &Propeller{}, fmt.Errorf("configuring propeller origin: %w", err)
		}

		return o, nil
	}

	//check if rendition URL
	parts := strings.Split(path, "/")
	if len(parts) == 2 { //["", "base64.m3u8"]
		renditionURL, err := decodeRenditionURL(parts[1])
		if err != nil {
			return &Manifest{}, fmt.Errorf("configuring rendition url: %w", err)
		}
		path = renditionURL
	}

	return NewManifest(c, path), nil
}

//NewManifest returns a new Origin struct
func NewManifest(c config.Config, path string) *Manifest {
	var absolute bool
	if strings.Contains(path, "http") {
		absolute = true
	}

	return &Manifest{
		Origin:   c.OriginHost,
		Path:     path,
		Absolute: absolute,
	}
}

//GetPlaybackURL will retrieve url
func (m *Manifest) GetPlaybackURL() string {
	if m.Absolute {
		return m.Path
	}

	return m.Origin + m.Path
}

//GetPath will return Path to manifest
func (m *Manifest) GetPath() string {
	path := strings.Split(m.Path, ".")[0] + "/"
	return path
}

//FetchManifest will grab manifest contents of configured origin
func (m *Manifest) FetchManifest(c config.Config) (string, error) {
	return fetch(c, m.GetPlaybackURL())
}

func fetch(c config.Config, manifestURL string) (string, error) {
	resp, err := c.Client.New().Get(manifestURL)
	if err != nil {
		return "", fmt.Errorf("fetching manifest: %w", err)
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading manifest response body: %w", err)
	}

	if sc := resp.StatusCode; sc/100 > 3 {
		return "", fmt.Errorf("fetching manifest: returning http status of %v", sc)
	}

	return string(contents), nil
}
