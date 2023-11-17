package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
