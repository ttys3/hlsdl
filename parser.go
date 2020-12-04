package hlsdl

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/grafov/m3u8"
)

func parseHlsSegments(hlsURL string) ([]*Segment, error) {
	baseURL, err := url.Parse(hlsURL)
	if err != nil {
		return nil, errors.New("Invalid m3u8 url")
	}

	p, t, err := getM3u8ListType(hlsURL)
	if err != nil {
		return nil, err
	}

	if t == m3u8.MASTER {
		if masterpl, ok := p.(*m3u8.MasterPlaylist); !ok {
			return nil, errors.New("got m3u8 master playlist but data not valid")
		} else if len(masterpl.Variants) == 0 {
			return nil, errors.New("got m3u8 master playlist but zero variants found")
		} else {
			allBw := make([]int, len(masterpl.Variants))
			allResolutions := make(map[int]*m3u8.Variant)
			fmt.Printf("%+v\n", masterpl.Variants[0])
			for k, v := range masterpl.Variants {
				allBw[k] = int(v.Bandwidth)
				allResolutions[int(v.Bandwidth)] = v
			}
			sort.Sort(sort.Reverse(sort.IntSlice(allBw)))
			bestVariant := allResolutions[allBw[0]]
			var variantURL *url.URL
			if !strings.Contains(bestVariant.URI, "http") {
				var err error
				variantURL, err = baseURL.Parse(bestVariant.URI)
				if err != nil {
					return nil, err
				}
			} else {
				variantURL, err = url.Parse(bestVariant.URI)
				if err != nil {
					return nil, errors.New("got m3u8 master playlist but bestVariant.URI failed")
				}
			}
			for k, _ := range allBw {
				v := allResolutions[allBw[k]]
				fmt.Printf("ProgramId=%v Resolution=%v Bandwidth=%v Codecs=%v FrameRate=%v\n",
					v.ProgramId, v.Resolution, v.Bandwidth, v.Codecs, v.FrameRate)
			}
			log.Printf("got m3u8 master playlist, use best variant resolution=%v URL=%s", bestVariant.Resolution, variantURL.String())
			return parseHlsSegments(variantURL.String())
		}
	}

	if t != m3u8.MEDIA {
		return nil, errors.New("No support the m3u8 format")
	}

	mediaList := p.(*m3u8.MediaPlaylist)
	segments := []*Segment{}
	for _, seg := range mediaList.Segments {
		if seg == nil {
			continue
		}

		if !strings.Contains(seg.URI, "http") {
			segmentURL, err := baseURL.Parse(seg.URI)
			if err != nil {
				return nil, err
			}

			seg.URI = segmentURL.String()
		}

		if seg.Key == nil && mediaList.Key != nil {
			seg.Key = mediaList.Key
		}

		if seg.Key != nil && !strings.Contains(seg.Key.URI, "http") {
			keyURL, err := baseURL.Parse(seg.Key.URI)
			if err != nil {
				return nil, err
			}

			seg.Key.URI = keyURL.String()
		}

		segment := &Segment{MediaSegment: seg}
		segments = append(segments, segment)
	}

	return segments, nil
}

func getM3u8ListType(hlsURL string) (m3u8.Playlist, m3u8.ListType, error) {
	res, err := http.Get(hlsURL)
	if err != nil {
		return nil, 0, err
	}

	if res.StatusCode != 200 {
		return nil, 0, errors.New(res.Status)
	}

	p, t, err := m3u8.DecodeFrom(res.Body, false)
	if err != nil {
		return nil, 0, err
	}

	return p, t, nil
}
