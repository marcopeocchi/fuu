package utils

import (
	"os/exec"
)

func GetCmd(input, output, format string) (*exec.Cmd, string) {
	if format == "" {
		format = "webp"
	}

	if IsImagePath(input) {
		return exec.Command(
			"convert", input,
			"-geometry", "x450",
			"-format", format,
			"-quality", "80",
			output,
		), "imagemagick"
	}
	return exec.Command(
		"ffmpeg",
		"-i", input,
		"-ss", "00:00:01.000",
		"-vframes", "1",
		"-filter:v", "scale=-1:450",
		"-f", format,
		output,
	), "ffmpeg"

}
