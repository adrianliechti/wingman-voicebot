package record

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/google/uuid"
)

func ffmpegData(ctx context.Context, format Format) ([]byte, error) {
	var args []string

	name := uuid.New().String() + format.Ext()
	path := filepath.Join(os.TempDir(), name)

	defer os.Remove(path)

	switch runtime.GOOS {
	case "darwin":
		args = []string{
			"-f", "avfoundation",
			"-i", ":0",
			"-af", "silencedetect=noise=-30dB:d=1",
			path,
		}
	case "windows":
		args = []string{
			"-f", "dshow",
			"-i", "audio=default",
			"-af", "silencedetect=noise=-30dB:d=1",
			path,
		}
	case "linux":
		args = []string{
			"-f", "alsa",
			"-i", "default",
			"-af", "silencedetect=noise=-30dB:d=1",
			path,
		}
	}

	if len(args) == 0 {
		return nil, errors.New("unsupported platform")
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024)
	silencePattern := regexp.MustCompile(`silence_start`)

	for {
		n, err := stderr.Read(buffer)

		if err != nil {
			break
		}

		output := string(buffer[:n])

		if silencePattern.MatchString(output) {
			break
		}
	}

	err = cmd.Process.Signal(os.Interrupt)

	if err != nil {
		err = cmd.Process.Kill()
	}

	if err != nil {
		fmt.Println("Error killing FFmpeg process:", err)
		return nil, err
	}

	cmd.Process.Wait()

	data, err := os.ReadFile(path)

	if err != nil {
		fmt.Println("error reading file:", err)
		return nil, err
	}

	return data, nil
}
