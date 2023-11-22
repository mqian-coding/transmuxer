package store

import (
	"log"
	"os"
	"path"

	"github.com/google/uuid"
)

type FileServer struct {
	StaticDir string
	TempDir   string
}

var TheServer *FileServer

func NewFileServer(staticDirName string) (*FileServer, error) {
	var err error
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	staticDir := path.Join(homeDir, staticDirName)
	if _, err = os.Stat(staticDir); err != nil {
		if os.Mkdir(staticDir, 0755) != nil {
			return nil, err
		}
	}

	tmpDir := path.Join(os.TempDir(), uuid.New().String())
	if err = os.Mkdir(tmpDir, 0755); err != nil {
		return nil, err
	}

	return &FileServer{
		StaticDir: staticDir,
		TempDir:   tmpDir,
	}, nil
}

func (f *FileServer) Cleanup() {
	log.Printf("START: Cleaning up files...")
	if IsServerInitialized() {
		os.RemoveAll(f.TempDir)
	}
	log.Printf("DONE: Cleaned up files")
}

func IsServerInitialized() bool {
	return TheServer != nil
}
