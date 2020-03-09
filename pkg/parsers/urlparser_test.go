package parsers

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

func TestURLParseUrl(t *testing.T) {
	tests := []struct {
		name                 string
		input                string
		expectedFilters      MediaFilters
		expectedManifestPath string
		expectedErr          bool
	}{
		{
			"one video type",
			"/v(hdr10)/",
			MediaFilters{
				Videos:     []VideoType{"hev1.2", "hvc1.2"},
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
				VideoSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []Codec{"hev1.2", "hvc1.2"},
				},
				AudioSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			"/",
			false,
		},
		{
			"two video types",
			"/v(hdr10,hevc)/",
			MediaFilters{
				Videos:     []VideoType{"hev1.2", "hvc1.2", videoHEVC},
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
				VideoSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []Codec{"hev1.2", "hvc1.2", codecHEVC},
				},
				AudioSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			"/",
			false,
		},
		{
			"two video types and two audio types",
			"/v(hdr10,hevc)/a(aac,noAd)/",
			MediaFilters{
				Videos:     []VideoType{"hev1.2", "hvc1.2", videoHEVC},
				Audios:     []AudioType{audioAAC, audioNoAudioDescription},
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
				VideoSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []Codec{"hev1.2", "hvc1.2", codecHEVC},
				},
				AudioSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []Codec{codecAAC, codecNoAudioDescription},
				},
			},
			"/",
			false,
		},
		{
			"videos, audio, captions and bitrate range",
			"/v(hdr10,hevc)/a(aac)/al(pt-BR,en)/c(en)/b(100,4000)/",
			MediaFilters{
				Videos:           []VideoType{"hev1.2", "hvc1.2", videoHEVC},
				Audios:           []AudioType{audioAAC},
				AudioLanguages:   []AudioLanguage{audioLangPTBR, audioLangEN},
				CaptionLanguages: []CaptionLanguage{captionEN},
				MaxBitrate:       4000,
				MinBitrate:       100,
				VideoSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []Codec{"hev1.2", "hvc1.2", codecHEVC},
				},
				AudioSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []Codec{codecAAC},
				},
			},
			"/",
			false,
		},
		{
			"bitrate range with minimum bitrate only",
			"/b(100,)/",
			MediaFilters{
				MaxBitrate: math.MaxInt32,
				MinBitrate: 100,
				VideoSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			"/",
			false,
		},
		{
			"bitrate range with maximum bitrate only",
			"/b(,3000)/",
			MediaFilters{
				MaxBitrate: 3000,
				MinBitrate: 0,
			},
			"/",
			false,
		},
		{
			"bitrate range with minimum greater than maximum throws error",
			"/b(30000,3000)/",
			MediaFilters{},
			"",
			true,
		},
		{
			"bitrate range with minimum equal to maximum throws error",
			"/b(3000,3000)/",
			MediaFilters{},
			"",
			true,
		},
		{
			"trim filter",
			"/t(100,1000)/path/to/test.m3u8",
			MediaFilters{
				Protocol:   ProtocolHLS,
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
				Trim: &Trim{
					Start: 100,
					End:   1000,
				},
			},
			"/path/to/test.m3u8",
			false,
		},
		{
			"trim filter where start time is greater than end time throws error",
			"/t(10000,1000)/path/to/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"trim filter where start time and end time are equal throws error",
			"/t(10000,1000)/path/to/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"detect a signle plugin for execution from url",
			"[plugin1]/some/path/master.m3u8",
			MediaFilters{
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
				Protocol:   ProtocolHLS,
				Plugins:    []string{"plugin1"},
			},
			"/some/path/master.m3u8",
			false,
		},
		{
			"detect plugins for execution from url",
			"/v(hdr10,hevc)/[plugin1,plugin2,plugin3]/some/path/master.m3u8",
			MediaFilters{
				Videos:     []VideoType{"hev1.2", "hvc1.2", videoHEVC},
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
				Protocol:   ProtocolHLS,
				Plugins:    []string{"plugin1", "plugin2", "plugin3"},
			},
			"/some/path/master.m3u8",
			false,
		},
		{
			"bitrate range and audio specifier",
			"/b(audio,0,1000)",
			MediaFilters{
				FilterBitrateTypes: []StreamType{"audio"},
				MinBitrate:         0,
				MaxBitrate:         1000,
			},
			"/",
		},
		{
			"bitrate range and audio and video specifiers",
			"/b(audio,video,0,1000)",
			MediaFilters{
				FilterBitrateTypes: []StreamType{"audio", "video"},
				MinBitrate:         0,
				MaxBitrate:         1000,
				VideoSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			"/",
		},
		{
			"nested audio and video bitrate filters",
			"/a(b(100,))/v(b(,5000))/",
			MediaFilters{
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
				VideoSubFilters: Subfilters{
					MaxBitrate: 5000,
				},
				AudioSubFilters: Subfilters{
					MinBitrate: 100,
					MaxBitrate: math.MaxInt32,
				},
			},
			"/",
		},
		{
			"nested codec and bitrate filters in audio",
			"/a(b(100,200),c(ac-3,aac))/",
			MediaFilters{
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
				VideoSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioSubFilters: Subfilters{
					MinBitrate: 100,
					MaxBitrate: 200,
					Codecs:     []Codec{codecAC3, codecAAC},
				},
			},
			"/",
		},
		{
			"nested codec and bitrate filters in video, plus overall bitrate filters",
			"/v(c(avc,hdr10),b(1000,2000))/",
			MediaFilters{
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
				VideoSubFilters: Subfilters{
					MaxBitrate: 2000,
					MinBitrate: 1000,
					Codecs:     []Codec{codecH264, "hev1.2", "hvc1.2"},
				},
				AudioSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			"/",
		},
		{
			"detect protocol hls for urls with .m3u8 extension",
			"/path/here/with/master.m3u8",
			MediaFilters{
				Protocol:   ProtocolHLS,
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
				VideoSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			"/path/here/with/master.m3u8",
			false,
		},
		{
			"detect protocol dash for urls with .mpd extension",
			"/path/here/with/manifest.mpd",
			MediaFilters{
				Protocol:   ProtocolDASH,
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
				VideoSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioSubFilters: Subfilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			"/path/here/with/manifest.mpd",
			false,
		},
		{
			"detect filters for propeller channels and set path properly",
			"/v(avc)/a(aac)/propeller/orgID/master.m3u8",
			MediaFilters{
				Videos:     []VideoType{videoH264},
				Audios:     []AudioType{audioAAC},
				Protocol:   ProtocolHLS,
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
			},
			"/propeller/orgID/master.m3u8",
			false,
		},
		{
			"set path properly for propeller channel with no filters",
			"/propeller/orgID/master.m3u8",
			MediaFilters{
				Protocol:   ProtocolHLS,
				MaxBitrate: math.MaxInt32,
				MinBitrate: 0,
			},
			"/propeller/orgID/master.m3u8",
			false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			masterManifestPath, output, err := URLParse(test.input)
			if !test.expectedErr && err != nil {
				t.Errorf("Did not expect an error returned, got: %v", err)
				return
			} else if test.expectedErr && err == nil {
				t.Errorf("Expected an error returned, got nil")
				return
			}

			jsonOutput, err := json.Marshal(output)
			if err != nil {
				t.Fatal(err)
			}

			jsonExpected, err := json.Marshal(test.expectedFilters)
			if err != nil {
				t.Fatal(err)
			}

			if test.expectedManifestPath != masterManifestPath {
				t.Errorf("wrong master manifest generated.\nwant %#v\n\ngot %#v", test.expectedManifestPath, masterManifestPath)
			}

			if !reflect.DeepEqual(jsonOutput, jsonExpected) {
				t.Errorf("wrong struct generated.\nwant %#v\ngot %#v", test.expectedFilters, output)
			}
		})
	}
}
