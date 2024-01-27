package transmuxer

import (
	"errors"
	"fmt"
	"log"
	"os"
	"transmuxer/internal/store"
	"transmuxer/pkg/utils"
)

type TransmuxInput struct {
	PlaylistURL string
	OutputName  string
}

func Transmux(in TransmuxInput) error {
	log.Printf(fmt.Sprintf("transmux request received: url: %s, name: %s", in.PlaylistURL, in.OutputName))

	if !store.IsServerInitialized() {
		return errors.New("the transmuxer file server is not initialized")
	}
	manager, err := NewFileManager(store.TheServer.TempDir)
	if err != nil {
		return err
	}
	log.Printf("created new file manager at directory: %s", manager.ParentDir)

	if err = manager.copyFFMPEGBinary(); err != nil {
		return err
	}
	log.Printf("copied ffmpeg binary to: %s", manager.ParentDir)

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
	utils.EnrichMediaPlaylistSegments(media, domain)

	// Download all the segments in the media playlist
	if err = manager.downloadSegments(media, "segment"); err != nil {
		return err
	}

	// Stitch into mkv
	if err = manager.SegmentsToMKV(in.OutputName, store.TheServer.StaticDir); err != nil {
		return err
	}
	log.Printf(fmt.Sprintf("transmux request complete: url: %s, name: %s", in.PlaylistURL, in.OutputName))
	return nil
}
