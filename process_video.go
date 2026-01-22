package main

import (
	"os/exec"
)

func processVideoForFastStart(filepath string) (string, error) {
	newOutput := filepath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filepath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", newOutput)

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return newOutput, nil
}