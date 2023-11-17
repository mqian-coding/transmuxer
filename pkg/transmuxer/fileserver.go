package transmuxer

type FileServer struct {
	TempDir   string
	StaticDir string
}

var TheServer *FileServer

func NewFileServer(tmpDir, staticDir string) *FileServer {
	return &FileServer{
		TempDir:   tmpDir,
		StaticDir: staticDir,
	}
}

func IsServerInitialized() bool {
	return TheServer != nil
}
