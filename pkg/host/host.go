package host

import (
	"concurrency-practice/internal/store"
	"concurrency-practice/pkg/utils"
	"errors"
	"fmt"
	"log"
	"os"
)

type PlayInput struct {
	PlaylistURL string
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
	media, mediaPlaylistURL, err := utils.ParseAsMediaPlaylist(in.PlaylistURL, 0)
	if media == nil {
		return errors.New("media playlist cannot be nil")
	}
	if err != nil {
		return err
	}

	// Enrich Segment URLs with domain name
	domain, err := utils.GetSegmentURLPrefix(mediaPlaylistURL)
	if err != nil {
		return err
	}

	// Enrich the copied playlist
	utils.EnrichMediaPlaylistSegments(media, domain)

	// Download all the segments in the media playlist
	if err = manager.downloadSegments(media, "segment"); err != nil {
		return err
	}

	// repoint the segment uris for serving
	utils.NormalizeMediaPlaylistSegments(media)

	if err = manager.saveSegmentsAndPlaylist(GetPlaylistDir(in.Filename), GetSegmentsDir(in.Filename), media); err != nil {
		return err
	}

	log.Printf(fmt.Sprintf("playlist download request complete: url: %s, name: %s", in.PlaylistURL, in.Filename))
	return nil
}
