package play

import (
	"context"
	"os/exec"
)

func soxFile(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "play", path)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
