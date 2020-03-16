package filters

import (
	"math"
	"strings"

	"github.com/cbsinteractive/bakery/pkg/parsers"
)

// Filter is an interface for HLS and DASH filters
type Filter interface {
	FilterManifest(filters *parsers.MediaFilters) (string, error)
}

// ContentType represents the content in the stream
type ContentType string

const (
	captionContentType ContentType = "text"
	audioContentType   ContentType = "audio"
	videoContentType   ContentType = "video"
)

// CodecFilterID is the formatted codec represented in a given playlist
type CodecFilterID string

const (
	hevcCodec  CodecFilterID = "hvc"
	avcCodec   CodecFilterID = "avc"
	dolbyCodec CodecFilterID = "dvh"
	aacCodec   CodecFilterID = "mp4a"
	ec3Codec   CodecFilterID = "ec-3"
	ac3Codec   CodecFilterID = "ac-3"
	stppCodec  CodecFilterID = "stpp"
	wvttCodec  CodecFilterID = "wvtt"
)

// ValidCodecs returns a map of all formatted values for a given codec filter
func ValidCodecs(codec string, filter CodecFilterID) bool {
	return strings.Contains(codec, string(filter))
}

// Returns true if given codec is an audio codec (mp4a, ec-3, or ac-3)
func isAudioCodec(codec string) bool {
	return (ValidCodecs(codec, aacCodec) ||
		ValidCodecs(codec, ec3Codec) ||
		ValidCodecs(codec, ac3Codec))
}

// Returns true if given codec is a video codec (hvc, avc, or dvh)
func isVideoCodec(codec string) bool {
	return (ValidCodecs(codec, hevcCodec) ||
		ValidCodecs(codec, avcCodec) ||
		ValidCodecs(codec, dolbyCodec))
}

// Returns true if goven codec is a caption codec (stpp or wvtt)
func isCaptionCodec(codec string) bool {
	return (ValidCodecs(codec, stppCodec) ||
		ValidCodecs(codec, wvttCodec))
		
// ValidBitrateRange returns true if the specified min and max bitrates create a valid range
func ValidBitrateRange(minBitrate int, maxBitrate int) bool {
	return (minBitrate >= 0 && maxBitrate <= math.MaxInt32) &&
		(minBitrate < maxBitrate) &&
		!(minBitrate == 0 && maxBitrate == math.MaxInt32)
}

// DefinesBitrateFilter returns true if a bitrate filter should be applied. This means that
// at least one of the overall, audio, and video bitrate filters are valid and not the default range
func DefinesBitrateFilter(f *parsers.MediaFilters) bool {
	overall := ValidBitrateRange(f.MinBitrate, f.MaxBitrate)
	audio := ValidBitrateRange(f.AudioFilters.MinBitrate, f.AudioFilters.MaxBitrate)
	video := ValidBitrateRange(f.VideoFilters.MinBitrate, f.VideoFilters.MaxBitrate)
	if overall {
		if audio && (f.AudioFilters.MinBitrate > f.MaxBitrate || f.AudioFilters.MaxBitrate < f.MinBitrate) {
			f.AudioFilters.MinBitrate = f.MinBitrate
			f.AudioFilters.MaxBitrate = f.MaxBitrate
		} else if video && (f.VideoFilters.MinBitrate > f.MaxBitrate || f.VideoFilters.MaxBitrate < f.MinBitrate) {
			f.VideoFilters.MinBitrate = f.MinBitrate
			f.VideoFilters.MaxBitrate = f.MaxBitrate
		} else {
			f.AudioFilters.MinBitrate = max(f.AudioFilters.MinBitrate, f.MinBitrate)
			f.AudioFilters.MaxBitrate = min(f.AudioFilters.MaxBitrate, f.MaxBitrate)
			f.VideoFilters.MinBitrate = max(f.VideoFilters.MinBitrate, f.MinBitrate)
			f.VideoFilters.MaxBitrate = min(f.VideoFilters.MaxBitrate, f.MaxBitrate)
		}
		return true
	} else {
		return audio || video
	}
}

// max returns the larger of int a and int b
func max(a int, b int) int {
	if a < b {
		return b
	}
	return a
}

// min returns the smaller of int a and int b
func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
