package play

import (
	"context"
	"os/exec"
)

func ffmpegFile(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "ffplay", "-autoexit", "-nodisp", path)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
