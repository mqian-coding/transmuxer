package transmuxer

import (
	"concurrency-practice/pkg/utils"
	"errors"
	"fmt"
	"log"
	"os"
)

type TransmuxInput struct {
	PlaylistURL           string
	OutputName            string
	NormalizeSegmentNames bool
}

func Transmux(in TransmuxInput) error {
	log.Printf(fmt.Sprintf("transmux request received: url: %s, name: %s", in.PlaylistURL, in.OutputName))

	if !IsServerInitialized() {
		return errors.New("the transmuxer file server is not initialized")
	}
	manager, err := NewFileManager(TheServer.TempDir)
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
	utils.NormalizeMediaPlaylistSegments(media, domain)

	// Download all the segments in the media playlist
	if err = manager.DownloadSegments(media, "segment"); err != nil {
		return err
	}

	// Stitch into mkv
	if err = manager.SegmentsToMKV(in.OutputName, TheServer.StaticDir); err != nil {
		return err
	}
	log.Printf(fmt.Sprintf("transmux request complete: url: %s, name: %s", in.PlaylistURL, in.OutputName))
	return nil
}
