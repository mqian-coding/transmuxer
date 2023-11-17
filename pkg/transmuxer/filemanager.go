package transmuxer

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
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type FileManager struct {
	ParentDir   string
	SegmentsDir string
	OutputDir   string
}

const maxRetryAttempts = 5

func NewFileManager(tmpDir string) (*FileManager, error) {
	var err error
	var newDirs []string
	myDir := tmpDir + "/" + uuid.New().String()
	newDirs = append(newDirs, myDir)
	newDirs = append(newDirs, myDir+"/segments")
	newDirs = append(newDirs, myDir+"/output")
	for _, dir := range newDirs {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	return &FileManager{
		ParentDir:   myDir,
		SegmentsDir: myDir + "/segments",
		OutputDir:   myDir + "/output",
	}, nil
}

func (f *FileManager) Cleanup() error {
	return os.RemoveAll("/" + f.ParentDir)
}

func (f *FileManager) DownloadSegments(media *m3u8.MediaPlaylist, segmentNamePrefix string) error {
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
				if err = f.downloadSegment(s, segmentNamePrefix); err != nil {
					log.Printf(fmt.Sprintf("segment %d errored on attempt: %d of %d", s.SeqId, i+1, maxRetryAttempts+1))
				} else {
					break
				}
				if err != nil {
					time.Sleep(time.Duration(math.Pow(2, float64(i))))
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

func (f *FileManager) downloadSegment(seg *m3u8.MediaSegment, segmentNamePrefix string) error {
	resp, err := http.Get(seg.URI)
	if err != nil {
		return err
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		return errors.New("was not 200 OK response")
	}
	defer resp.Body.Close()

	// Create a file to save the segment
	file, err := os.Create(utils.GetSegmentFileName(f.SegmentsDir, segmentNamePrefix, seg.SeqId))
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

func (f *FileManager) SegmentsToMKV() error {
	return utils.TransmuxToMKV(f.ParentDir)
}

func (f *FileManager) copyFFMPEGBinary() error {
	ffmpegBinaryPath := "bin/ffmpeg/ffmpeg"
	if _, err := os.Stat("bin/ffmpeg/ffmpeg"); err != nil {
		return err
	}
	ffmpegDestPath := filepath.Join(f.SegmentsDir, "ffmpeg")
	if err := copyFile(ffmpegBinaryPath, ffmpegDestPath); err != nil {
		return err
	}
	if err := os.Chmod(ffmpegDestPath, 0755); err != nil {
		return err
	}
	return nil
}

// copyFile copies a file from source to destination
func copyFile(source, destination string) error {
	input, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	if err = os.WriteFile(destination, input, 0644); err != nil {
		return err
	}
	return nil
}