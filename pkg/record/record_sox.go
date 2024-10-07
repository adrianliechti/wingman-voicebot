package record

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

func soxData(ctx context.Context, format Format) ([]byte, error) {
	name := uuid.New().String() + format.Ext()
	path := filepath.Join(os.TempDir(), name)

	defer os.Remove(path)

	args := []string{
		"-d",
		path,
		"silence",
		"1", "0.0", "1%",
		"1", "1.5", "1%",
	}

	cmd := exec.CommandContext(ctx, "sox", args...)

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	return data, nil
}
