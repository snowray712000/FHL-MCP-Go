package tools

import (
	"context"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
)

// registerAudio wires the audio Bible tools.
func (r *Registry) registerAudio(svc *bible.Service) {
	r.handlers["get_audio_bible"] = func(ctx context.Context, a args) (string, error) {
		book, err := a.requiredString("book")
		if err != nil {
			return "", err
		}
		chapter, err := a.requiredInt("chapter")
		if err != nil {
			return "", err
		}
		version, err := a.optString("audio_version", "unv")
		if err != nil {
			return "", err
		}
		// The contract has no use_simplified for this tool, so the request is always traditional.
		res, err := svc.GetAudio(ctx, book, chapter, version, false)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}

	r.handlers["get_audio_chapter_with_text"] = func(ctx context.Context, a args) (string, error) {
		book, err := a.requiredString("book")
		if err != nil {
			return "", err
		}
		chapter, err := a.requiredInt("chapter")
		if err != nil {
			return "", err
		}
		audioVersion, err := a.optString("audio_version", "unv")
		if err != nil {
			return "", err
		}
		textVersion, err := a.optString("text_version", "unv")
		if err != nil {
			return "", err
		}
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		res, err := svc.GetAudioWithText(ctx, book, chapter, audioVersion, textVersion, simplified)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}

	r.handlers["list_audio_versions"] = func(context.Context, args) (string, error) {
		return bible.MarshalIndent2(bible.ListAudioVersions())
	}
}
