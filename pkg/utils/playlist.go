package utils

import (
	"bufio"
	"errors"
	"github.com/grafov/m3u8"
	"log"
	"net/http"
	"path"
)

func ParseAsMediaPlaylist(u string, depth int) (*m3u8.MediaPlaylist, string, error) {
	if depth > 1 {
		return nil, "", errors.New("failed to unwrap master playlist into mediaplaylist")
	}
	if !IsValidManifestURL(u) {
		return nil, "", errors.New("invalid manifest url, must end with .m3u8")
	}
	resp, err := http.Get(u)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	pl, plType, err := m3u8.DecodeFrom(bufio.NewReader(resp.Body), true)
	if err != nil {
		return nil, "", err
	}

	var variant *m3u8.Variant
	switch plType {
	case m3u8.MASTER:
		master := pl.(*m3u8.MasterPlaylist)
		if len(master.Variants) == 0 {
			return nil, "", errors.New("master playlist must have at least one variant")
		}
		variant = master.Variants[0]
		for _, v := range master.Variants {
			if v.Bandwidth > variant.Bandwidth {
				variant = v
			}
		}
		if variant.Chunklist != nil {
			return variant.Chunklist, u, nil
		}
		log.Printf("unwrapping master playlist...")
		return ParseAsMediaPlaylist(variant.URI, depth+1)
	case m3u8.MEDIA:
		return pl.(*m3u8.MediaPlaylist), u, nil
	}

	return nil, "", errors.New("neither master playlist nor media playlist")
}

func EnrichMediaPlaylistSegments(media *m3u8.MediaPlaylist, dir string) {
	for i, s := range media.Segments {
		if s == nil {
			break
		}
		media.Segments[i].URI = EnrichSegmentWithDir(dir, s.URI)
	}
}

func NormalizeMediaPlaylistSegments(media *m3u8.MediaPlaylist) {
	for i, s := range media.Segments {
		if s == nil {
			break
		}
		media.Segments[i].URI = path.Join(GetSegmentFileNameNoDirNoExt(s.SeqId), "seg.ts")
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
