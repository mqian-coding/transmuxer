package utils

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func TransmuxToMKV(directoryPath, name, staticDir string) error {
	outputMKVFilePath := staticDir + "/" + name + ".mkv"

	input, err := os.Create(filepath.Join(directoryPath, "segments", "input.txt"))
	if err != nil {
		return err
	}
	defer os.Remove(input.Name())

	inputFilesPattern := filepath.Join(directoryPath, "segments", "segment_*.ts")
	files, err := filepath.Glob(inputFilesPattern)
	if err != nil {
		return err
	}
	for _, file := range files {
		input.WriteString("file '" + filepath.Base(file) + "'\n")
	}
	input.Close()

	cmd := exec.Command(
		filepath.Join(directoryPath, "segments", "ffmpeg"),
		"-f", "concat",
		"-i", input.Name(),
		"-c", "copy",
		outputMKVFilePath,
	)
	log.Println(cmd.String())
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("error running ffmpeg command: %v", err)
	}

	log.Printf("Transmuxing completed. Output file: %s\n", outputMKVFilePath)
	return nil
}

// CopyDir copies the content of src to dst. src should be a full path.
func CopyDir(dst, src string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// copy to this path
		outpath := filepath.Join(dst, strings.TrimPrefix(path, src))

		if info.IsDir() {
			os.MkdirAll(outpath, info.Mode())
			return nil // means recursive
		}

		// handle irregular files
		if !info.Mode().IsRegular() {
			switch info.Mode().Type() & os.ModeType {
			case os.ModeSymlink:
				link, err := os.Readlink(path)
				if err != nil {
					return err
				}
				return os.Symlink(link, outpath)
			}
			return nil
		}

		// copy contents of regular file efficiently

		// open input
		in, _ := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		// create output
		fh, err := os.Create(outpath)
		if err != nil {
			return err
		}
		defer fh.Close()

		// make it the same
		fh.Chmod(info.Mode())

		// copy content
		_, err = io.Copy(fh, in)
		return err
	})
}

// CopyFile copies a file from source to destination
func CopyFile(source, destination string) error {
	input, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err = os.WriteFile(destination, input, 0644); err != nil {
		return err
	}
	return nil
}
