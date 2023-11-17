package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func TransmuxToMKV(directoryPath string) error {
	ffmpegBinaryPath := filepath.Join(directoryPath, "segments", "ffmpeg")
	outputMKVFilePath := filepath.Join(directoryPath, "output", "output.mkv")
	inputFilesPattern := filepath.Join(directoryPath, "segments", "segment_*.ts")
	// Get a list of all files in the input directory
	files, err := filepath.Glob(inputFilesPattern)
	if err != nil {
		return err
	}

	input, err := os.Create(filepath.Join(directoryPath, "segments", "input.txt"))
	if err != nil {
		return err
	}
	defer os.Remove(input.Name())
	for _, file := range files {
		input.WriteString("file '" + filepath.Base(file) + "'\n")
	}
	input.Close()
	// Run FFmpeg with the file list and fixed framerate
	cmd := exec.Command(
		ffmpegBinaryPath,
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
