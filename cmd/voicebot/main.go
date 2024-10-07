package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"golang.org/x/text/language"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/adrianliechti/llama-voicebot/pkg/play"
	"github.com/adrianliechti/llama-voicebot/pkg/record"
	"github.com/adrianliechti/llama-voicebot/pkg/say"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	client := openaiClient()

	chatmodel := "gpt-4o-mini"
	audiomodel := "whisper-1"
	speakmodel := "tts-1"

	system := "Your knowledge cutoff is 2023-10. You are a helpful, witty, and friendly AI. Act like a human, but remember that you aren't a human and that you can't do human things in the real world. Your voice and personality should be warm and engaging, with a lively and playful tone. If interacting in a non-English language, start by using the standard accent or dialect familiar to the user. Talk quickly. You should always call a function if you can. Answer as briefly and concisely as possible. Keep it short."

	messages := []openai.ChatCompletionMessageParamUnion{}

	if system != "" {
		messages = append(messages, openai.SystemMessage(system))
	}

	handleError := func(err error) {
		println("🚨 " + err.Error())
	}

	for ctx.Err() == nil {
		println("🙉 Listening...")

		data, err := record.Data(ctx, record.FormatWAV)

		if err != nil {
			handleError(err)
			continue
		}

		input, language, err := transcribe(ctx, client, audiomodel, data)

		if err != nil {
			handleError(err)
			continue
		}

		if input == "" {
			continue
		}

		fmt.Println("💬 " + input)

		messages = append(messages, openai.UserMessage(input))

		print("📣 ")

		stream := func(s string) {
			print(s)
		}

		output, err := complete(ctx, client, chatmodel, messages, stream)

		if err != nil {
			handleError(err)
			continue
		}

		println()

		messages = append(messages, openai.AssistantMessage(output))

		if speakmodel == "" {
			if err := say.Say(output, language); err != nil {
				handleError(err)
			}

			continue
		}

		name := uuid.New().String() + ".wav"
		path := filepath.Join(os.TempDir(), name)

		audio, err := speech(ctx, client, speakmodel, output, language)

		if err != nil {
			handleError(err)
			continue
		}

		defer os.Remove(path)

		if err := os.WriteFile(path, audio, 0644); err != nil {
			handleError(err)
			continue
		}

		if err := play.File(ctx, path); err != nil {
			handleError(err)
		}

		os.Remove(path)
	}
}

func openaiClient() *openai.Client {
	url := os.Getenv("OPENAI_API_BASE")

	if url == "" {
		url = "http://localhost:8080/v1"
	}

	url = strings.TrimRight(url, "/") + "/"

	options := []option.RequestOption{
		option.WithBaseURL(url),
	}

	return openai.NewClient(options...)
}

func transcribe(ctx context.Context, client *openai.Client, model string, data []byte) (string, string, error) {
	transcription, err := client.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		Model: openai.F(model),
		File:  openai.FileParam(bytes.NewReader(data), "file.wav", "audio/wav"),
	})

	if err != nil {
		return "", "", err
	}

	var metadata struct {
		Language string  `json:"language"`
		Duration float64 `json:"duration"`
	}

	json.Unmarshal([]byte(transcription.JSON.RawJSON()), &metadata)

	input := strings.TrimSpace(transcription.Text)
	language := parseLanguage(metadata.Language)

	return input, language, nil
}

func complete(ctx context.Context, client *openai.Client, model string, messages []openai.ChatCompletionMessageParamUnion, streamer func(string)) (string, error) {
	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:    openai.F(model),
		Messages: openai.F(messages),
	})

	var output string

	for stream.Next() {
		chunk := stream.Current()

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			output += content

			if streamer != nil {
				streamer(content)
			}
		}
	}

	if err := stream.Err(); err != nil {
		return "", err
	}

	return output, nil
}

func speech(ctx context.Context, client *openai.Client, model string, input, language string) ([]byte, error) {
	result, err := client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
		Model: openai.F(model),
		Input: openai.F(input),

		Voice:          openai.F(openai.AudioSpeechNewParamsVoiceAlloy),
		ResponseFormat: openai.F(openai.AudioSpeechNewParamsResponseFormatWAV),
	})

	if err != nil {
		return nil, err
	}

	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func parseLanguage(s string) string {
	languages := map[string]language.Tag{
		"afrikaans":   language.Afrikaans,
		"arabic":      language.Arabic,
		"armenian":    language.Armenian,
		"azerbaijani": language.Azerbaijani,
		//"belarusian": language.Belarusian,
		//"bosnian": language.Bosnian,
		"bulgarian": language.Bulgarian,
		"catalan":   language.Catalan,
		"chinese":   language.Chinese,
		"croatian":  language.Croatian,
		"czech":     language.Czech,
		"danish":    language.Danish,
		"dutch":     language.Dutch,
		"english":   language.English,
		"estonian":  language.Estonian,
		"finnish":   language.Finnish,
		"french":    language.French,
		//"galician": language.Galician,
		"german":     language.German,
		"greek":      language.Greek,
		"hebrew":     language.Hebrew,
		"hindi":      language.Hindi,
		"hungarian":  language.Hungarian,
		"icelandic":  language.Icelandic,
		"indonesian": language.Indonesian,
		"italian":    language.Italian,
		"japanese":   language.Japanese,
		"kannada":    language.Kannada,
		"kazakh":     language.Kazakh,
		"korean":     language.Korean,
		"latvian":    language.Latvian,
		"lithuanian": language.Lithuanian,
		"macedonian": language.Macedonian,
		"malay":      language.Malay,
		"marathi":    language.Marathi,
		//"maori": language.Maori,
		"nepali":     language.Nepali,
		"norwegian":  language.Norwegian,
		"persian":    language.Persian,
		"polish":     language.Polish,
		"portuguese": language.Portuguese,
		"romanian":   language.Romanian,
		"russian":    language.Russian,
		"serbian":    language.Serbian,
		"slovak":     language.Slovak,
		"slovenian":  language.Slovenian,
		"spanish":    language.Spanish,
		"swahili":    language.Swahili,
		"swedish":    language.Swedish,
		//"tagalog": language.Tagalog,
		"tamil":      language.Tamil,
		"thai":       language.Thai,
		"turkish":    language.Turkish,
		"ukrainian":  language.Ukrainian,
		"urdu":       language.Urdu,
		"vietnamese": language.Vietnamese,
		//"welsh": language.Welsh,
	}

	tag, found := languages[strings.ToLower(s)]

	if found {
		return tag.String()
	}

	return ""
}
