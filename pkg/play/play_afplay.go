package play

import (
	"context"
	"os/exec"
)

func afplayFile(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "afplay", path)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
