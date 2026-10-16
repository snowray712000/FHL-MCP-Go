package bible

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// FootnoteAPI is the part of the upstream client the footnote tool needs (rt.php).
type FootnoteAPI interface {
	RT(ctx context.Context, bid, id int, simplified bool) (*fhlapi.RTResponse, error)
}

// Footnote is the get_bible_footnote result when the footnote exists. Only the TCV version has footnotes.
type Footnote struct {
	Status      string          `json:"status"`
	QueryType   string          `json:"query_type"`
	Version     string          `json:"version"`
	VersionName string          `json:"version_name"`
	BookID      int             `json:"book_id"`
	BookName    string          `json:"book_name"`
	FootnoteID  json.RawMessage `json:"footnote_id"`
	Text        string          `json:"text"`
}

// FootnoteNotFound is the get_bible_footnote result when there is no such footnote (a normal, successful answer).
type FootnoteNotFound struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// GetFootnote looks up one TCV footnote. The result is a *Footnote, or a *FootnoteNotFound when the
// upstream has none. Like the Python reference it does not check the book number: an unknown book is "not found".
func (s *Service) GetFootnote(ctx context.Context, bookID, footnoteID int, simplified bool) (any, error) {
	resp, err := s.API.RT(ctx, bookID, footnoteID, simplified)
	if err != nil {
		return nil, err
	}
	if resp.RecordCount <= 0 || len(resp.Record) == 0 {
		return &FootnoteNotFound{Status: "not_found", Message: fmt.Sprintf("找不到書卷 %d 的註腳 #%d", bookID, footnoteID)}, nil
	}
	rec := resp.Record[0]
	id := rec.ID
	if len(id) == 0 {
		id = json.RawMessage(fmt.Sprint(footnoteID))
	}
	return &Footnote{
		Status: "success", QueryType: "bible_footnote", Version: "TCV", VersionName: "現代中文譯本",
		BookID: bookID, BookName: strOr(resp.Engs, ""), FootnoteID: id, Text: strOr(rec.Text, ""),
	}, nil
}
