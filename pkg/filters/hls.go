package filters

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cbsinteractive/bakery/pkg/config"
	"github.com/cbsinteractive/bakery/pkg/parsers"
	"github.com/grafov/m3u8"
)

type execPluginHLS func(variant *m3u8.Variant)

// HLSFilter implements the Filter interface for HLS
// manifests
type HLSFilter struct {
	manifestURL     string
	manifestContent string
	config          config.Config
}

var matchFunctions = map[ContentType]func(string) bool{
	audioContentType:   isAudioCodec,
	videoContentType:   isVideoCodec,
	captionContentType: isCaptionCodec,
}

// NewHLSFilter is the HLS filter constructor
func NewHLSFilter(manifestURL, manifestContent string, c config.Config) *HLSFilter {
	return &HLSFilter{
		manifestURL:     manifestURL,
		manifestContent: manifestContent,
		config:          c,
	}
}

// FilterManifest will be responsible for filtering the manifest
// according  to the MediaFilters
func (h *HLSFilter) FilterManifest(filters *parsers.MediaFilters) (string, error) {
	m, manifestType, err := m3u8.DecodeFrom(strings.NewReader(h.manifestContent), true)
	if err != nil {
		return "", err
	}

	if manifestType != m3u8.MASTER {
		return h.filterRenditionManifest(filters, m.(*m3u8.MediaPlaylist))
	}

	// convert into the master playlist type
	manifest := m.(*m3u8.MasterPlaylist)
	filteredManifest := m3u8.NewMasterPlaylist()

	for _, v := range manifest.Variants {
		absoluteURL, _ := filepath.Split(h.manifestURL)
		absolute, aErr := url.Parse(absoluteURL)
		if aErr != nil {
			return h.manifestContent, aErr
		}

		normalizedVariant, err := h.normalizeVariant(v, *absolute)
		if err != nil {
			return "", err
		}

		validatedFilters, err := h.validateVariants(filters, normalizedVariant)
		if err != nil {
			return "", err
		}

		if validatedFilters {
			continue
		}

		uri := normalizedVariant.URI
		if filters.Trim != nil {
			uri, err = h.normalizeTrimmedVariant(filters, uri)
			if err != nil {
				return "", err
			}
		}

		filteredManifest.Append(uri, normalizedVariant.Chunklist, normalizedVariant.VariantParams)
	}

	return filteredManifest.String(), nil
}

// Returns true if specified variant should be removed from filter
func (h *HLSFilter) validateVariants(filters *parsers.MediaFilters, v *m3u8.Variant) (bool, error) {
	if filters.DefinesBitrateFilter() {
		if !(h.validateBandwidthVariant(filters.MinBitrate, filters.MaxBitrate, v)) {
			return true, nil
		}
	}

	variantCodecs := strings.Split(v.Codecs, ",")

	if filters.Audios != nil {
		supportedAudioTypes := map[string]struct{}{}
		for _, at := range filters.Audios {
			supportedAudioTypes[string(at)] = struct{}{}
		}
		res, err := validateVariantCodecs(audioContentType, variantCodecs, supportedAudioTypes, matchFunctions)
		if res {
			return true, err
		}
	}

	if filters.Videos != nil {
		supportedVideoTypes := map[string]struct{}{}
		for _, vt := range filters.Videos {
			supportedVideoTypes[string(vt)] = struct{}{}
		}
		res, err := validateVariantCodecs(videoContentType, variantCodecs, supportedVideoTypes, matchFunctions)
		if res {
			return true, err
		}
	}

	if filters.CaptionTypes != nil {
		supportedCaptionTypes := map[string]struct{}{}
		for _, ct := range filters.CaptionTypes {
			supportedCaptionTypes[string(ct)] = struct{}{}
		}
		res, err := validateVariantCodecs(captionContentType, variantCodecs, supportedCaptionTypes, matchFunctions)
		if res {
			return true, err
		}
	}

	return false, nil
}

