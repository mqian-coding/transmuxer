package utils

import (
	"bufio"
	"errors"
	"github.com/grafov/m3u8"
	"net/http"
)

func ParseAsMediaPlaylist(u string) (*m3u8.MediaPlaylist, error) {
	if !IsValidManifestURL(u) {
		return nil, errors.New("invalid manifest url, must end with .m3u8")
	}
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	pl, plType, err := m3u8.DecodeFrom(bufio.NewReader(resp.Body), true)
	media := pl.(*m3u8.MediaPlaylist)
	if plType != m3u8.MEDIA {
		return nil, errors.New("must be media playlist, not master")
	}
	return media, nil
}

func NormalizeMediaPlaylistSegments(media *m3u8.MediaPlaylist, dir string) {
	for i, s := range media.Segments {
		if s == nil {
			break
		}
		media.Segments[i].URI = EnrichSegmentWithDir(dir, s.URI)
	}
}

func NumSegs(media *m3u8.MediaPlaylist) int {
	var count int
	for _, s := range media.Segments {
		if s == nil {
			break
		}
		count++
	}
	return count
}

func FirstSeqId(media *m3u8.MediaPlaylist) int {
	for _, s := range media.Segments {
		if s != nil {
			return int(s.SeqId)
		}
	}
	return 0
}
