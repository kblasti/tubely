package main

import (
	"os/exec"
	"bytes"
	"encoding/json"
	"fmt"
)

func getVideoAspectRatio(filepath string) (string, error) {
	buf := &bytes.Buffer{}
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	data := buf.Bytes()
	type FFProbeStream struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}

	type FFProbeOutput struct {
		Streams []FFProbeStream `json:"streams"`
	}
	out := FFProbeOutput{}

	err = json.Unmarshal(data, &out)
	if err != nil {
		return "", err
	}

	if len(out.Streams) == 0 {
		return "", fmt.Errorf("no streams found in ffprobe output")
	}

	w := out.Streams[0].Width
	h := out.Streams[0].Height

	diff169 := w*9 - h*16
	if diff169 < 0 {
		diff169 = -diff169
	}

	diff916 := w*16 - h*9
	if diff916 < 0 {
		diff916 = -diff916
	}

	const tolerance = 10 // or some small number of “pixels”

	switch {
	case diff169 <= tolerance:
		return "16:9", nil
	case diff916 <= tolerance:
		return "9:16", nil
	default:
		return "other", nil
	}
}