package host

import (
	"errors"
	"fmt"
	"github.com/grafov/m3u8"
	"log"
	"os"
	"path"
	"transmuxer/internal/store"
	"transmuxer/pkg/utils"
)

type PlayInput struct {
	PlaylistURL string
	CaptionsURL string
	Filename    string
}

func GeneratePlaylist(in PlayInput) error {
	log.Printf(fmt.Sprintf("playlist request received: url: %s, name: %s", in.PlaylistURL, in.Filename))

	if !store.IsServerInitialized() {
		return errors.New("the file server is not initialized")
	}

	manager, err := NewFileManager(store.TheServer.TempDir)
	defer manager.Cleanup()
	if err != nil {
		return err
	}

	log.Printf("created new file manager at directory: %s", manager.ParentDir)

	defer func() {
		if _, err := os.Stat(manager.ParentDir); err != nil {
			log.Println(err.Error())
		}
		if err := os.RemoveAll(manager.ParentDir); err != nil {
			log.Println(err.Error())
		}
	}()

	// Get the Media Playlist
	media, _, err := utils.ParseAsMediaPlaylist(in.PlaylistURL, 0)
	if media == nil {
		return errors.New("media playlist cannot be nil")
	}
	if err != nil {
		return err
	}

	// Download all the segments in the media playlist
	if err = manager.downloadSegments(media, "segment"); err != nil {
		return err
	}

	// repoint the segment uris for serving
	utils.NormalizeMediaPlaylistSegments(media)

	var variants []*m3u8.Variant
	// if there was a captions file, attach it
	variants = append(variants, &m3u8.Variant{
		URI:       path.Join("http://localhost:8080", in.Filename, "chunklist.m3u8"),
		Chunklist: media,
	})

	master := &m3u8.MasterPlaylist{
		Variants: variants,
	}
	master.SetVersion(3)

	if err = manager.saveSegmentsAndPlaylists(GetManifestPath(in.Filename), GetSegmentsDir(in.Filename), media, master); err != nil {
		return err
	}

	log.Printf(fmt.Sprintf("playlist download request complete: url: %s, name: %s", in.PlaylistURL, in.Filename))
	return nil
}