// Returns true if the given variant (variantCodecs) should be allowed filtered out for supportedCodecs of filterType
func validateVariantCodecs(filterType ContentType, variantCodecs []string, supportedCodecs map[string]struct{}, supportedFilterTypes map[ContentType]func(string) bool) (bool, error) {
	var matchFilterType func(string) bool

	matchFilterType, found := supportedFilterTypes[filterType]

	if !found {
		return false, errors.New("filter type is unsupported")
	}

	variantFound := false
	for _, codec := range variantCodecs {
		if matchFilterType(codec) {
			for sc := range supportedCodecs {
				if ValidCodecs(codec, CodecFilterID(sc)) {
					variantFound = true
					break
				}
			}
		}
	}

	return variantFound, nil
}

func (h *HLSFilter) validateBandwidthVariant(minBitrate int, maxBitrate int, v *m3u8.Variant) bool {
	bw := int(v.VariantParams.Bandwidth)
	if bw > maxBitrate || bw < minBitrate {
		return false
	}

	return true
}

func (h *HLSFilter) normalizeVariant(v *m3u8.Variant, absolute url.URL) (*m3u8.Variant, error) {
	for _, a := range v.VariantParams.Alternatives {
		aURL, aErr := combinedIfRelative(a.URI, absolute)
		if aErr != nil {
			return v, aErr
		}
		a.URI = aURL
	}

	vURL, vErr := combinedIfRelative(v.URI, absolute)
	if vErr != nil {
		return v, vErr
	}
	v.URI = vURL
	return v, nil
}

func (h *HLSFilter) normalizeTrimmedVariant(filters *parsers.MediaFilters, uri string) (string, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(uri))
	start := filters.Trim.Start
	end := filters.Trim.End
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v://%v/t(%v,%v)/%v.m3u8", u.Scheme, h.config.Hostname, start, end, encoded), nil
}

func combinedIfRelative(uri string, absolute url.URL) (string, error) {
	if len(uri) == 0 {
		return uri, nil
	}
	relative, err := isRelative(uri)
	if err != nil {
		return uri, err
	}
	if relative {
		combined, err := absolute.Parse(uri)
		if err != nil {
			return uri, err
		}
		return combined.String(), err
	}
	return uri, nil
}

func isRelative(urlStr string) (bool, error) {
	u, e := url.Parse(urlStr)
	if e != nil {
		return false, e
	}
	return !u.IsAbs(), nil
}

// FilterRenditionManifest will be responsible for filtering the manifest
// according  to the MediaFilters
func (h *HLSFilter) filterRenditionManifest(filters *parsers.MediaFilters, m *m3u8.MediaPlaylist) (string, error) {
	filteredPlaylist, _ := m3u8.NewMediaPlaylist(uint(len(m.Segments)), uint(len(m.Segments)))
	seqID := 0

	for _, segment := range m.Segments {
		if segment != nil {
			if inRange(int64(filters.Trim.Start), int64(filters.Trim.End), segment.ProgramDateTime.Unix()) {
				isRelativeURL, err := isRelative(segment.URI)
				if err != nil {
					return "", fmt.Errorf("trimming segments: %w", err)
				}

				if isRelativeURL {
					absolute, err := getAbsoluteURLString(h.manifestURL)
					if err != nil {
						return "", fmt.Errorf("fetching absolute url: %w", err)
					}

					segment.URI = absolute + segment.URI
				}

				err = filteredPlaylist.AppendSegment(segment)
				if err != nil {
					return "", fmt.Errorf("trimming segments: %w", err)
				}

				seqID++
			}
		}
	}

	filteredPlaylist.SeqNo = uint64(seqID)
	filteredPlaylist.Close()

	return filteredPlaylist.Encode().String(), nil
}

func inRange(start int64, end int64, value int64) bool {
	return (start <= value) && (value <= end)
}

//Returns absolute url of given manifest as a string
func getAbsoluteURLString(path string) (string, error) {
	absoluteURL, _ := filepath.Split(path)
	absolute, aErr := url.Parse(absoluteURL)
	if aErr != nil {
		return path, aErr
	}

	return absolute.String(), nil
}
