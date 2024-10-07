//go:build darwin

package say

import (
	"os/exec"
	"regexp"
	"strings"
)

var (
	voices []Voice

	preferred = []string{
		// English
		"Samantha",
		"Daniel",
		"Moira",
		"Tessa",

		// German
		"Anna",
		"Markus",
		"Petra",

		// French
		"Thomas",
		"Audrey",
		"Amélie",

		// Spanish
		"Monica",
		"Jorge",
		"Juan",

		// Italian
		"Alice",
		"Federico",

		// Chinese
		"Ting-Ting",

		// Japanese
		"Kyoko",
		"Otoya",

		// Arabic
		"Maged",

		// Dutch
		"Xander",

		// Swedish
		"Alva",

		// Korean
		"Yuna",

		// Portuguese
		"Luciana",
	}
)

func Say(text, language string) error {
	if len(voices) == 0 {
		var err error
		voices, err = listVoices()

		if err != nil {
			return err
		}
	}

	voice := selectVoice(voices, language)

	var args []string

	if voice.Name != "" {
		args = append(args, "-v", voice.Name)
	}

	args = append(args, text)

	return exec.Command("say", args...).Run()
}

type Voice struct {
	Type VoiceType

	Language string

	Name        string
	Description string
}

type VoiceType string

const (
	VoiceTypeStandard VoiceType = ""
	VoiceTypeEnhanced VoiceType = "enhanced"
	VoiceTypePremium  VoiceType = "premium"
)

func selectVoice(voices []Voice, language string) Voice {
	if language == "" {
		language = "en_US"
	}

	for _, l := range []string{language, strings.Split(language, "_")[0] + "_"} {
		for _, p := range append(preferred, "") {
			for _, t := range []VoiceType{VoiceTypePremium, VoiceTypeEnhanced, VoiceTypeStandard} {
				for _, v := range voices {
					if v.Type != t {
						continue
					}

					if !strings.HasPrefix(strings.ToLower(v.Language), strings.ToLower(l)) {
						continue
					}

					if !strings.HasPrefix(strings.ToLower(v.Name), strings.ToLower(p)) {
						continue
					}

					return v
				}
			}
		}
	}

	return Voice{}
}

func listVoices() ([]Voice, error) {
	regex := regexp.MustCompile(`(?P<voice>.+)\s+(?P<language>[a-z]+_[A-Z]+)\s+#\s*(?P<comment>.+)`)

	list, err := exec.Command("say", "-v", "?").Output()

	if err != nil {
		return nil, err
	}

	var result []Voice

	lines := strings.Split(string(list), "\n")

	for _, line := range lines {
		matches := regex.FindStringSubmatch(line)

		if len(matches) != 4 {
			continue
		}

		voice := Voice{
			Type: VoiceTypeStandard,
			Name: strings.TrimSpace(matches[1]),

			Language:    strings.TrimSpace(matches[2]),
			Description: strings.TrimSpace(matches[3]),
		}

		if strings.Contains(voice.Name, "(Enhanced)") {
			voice.Type = VoiceTypeEnhanced
		}

		if strings.Contains(voice.Name, "(Premium)") {
			voice.Type = VoiceTypePremium
		}

		result = append(result, voice)
	}

	return result, nil
}
