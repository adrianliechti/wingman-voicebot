package record

import (
	"context"
	"errors"
	"os/exec"
)

type Format string

const (
	FormatWAV Format = "wav"
)

func (f Format) Ext() string {
	switch f {
	case FormatWAV:
		return ".wav"
	}
	return ""
}

func Data(ctx context.Context, format Format) ([]byte, error) {
	if _, err := exec.LookPath("sox"); err == nil {
		return soxData(ctx, format)
	}

	if _, err := exec.LookPath("ffmpeg"); err == nil {
		return ffmpegData(ctx, format)
	}

	return nil, errors.New("neither FFmpeg nor SoX are installed")
}
