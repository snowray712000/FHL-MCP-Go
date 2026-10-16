package bible

import (
	"context"
	"encoding/json"
	"strings"
)

// AudioVersion is one audio Bible recording set.
type AudioVersion struct {
	Code string `json:"code"`
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// audioVersions is the static table of the Python reference (AUDIO_VERSIONS), in id order (id 16 does not exist).
var audioVersions = []AudioVersion{
	{"unv", 0, "和合本"},
	{"taiwanese", 1, "台語"},
	{"hakka", 2, "客家話"},
	{"cantonese", 3, "廣東話"},
	{"tcv", 4, "現代中文譯本"},
	{"taiwanese_nt", 5, "台語新約"},
	{"red_bible", 6, "紅皮聖經"},
	{"hebrew", 7, "希伯來文"},
	{"fuzhou", 8, "福州話"},
	{"greek", 9, "希臘文"},
	{"spring_taiwanese", 10, "Spring台語"},
	{"spring_unv", 11, "Spring和合本"},
	{"netbible_chinese", 12, "NetBible中文版"},
	{"pct", 13, "全民台語聖經"},
	{"tsou", 14, "鄒語"},
	{"taiwanese_south", 15, "台語南部腔"},
	{"modern_taiwanese", 17, "現代台語譯本"},
	{"modern_hakka", 18, "現代客語譯本"},
	{"tao", 19, "達悟語"},
}

// AudioVersionList is the list_audio_versions result.
type AudioVersionList struct {
	TotalCount int            `json:"total_count"`
	Versions   []AudioVersion `json:"versions"`
}

// ListAudioVersions lists the audio versions (static data, no upstream request).
func ListAudioVersions() *AudioVersionList {
	return &AudioVersionList{TotalCount: len(audioVersions), Versions: audioVersions}
}

// AudioFiles holds the file links; a link that is empty upstream is left out.
type AudioFiles struct {
	MP3 json.RawMessage `json:"mp3,omitempty"`
	OGG json.RawMessage `json:"ogg,omitempty"`
}

// AudioNav is one neighbouring chapter. The values are passed through as the upstream sent them
// (strings such as "43"), like the Python reference.
type AudioNav struct {
	BookID  json.RawMessage `json:"book_id"`
	Book    json.RawMessage `json:"book"`
	Chapter json.RawMessage `json:"chapter"`
}

// AudioNavigation is present only when at least one neighbour is known.
type AudioNavigation struct {
	Prev *AudioNav `json:"prev,omitempty"`
	Next *AudioNav `json:"next,omitempty"`
}

// AudioResult is the get_audio_bible result (field order is part of the contract).
type AudioResult struct {
	Version     string           `json:"version"`
	VersionName json.RawMessage  `json:"version_name"`
	Book        json.RawMessage  `json:"book"`
	BookEng     json.RawMessage  `json:"book_eng"`
	Chapter     json.RawMessage  `json:"chapter"`
	AudioFiles  AudioFiles       `json:"audio_files"`
	Navigation  *AudioNavigation `json:"navigation,omitempty"`
}

// AudioWithText is the get_audio_chapter_with_text result.
type AudioWithText struct {
	Audio *AudioResult `json:"audio"`
	Text  *VerseResult `json:"text"`
}

func audioVersionID(code string) (int, error) {
	codes := make([]string, 0, len(audioVersions))
	for _, v := range audioVersions {
		if v.Code == code {
			return v.ID, nil
		}
		codes = append(codes, v.Code)
	}
	return 0, &InvalidParameterError{Name: "audio_version", Value: code, Reason: "Must be one of: " + strings.Join(codes, ", ")}
}

// jsonOrNull keeps a raw upstream value, or null when the field was absent (Python `.get`).
func jsonOrNull(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage("null")
	}
	return raw
}

func present(raw json.RawMessage) bool { return truthy(rawOr(raw, nil)) }

// GetAudio returns the audio links of one chapter (au.php). Like the Python reference it does not
// check that the chapter exists: the upstream answers with links for any chapter number.
// simplified (only get_audio_chapter_with_text has the parameter) asks for simplified book names (D4).
func (s *Service) GetAudio(ctx context.Context, book string, chapter int, audioVersion string, simplified bool) (*AudioResult, error) {
	vid, err := audioVersionID(audioVersion)
	if err != nil {
		return nil, err
	}
	bid, ok := ResolveBookID(book)
	if !ok {
		return nil, &InvalidParameterError{Name: "book", Value: book, Reason: "Book not found. Please check the spelling or use a valid book code."}
	}
	resp, err := s.API.AU(ctx, vid, bid, chapter, simplified)
	if err != nil {
		return nil, err
	}
	out := &AudioResult{
		Version:     audioVersion,
		VersionName: jsonOrNull(resp.Name),
		Book:        jsonOrNull(resp.Chinesef),
		BookEng:     jsonOrNull(resp.Engf),
		Chapter:     jsonOrNull(resp.Chap),
	}
	if present(resp.MP3) {
		out.AudioFiles.MP3 = resp.MP3
	}
	if present(resp.OGG) {
		out.AudioFiles.OGG = resp.OGG
	}
	var nav AudioNavigation
	if present(resp.PBID) && present(resp.PChap) {
		nav.Prev = &AudioNav{BookID: resp.PBID, Book: jsonOrNull(resp.PChinesef), Chapter: resp.PChap}
	}
	if present(resp.NBID) && present(resp.NChap) {
		nav.Next = &AudioNav{BookID: resp.NBID, Book: jsonOrNull(resp.NChinesef), Chapter: resp.NChap}
	}
	if nav.Prev != nil || nav.Next != nil {
		out.Navigation = &nav
	}
	return out, nil
}

// GetAudioWithText returns the audio links and the chapter text together.
func (s *Service) GetAudioWithText(ctx context.Context, book string, chapter int, audioVersion, textVersion string, simplified bool) (*AudioWithText, error) {
	audio, err := s.GetAudio(ctx, book, chapter, audioVersion, simplified)
	if err != nil {
		return nil, err
	}
	text, err := s.GetChapter(ctx, book, chapter, textVersion, simplified)
	if err != nil {
		return nil, err
	}
	return &AudioWithText{Audio: audio, Text: text}, nil
}
