package host

import (
	"concurrency-practice/pkg/utils"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/grafov/m3u8"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type FileManager struct {
	ParentDir   string
	SegmentsDir string
}

const maxRetryAttempts = 10
const retryBase = 1.25

func NewFileManager(tmpDir string) (*FileManager, error) {
	var err error
	var newDirs []string
	myDir := tmpDir + "/" + uuid.New().String()
	newDirs = append(newDirs, myDir)
	newDirs = append(newDirs, myDir+"/segments")
	for _, dir := range newDirs {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	return &FileManager{
		ParentDir:   myDir,
		SegmentsDir: myDir + "/segments",
	}, nil
}

func (f *FileManager) Cleanup() error {
	return os.RemoveAll("/" + f.ParentDir)
}

func (f *FileManager) downloadSegments(media *m3u8.MediaPlaylist, segmentNamePrefix string) error {
	var wg sync.WaitGroup
	var mu sync.Mutex

	errs := make(chan error, len(media.Segments))
	segsRemaining := int64(utils.NumSegs(media))
	wg.Add(int(segsRemaining))
	log.Printf(fmt.Sprintf("begin processing %d segments ", atomic.LoadInt64(&segsRemaining)))

	for _, s := range media.Segments {
		if s == nil {
			break
		}
		go func(s *m3u8.MediaSegment) {
			var err error
			for i := 0; i <= maxRetryAttempts; i++ {
				if err = f.downloadSegment(s); err != nil {
					log.Printf(fmt.Sprintf("segment %d errored on attempt: %d of %d", s.SeqId, i+1, maxRetryAttempts+1))
				} else {
					break
				}
				if err != nil {
					time.Sleep(time.Duration(math.Pow(retryBase, float64(i))))
				}
			}
			if err != nil {
				errs <- err
			}

			wg.Done()
			mu.Lock()
			segsRemaining--
			log.Printf(fmt.Sprintf("segments remaining: %v", strconv.FormatInt(segsRemaining, 10)))
			mu.Unlock()
		}(s)
	}
	wg.Wait()
	close(errs)
	select {
	case err := <-errs:
		if err != nil {
			return err
		}
	default:
	}
	return nil
}

func (f *FileManager) downloadSegment(seg *m3u8.MediaSegment) error {
	resp, err := http.Get(seg.URI)
	if err != nil {
		return err
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		return errors.New("was not 200 OK response")
	}
	defer resp.Body.Close()

	// Create a file to save the segment
	file, err := os.Create(utils.GetSegmentFileName(f.SegmentsDir, seg.SeqId))
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the segment content to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(file.Name())
		return err
	}
	return nil
}

func (f *FileManager) saveSegmentsAndPlaylists(outputManifestDir, outputMediaSegmentsDir string, media *m3u8.MediaPlaylist, master *m3u8.MasterPlaylist) error {
	if err := func(outputManifestDir, outputMediaSegmentsDir string, media *m3u8.MediaPlaylist) error {
		// Save chunklist and master playlists
		if outputManifestDir == "" {
			return errors.New("file server has no valid output path for hosting media")
		}
		if err := os.MkdirAll(outputManifestDir, 0755); err != nil {
			return err
		}
		if master == nil {
			return errors.New("master playlist cannot be nil")
		}
		if media == nil {
			return errors.New("media playlist cannot be nil")
		}

		masterPlaylist, err := os.Create(outputManifestDir + "/playlist.m3u8")
		if err != nil {
			return err
		}
		if _, err = masterPlaylist.WriteString(master.String()); err != nil {
			return err
		}

		chunklistPlaylist, err := os.Create(outputManifestDir + "/chunklist.m3u8")
		if err != nil {
			return err
		}
		if _, err = chunklistPlaylist.WriteString(media.String()); err != nil {
			return err
		}

		// Save Segments
		if outputMediaSegmentsDir == "" {
			return errors.New("file server has no valid output path for hosting media segments")
		}
		if err = os.MkdirAll(outputMediaSegmentsDir, 0755); err != nil {
			return err
		}
		return utils.CopyDir(outputMediaSegmentsDir, f.SegmentsDir)
	}(outputManifestDir, outputMediaSegmentsDir, media); err != nil {
		os.RemoveAll(outputManifestDir)
		os.RemoveAll(outputMediaSegmentsDir)
		return err
	}
	return nil
}
