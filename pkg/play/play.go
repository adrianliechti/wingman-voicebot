package play

import (
	"context"
	"errors"
	"os/exec"
)

func File(ctx context.Context, path string) error {
	if _, err := exec.LookPath("afplay"); err == nil {
		return afplayFile(ctx, path)
	}

	if _, err := exec.LookPath("play"); err == nil {
		return soxFile(ctx, path)
	}

	if _, err := exec.LookPath("ffplay"); err == nil {
		return ffmpegFile(ctx, path)
	}

	return errors.New("no supported player found")
}
