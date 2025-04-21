package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/adrianliechti/wingman-cli/pkg/util"
	wingman "github.com/adrianliechti/wingman/pkg/client"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	tools, _ := util.ParseMCP()

	url := os.Getenv("WINGMAN_URL")
	model := os.Getenv("WINGMAN_MODEL")

	if url == "" {
		url = "http://localhost:8080"
	}

	var options []wingman.RequestOption

	if token := os.Getenv("WINGMAN_TOKEN"); token != "" {
		options = append(options, wingman.WithToken(token))
	}

	client := wingman.New(url, options...)

	if err := Run(ctx, client, model, tools); err != nil {
		panic(err)
	}
}
