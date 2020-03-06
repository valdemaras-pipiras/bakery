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

type Codec string

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

	codecHDR10              Codec = "hdr10"
	codecDolbyVision        Codec = "dovi"
	codecHEVC               Codec = "hevc"
	codecH264               Codec = "avc"
	codecAAC                Codec = "aac"
	codecAC3                Codec = "ac-3"
	codecEnhancedAC3        Codec = "ec-3"
	codecNoAudioDescription Codec = "noAd"

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
<<<<<<< HEAD
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
=======
	Videos             []VideoType       `json:",omitempty"`
	VideoSubFilters    Subfilters        `json:",omitempty"`
	Audios             []AudioType       `json:",omitempty"`
	AudioSubFilters    Subfilters        `json:",omitempty"`
	AudioLanguages     []AudioLanguage   `json:",omitempty"`
	CaptionLanguages   []CaptionLanguage `json:",omitempty"`
	CaptionTypes       []CaptionType     `json:",omitempty"`
	FilterStreamTypes  []StreamType      `json:",omitempty"`
	FilterBitrateTypes []StreamType      `json:",omitempty"`
	MaxBitrate         int               `json:",omitempty"`
	MinBitrate         int               `json:",omitempty"`
	Protocol           Protocol          `json:"protocol"`
>>>>>>> implement basic method of applying bitrate filter to audio/video types
}

<<<<<<< HEAD
=======
type Subfilters struct {
	MinBitrate int     `json:",omitempty"`
	MaxBitrate int     `json:",omitempty"`
	Codecs     []Codec `json:",omitempty"`
}

>>>>>>> very basic nested audio/video bitrate filter
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
	//mf.MinBitrate = 0
	//mf.MaxBitrate = math.MaxInt32
	mf.initializeBitrateRange()

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
<<<<<<< HEAD
<<<<<<< HEAD

		var err error
=======
		fmt.Println(filters)
>>>>>>> implement basic method of applying bitrate filter to audio/video types
=======

<<<<<<< HEAD
>>>>>>> add some tests
=======
		tesRX := regexp.MustCompile(`\),`)
		subfilters := SplitAfter(subparts[2], tesRX)

>>>>>>> very basic nested audio/video bitrate filter
		switch key := subparts[1]; key {
		case "v":
			for _, sf := range subfilters {
				splitSubfilter := re.FindStringSubmatch(sf)
				var key string
				var param []string
				if len(splitSubfilter) == 0 {
					key = "c"
					param = strings.Split(sf, ",")
				} else {
					key = splitSubfilter[1]
					param = strings.Split(splitSubfilter[2], ",")
				}
				mf.normalizeSubfilter(StreamType("video"), key, param)
			}
			for _, videoType := range filters {
				if videoType == "hdr10" {
					mf.Videos = append(mf.Videos, VideoType("hev1.2"), VideoType("hvc1.2"))
					continue
				}

				mf.Videos = append(mf.Videos, VideoType(videoType))
			}
		case "a":
			for _, sf := range subfilters {
				splitSubfilter := re.FindStringSubmatch(sf) // right now, assuming the nested filters are of valid form, maybe add a check here for that
				var key string
				var param []string
				if len(splitSubfilter) == 0 {
					key = "c"
					param = strings.Split(sf, ",")
				} else {
					key = splitSubfilter[1]
					param = strings.Split(splitSubfilter[2], ",")
				}
				mf.normalizeSubfilter(StreamType("audio"), key, param)
			}
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
<<<<<<< HEAD
<<<<<<< HEAD
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
=======
			for i := 0; i < len(filters)-2; i++ {
				mf.FilterBitrateTypes = append(mf.FilterBitrateTypes, StreamType(filters[i]))
			}
			if filters[len(filters)-2] != "" {
				mf.MinBitrate, _ = strconv.Atoi(filters[len(filters)-2])
			}

			if filters[len(filters)-1] != "" {
				mf.MaxBitrate, _ = strconv.Atoi(filters[len(filters)-1])
>>>>>>> implement basic method of applying bitrate filter to audio/video types
			}

			if isGreater(int(trim.Start), int(trim.End)) {
				return keyError("trim", fmt.Errorf("Start Time is greater than or equal to End Time"))
=======
			if filters[0] != "" {
				mf.MinBitrate, _ = strconv.Atoi(filters[0])
			}

			if filters[1] != "" {
				mf.MaxBitrate, _ = strconv.Atoi(filters[1])
>>>>>>> very basic nested audio/video bitrate filter
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

func (f *Subfilters) BitrateSubfilterApplies() bool {
	return f.MinBitrate != 0 || f.MaxBitrate != math.MaxInt32
}

func (f *MediaFilters) BitrateFilterApplies() bool {
	overall := f.DefinesBitrateFilter()
	audio := f.AudioSubFilters.BitrateSubfilterApplies()
	video := f.VideoSubFilters.BitrateSubfilterApplies()
	return overall || audio || video
}

func (f *MediaFilters) initializeBitrateRange() {
	f.MinBitrate = 0
	f.MaxBitrate = math.MaxInt32
	f.AudioSubFilters.MinBitrate = 0
	f.AudioSubFilters.MaxBitrate = math.MaxInt32
	f.VideoSubFilters.MinBitrate = 0
	f.VideoSubFilters.MaxBitrate = math.MaxInt32
}

func SplitAfter(s string, re *regexp.Regexp) []string {
	var splitResults []string
	var position int
	indices := re.FindAllStringIndex(s, -1)
	if indices == nil {
		return append(splitResults, s)
	}
	for _, idx := range indices {
		section := s[position:idx[1]]
		splitResults = append(splitResults, section)
		position = idx[1]
	}
	return append(splitResults, s[position:])
}

func (f *MediaFilters) normalizeSubfilter(streamType StreamType, key string, values []string) {
	var streamSubfilters *Subfilters
	switch streamType {
	case "audio":
		streamSubfilters = &f.AudioSubFilters
	case "video":
		streamSubfilters = &f.VideoSubFilters
	}
	fmt.Printf("key: %v value: %v\n", key, values)
	switch key {
	case "c":
		for _, v := range values {
			streamSubfilters.Codecs = append(streamSubfilters.Codecs, Codec(v))
		}
	case "b":
		if values[0] != "" {
			streamSubfilters.MinBitrate, _ = strconv.Atoi(values[0])
		}

		if values[1] != "" {
			streamSubfilters.MaxBitrate, _ = strconv.Atoi(values[1])
		}
	}
}
