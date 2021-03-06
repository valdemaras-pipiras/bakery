package parsers

import (
	"fmt"
	"math"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// VideoType is the video codec we need in a given playlist
type VideoType string

// AudioType is the audio codec we need in a given playlist
type AudioType string

// AudioLanguage is the audio language we need in a given playlist
type AudioLanguage string

// CaptionLanguage is the audio language we need in a given playlist
type CaptionLanguage string

// CaptionType is an allowed caption format for the stream
type CaptionType string

// StreamType represents one stream type (e.g. video, audio, text)
type StreamType string

// Protocol describe the valid protocols
type Protocol string

const (
	videoHDR10       VideoType = "hdr10"
	videoDolbyVision VideoType = "dovi"
	videoHEVC        VideoType = "hevc"
	videoH264        VideoType = "avc"

	audioAAC                AudioType = "aac"
	audioAC3                AudioType = "ac-3"
	audioEnhacedAC3         AudioType = "ec-3"
	audioNoAudioDescription AudioType = "noAd"

	audioLangPTBR AudioLanguage = "pt-BR"
	audioLangES   AudioLanguage = "es-MX"
	audioLangEN   AudioLanguage = "en"

	captionPTBR CaptionLanguage = "pt-BR"
	captionES   CaptionLanguage = "es-MX"
	captionEN   CaptionLanguage = "en"

	// ProtocolHLS for manifest in hls
	ProtocolHLS Protocol = "hls"
	// ProtocolDASH for manifests in dash
	ProtocolDASH Protocol = "dash"
)

// Trim is a struct that carries the start and end times to trim playlist
type Trim struct {
	Start int64 `json:",omitempty"`
	End   int64 `json:",omitempty"`
}

// MediaFilters is a struct that carry all the information passed via url
type MediaFilters struct {
	Videos            []VideoType       `json:",omitempty"`
	Audios            []AudioType       `json:",omitempty"`
	AudioLanguages    []AudioLanguage   `json:",omitempty"`
	CaptionLanguages  []CaptionLanguage `json:",omitempty"`
	CaptionTypes      []CaptionType     `json:",omitempty"`
	FilterStreamTypes []StreamType      `json:",omitempty"`
	MaxBitrate        int               `json:",omitempty"`
	MinBitrate        int               `json:",omitempty"`
	Plugins           []string          `json:",omitempty"`
	Trim              *Trim             `json:",omitempty"`
	Protocol          Protocol          `json:"protocol"`
}

var urlParseRegexp = regexp.MustCompile(`(.*?)\((.*)\)`)

// URLParse will generate a MediaFilters struct with
// all the filters that needs to be applied to the
// master manifest. It will also return the master manifest
// url without the filters.
func URLParse(urlpath string) (string, *MediaFilters, error) {
	mf := new(MediaFilters)
	parts := strings.Split(urlpath, "/")
	re := urlParseRegexp
	masterManifestPath := "/"

	if strings.Contains(urlpath, ".m3u8") {
		mf.Protocol = ProtocolHLS
	} else if strings.Contains(urlpath, ".mpd") {
		mf.Protocol = ProtocolDASH
	}

	// set bitrate defaults
	mf.MinBitrate = 0
	mf.MaxBitrate = math.MaxInt32

	for _, part := range parts {
		// FindStringSubmatch should return a slice with
		// the full string, the key and filters (3 elements).
		// If it doesn't match, it means that the path is part
		// of the official manifest path so we concatenate to it.
		subparts := re.FindStringSubmatch(part)
		if len(subparts) != 3 {
			if mf.filterPlugins(part) {
				continue
			}
			masterManifestPath = path.Join(masterManifestPath, part)
			continue
		}

		filters := strings.Split(subparts[2], ",")

		var err error
		switch key := subparts[1]; key {
		case "v":
			for _, videoType := range filters {
				if videoType == "hdr10" {
					mf.Videos = append(mf.Videos, VideoType("hev1.2"), VideoType("hvc1.2"))
					continue
				}

				mf.Videos = append(mf.Videos, VideoType(videoType))
			}
		case "a":
			for _, audioType := range filters {
				mf.Audios = append(mf.Audios, AudioType(audioType))
			}
		case "al":
			for _, audioLanguage := range filters {
				mf.AudioLanguages = append(mf.AudioLanguages, AudioLanguage(audioLanguage))
			}
		case "c":
			for _, captionLanguage := range filters {
				mf.CaptionLanguages = append(mf.CaptionLanguages, CaptionLanguage(captionLanguage))
			}
		case "ct":
			if mf.CaptionTypes == nil {
				mf.CaptionTypes = []CaptionType{}
			}

			for _, captionType := range filters {
				mf.CaptionTypes = append(mf.CaptionTypes, CaptionType(captionType))
			}
		case "fs":
			for _, streamType := range filters {
				mf.FilterStreamTypes = append(mf.FilterStreamTypes, StreamType(streamType))
			}
		case "b":
			if filters[0] != "" {
				mf.MinBitrate, err = strconv.Atoi(filters[0])
				if err != nil {
					return keyError("trim", err)
				}
			}

			if filters[1] != "" {
				mf.MaxBitrate, err = strconv.Atoi(filters[1])
				if err != nil {
					return keyError("trim", err)
				}
			}

			if isGreater(mf.MinBitrate, mf.MaxBitrate) {
				return keyError("bitrate", fmt.Errorf("Min Bitrate is greater than or equal to Max Bitrate"))
			}
		case "t":
			var trim Trim
			if filters[0] != "" {
				trim.Start, err = strconv.ParseInt(filters[0], 10, 64)
				if err != nil {
					return keyError("trim", err)
				}
			}

			if filters[1] != "" {
				trim.End, err = strconv.ParseInt(filters[1], 10, 64)
				if err != nil {
					return keyError("trim", err)
				}
			}

			if isGreater(int(trim.Start), int(trim.End)) {
				return keyError("trim", fmt.Errorf("Start Time is greater than or equal to End Time"))
			}

			mf.Trim = &trim
		}
	}

	return masterManifestPath, mf, nil
}

// validate ranges like Trim and Bitrate
func isGreater(x int, y int) bool {
	return x >= y
}

func keyError(key string, e error) (string, *MediaFilters, error) {
	return "", &MediaFilters{}, fmt.Errorf("Error parsing filter key: %v. Got error: %w", key, e)
}

func (f *MediaFilters) filterPlugins(path string) bool {
	re := regexp.MustCompile(`\[(.*)\]`)
	subparts := re.FindStringSubmatch(path)

	if len(subparts) == 2 {
		for _, plugin := range strings.Split(subparts[1], ",") {
			f.Plugins = append(f.Plugins, plugin)
		}
		return true
	}

	return false
}

//DefinesBitrateFilter will check if bitrate filter is set
func (f *MediaFilters) DefinesBitrateFilter() bool {
	return (f.MinBitrate >= 0 && f.MaxBitrate <= math.MaxInt32) &&
		(f.MinBitrate < f.MaxBitrate) &&
		!(f.MinBitrate == 0 && f.MaxBitrate == math.MaxInt32)
}
