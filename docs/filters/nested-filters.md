---
title: Nested Filters
parent: Filters
nav_order: 5
---

# Nested Filters
A way to apply <a href="codec.html">codec</a> and <a href="bandwidth.html">bandwidth</a> filters to a specific type of content. The nested codec and bandwidth filters behave like their non-nested versions.


## Protocol Support

HLS | DASH |
:--:|:----:|
yes | yes  |

## Supported Values

| content types | example                     |
|:-------------:|:---------------------------:|
| audio         | a(codec(ac-3),b(0,1000000)) |
| video         | v(codec(avc),b(0,1000000))  |

| subfilters | example         |
|:----------:|:---------------:|
| codec      | a(codec(ac-3))  |
| bitrate    | v(b(0,1000000)) |

## Usage Example
### Single Nested Filter:

    // Removes MPEG-4 audio
    $ http http://bakery.dev.cbsivideo.com/a(codec(mp4a))/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video
    $ http http://bakery.dev.cbsivideo.com/v(codec(avc))/star_trek_discovery/S01/E01.m3u8

    // Removes audio outside of 500Kbps to 1Mbps
    $ http http://bakery.dev.cbsivideo.com/a(b(500000,1000000))/star_trek_discovery/S01/E01.m3u8

    // Removes video outside of 500Kbps to 1Mbps
    $ http http://bakery.dev.cbsivideo.com/v(b(500000,1000000))/star_trek_discovery/S01/E01.m3u8

### Multiple Nested Filter:
To use multiple nested filters, separate with `,` with no space between nested filters.

    // Removes MPEG-4 audio and audio not in range of 500Kbps to 1Mbps
    $ http http://bakery.dev.cbsivideo.com/a(codec(mp4a),b(500000,1000000))/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video and video not in range of 500Kbps to 1Mbps
    $ http http://bakery.dev.cbsivideo.com/v(codec(avc),b(500000,1000000))/star_trek_discovery/S01/E01.m3u8

### Multiple Filters:
To use multiple filters, separated with `/` with no space between filters. You can use nested filters in conjunction with the general filters, such as the bandwidth filter.

    // Removes AVC video, MPEG-4 audio, audio not in range of 500Kbps to 1Mbps
    $ http http://bakery.dev.cbsivideo.com/v(codec(avc))/a(codec(mp4a),b(500000,1000000))/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video, MPEG-4 audio, and everything not in range of 500Kbps to 1Mbps
    $ http http://bakery.dev.cbsivideo.com/v(codec(avc))/a(codec(mp4a))/b(500000,1000000)/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video, all video not in range 750Kbps to 1Mbps, MPEG-4 audio, and non-video not in range of 500Kbps to 1Mbps
    $ http http://bakery.dev.cbsivideo.com/v(codec(avc),b(750000))/a(codec(mp4a))/b(500000,1000000)/star_trek_discovery/S01/E01.m3u8