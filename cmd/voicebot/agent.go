package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrianliechti/go-cli"
	"github.com/adrianliechti/wingman-cli/pkg/tool"
	"github.com/adrianliechti/wingman-cli/pkg/util"
	"github.com/adrianliechti/wingman-voicebot/pkg/play"
	"github.com/adrianliechti/wingman-voicebot/pkg/record"
	"github.com/google/uuid"

	"github.com/adrianliechti/wingman-cli/pkg/markdown"
	wingman "github.com/adrianliechti/wingman/pkg/client"
)

func Run(ctx context.Context, client *wingman.Client, model string, tools []tool.Tool) error {
	input := wingman.CompletionRequest{
		Model: model,

		CompleteOptions: wingman.CompleteOptions{
			Tools: util.ConvertTools(tools),
		},
	}

	input.Messages = append(input.Messages, wingman.SystemMessage("Your knowledge cutoff is 2024-06. You are a helpful, witty, and friendly AI. Act like a human, but remember that you aren't a human and that you can't do human things in the real world. Your voice and personality should be warm and engaging, with a lively and playful tone. If interacting in a non-English language, start by using the standard accent or dialect familiar to the user. Talk quickly. You should always call a function if you can. Answer as briefly and concisely as possible. Keep it short."))

	for {
		if ctx.Err() != nil {
			break
		}

		println("🙉 Listening...")

		data, err := record.Data(ctx, record.FormatWAV)

		if err != nil {
			println("🚨 " + err.Error())
			break
		}

		transcription, err := client.Transcriptions.New(ctx, wingman.TranscribeRequest{
			Name:   "audio.wav",
			Reader: bytes.NewReader(data),
		})

		if err != nil {
			println("🚨 " + err.Error())
			break
		}

		if transcription.Text == "" {
			continue
		}

		fmt.Println("💬 " + transcription.Text)

		input.Messages = append(input.Messages, wingman.UserMessage(transcription.Text))

		var message *wingman.Message

		for {
			var completion *wingman.Completion

			fn := func() error {
				completion, err = client.Completions.New(ctx, input)
				return err
			}

			if err := cli.Run("Thinking...", fn); err != nil {
				return err
			}

			message = completion.Message
			input.Messages = append(input.Messages, *message)

			calls := message.ToolCalls()

			if len(calls) == 0 {
				break
			}

			for _, call := range calls {
				content, err := handleToolCall(ctx, tools, call)

				if err != nil {
					content = err.Error()
				}

				input.Messages = append(input.Messages, wingman.ToolMessage(call.ID, content))
			}
		}

		if message == nil {
			return nil
		}

		markdown.Render(os.Stdout, message.Text())

		synthesis, err := client.Syntheses.New(ctx, wingman.SynthesizeRequest{
			Input: message.Text(),
		})

		if err != nil {
			println("🚨 " + err.Error())
			continue
		}

		audio, err := io.ReadAll(synthesis.Reader)

		name := uuid.New().String() + ".wav"
		path := filepath.Join(os.TempDir(), name)

		defer os.Remove(path)

		if err := os.WriteFile(path, audio, 0644); err != nil {
			println("🚨 " + err.Error())
			continue
		}

		play.File(ctx, path)
		os.Remove(path)
	}

	return nil
}

func handleToolCall(ctx context.Context, tools []tool.Tool, call wingman.ToolCall) (string, error) {
	var handler tool.ExecuteFn

	for _, t := range tools {
		if !strings.EqualFold(t.Name, call.Name) {
			continue
		}

		handler = t.Execute
	}

	if handler == nil {
		return "", errors.New("Unknown tool: " + call.Name)
	}

	var args map[string]any
	json.Unmarshal([]byte(call.Arguments), &args)

	result, err := handler(ctx, args)

	if err != nil {
		return "", err
	}

	var content string

	switch v := result.(type) {
	case string:
		content = v
	default:
		data, _ := json.Marshal(v)
		content = string(data)
	}

	return content, nil
}
