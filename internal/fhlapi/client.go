// Package fhlapi is the single, injectable client for the FHL (信望愛) upstream APIs.
package fhlapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the upstream Bible API root measured in M0 (every qb/ab/... endpoint lives here).
	DefaultBaseURL = "https://bible.fhl.net/api"
	// DefaultUserAgent identifies this server upstream.
	DefaultUserAgent = "FHL-MCP-Go"
	// maxBodyBytes bounds how much of an upstream response is read.
	maxBodyBytes = 8 << 20
)

// Client talks to the upstream API. HTTP and BaseURL are replaceable so tests can point it at a local mock.
type Client struct {
	BaseURL   string
	HTTP      *http.Client
	UserAgent string
}

// NewClient returns a Client with production defaults (30 s timeout, like the Python reference).
func NewClient() *Client {
	return &Client{
		BaseURL:   DefaultBaseURL,
		HTTP:      &http.Client{Timeout: 30 * time.Second},
		UserAgent: DefaultUserAgent,
	}
}

// Ref points at one verse (used for prev/next navigation).
type Ref struct {
	BID      int    `json:"bid"`
	Chineses string `json:"chineses"`
	Engs     string `json:"engs"`
	Chap     int    `json:"chap"`
	Sec      int    `json:"sec"`
}

// Record is one verse returned by qb.php.
type Record struct {
	BID       int    `json:"bid"`
	Engs      string `json:"engs"`
	Chineses  string `json:"chineses"`
	Chap      int    `json:"chap"`
	Sec       int    `json:"sec"`
	BibleText string `json:"bible_text"`
}

// QBResponse is the qb.php (verse query) response.
type QBResponse struct {
	Status      string   `json:"status"`
	RecordCount int      `json:"record_count"`
	VName       string   `json:"v_name"`
	Version     string   `json:"version"`
	Record      []Record `json:"record"`
	Prev        *Ref     `json:"prev"`
	Next        *Ref     `json:"next"`
}

// QBRequest are the qb.php query parameters.
type QBRequest struct {
	BID        int
	Chapter    int
	Verse      string // "sec"; empty means the whole chapter
	Version    string
	Strong     bool
	Simplified bool // gb=1 (D4: really sent, unlike the Python reference)
}

// Error kinds, for callers and tests.
var (
	ErrNetwork  = errors.New("upstream network error")
	ErrHTTP     = errors.New("upstream HTTP error")
	ErrParse    = errors.New("upstream response is not valid JSON")
	ErrUpstream = errors.New("upstream reported a failure")
)

// Error describes an upstream failure.
type Error struct {
	Kind    error // one of the Err* values
	Status  int   // HTTP status, when relevant
	Message string
}

func (e *Error) Error() string { return fmt.Sprintf("%v: %s", e.Kind, e.Message) }
func (e *Error) Unwrap() error { return e.Kind }

// QB queries qb.php.
func (c *Client) QB(ctx context.Context, r QBRequest) (*QBResponse, error) {
	q := url.Values{}
	q.Set("bid", strconv.Itoa(r.BID))
	q.Set("chap", strconv.Itoa(r.Chapter))
	q.Set("version", r.Version)
	q.Set("strong", boolDigit(r.Strong))
	if r.Verse != "" {
		q.Set("sec", r.Verse)
	}
	q.Set("gb", boolDigit(r.Simplified))

	var out QBResponse
	if err := c.getJSON(ctx, "qb.php", q, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

// ABRecord is one Bible version returned by ab.php. Optional fields missing upstream stay zero.
type ABRecord struct {
	Book        string `json:"book"`
	CName       string `json:"cname"`
	Proc        int    `json:"proc"`
	Strong      int    `json:"strong"`
	NTOnly      int    `json:"ntonly"`
	OTOnly      int    `json:"otonly"`
	CanDownload int    `json:"candownload"`
	Version     string `json:"version"`
}

// ABResponse is the ab.php (version list) response.
type ABResponse struct {
	Status      string     `json:"status"`
	RecordCount int        `json:"record_count"`
	Record      []ABRecord `json:"record"`
}

// AB queries ab.php for the list of Bible versions. simplified sends gb=1 (D4).
func (c *Client) AB(ctx context.Context, simplified bool) (*ABResponse, error) {
	q := url.Values{}
	q.Set("gb", boolDigit(simplified))
	var out ABResponse
	if err := c.getJSON(ctx, "ab.php", q, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

// SERequest are the se.php (search) query parameters. count_only / index_only are always 0:
// no tool exposes them.
type SERequest struct {
	Query      string
	Version    string
	Orig       int // 0 keyword, 1 Greek Strong's number, 2 Hebrew Strong's number
	Range      int // 0 whole Bible, 1 NT, 2 OT, 3 book range (RangeBID..RangeEID)
	Offset     int
	Limit      int
	RangeBID   int
	RangeEID   int
	Simplified bool
}

// SERecord is one matching verse returned by se.php.
type SERecord struct {
	Chineses  string `json:"chineses"`
	Engs      string `json:"engs"`
	Chap      int    `json:"chap"`
	Sec       int    `json:"sec"`
	BibleText string `json:"bible_text"`
}

// SEResponse is the se.php response. RecordCount is what the upstream reports (the
// Python reference publishes it as total_count unchanged).
type SEResponse struct {
	Status      string     `json:"status"`
	RecordCount int        `json:"record_count"`
	Record      []SERecord `json:"record"`
}

// SE queries se.php. A status other than "success" (for example "Fail: Empty Query") is an ErrUpstream.
func (c *Client) SE(ctx context.Context, r SERequest) (*SEResponse, error) {
	q := url.Values{}
	q.Set("VERSION", r.Version)
	q.Set("orig", strconv.Itoa(r.Orig))
	q.Set("q", r.Query)
	q.Set("RANGE", strconv.Itoa(r.Range))
	q.Set("offset", strconv.Itoa(r.Offset))
	q.Set("count_only", "0")
	q.Set("index_only", "0")
	q.Set("limit", strconv.Itoa(r.Limit))
	if r.Range == 3 {
		q.Set("range_bid", strconv.Itoa(r.RangeBID))
		q.Set("range_eid", strconv.Itoa(r.RangeEID))
	}
	q.Set("gb", boolDigit(r.Simplified))
	var out SEResponse
	if err := c.getJSON(ctx, "se.php", q, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

// SDRelated is one related word of a Strong's dictionary entry. Values stay raw JSON
// because the Python reference passes them through with their upstream types.
type SDRelated struct {
	Word json.RawMessage `json:"word"`
	CSN  json.RawMessage `json:"csn"`
	CCnt json.RawMessage `json:"ccnt"`
	CExp json.RawMessage `json:"cexp"`
}

// SDRecord is one Strong's dictionary entry. Optional fields keep their raw JSON so an
// absent field and an explicit null can be told apart (the Python reference does).
type SDRecord struct {
	SN       string          `json:"sn"`
	Orig     json.RawMessage `json:"orig"`
	DicText  json.RawMessage `json:"dic_text"`
	EDicText json.RawMessage `json:"edic_text"`
	Same     []SDRelated     `json:"same"`
}

// SDResponse is the sd.php response.
type SDResponse struct {
	Status      string     `json:"status"`
	RecordCount int        `json:"record_count"`
	Record      []SDRecord `json:"record"`
}

// SD queries sd.php (Strong's dictionary). nt selects Greek (N=0), otherwise Hebrew (N=1).
func (c *Client) SD(ctx context.Context, nt bool, number int, simplified bool) (*SDResponse, error) {
	q := url.Values{}
	if nt {
		q.Set("N", "0")
	} else {
		q.Set("N", "1")
	}
	q.Set("k", strconv.Itoa(number))
	q.Set("gb", boolDigit(simplified))
	var out SDResponse
	if err := c.getJSON(ctx, "sd.php", q, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

// QPRecord is one row of a qp.php (original-language word analysis) response. Record 0 is the
// verse summary (wid 0); the others are words. Text fields stay raw JSON so an absent field
// and an explicit null can be told apart, which the Python reference's output depends on.
type QPRecord struct {
	WID      json.RawMessage `json:"wid"`
	Word     json.RawMessage `json:"word"`
	Exp      json.RawMessage `json:"exp"`
	SN       json.RawMessage `json:"sn"`
	Pro      json.RawMessage `json:"pro"`
	WForm    json.RawMessage `json:"wform"`
	Orig     json.RawMessage `json:"orig"`
	Remark   json.RawMessage `json:"remark"`
	Chinesef json.RawMessage `json:"chinesef"`
}

// QPResponse is the qp.php response. N is 0 for the New Testament, 1 for the Old Testament.
type QPResponse struct {
	Status      string     `json:"status"`
	RecordCount int        `json:"record_count"`
	N           int        `json:"N"`
	Record      []QPRecord `json:"record"`
}

// QP queries qp.php for one verse. (A verse that does not exist makes the upstream return
// malformed JSON; that surfaces as ErrParse.)
func (c *Client) QP(ctx context.Context, bid, chapter, verse int, simplified bool) (*QPResponse, error) {
	q := url.Values{}
	q.Set("bid", strconv.Itoa(bid))
	q.Set("chap", strconv.Itoa(chapter))
	q.Set("sec", strconv.Itoa(verse))
	q.Set("gb", boolDigit(simplified))
	var out QPResponse
	if err := c.getJSON(ctx, "qp.php", q, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

// SCListRecord is one commentary book returned by sc.php?validbook=1.
type SCListRecord struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SCListResponse is the commentary-book list response.
type SCListResponse struct {
	Status      string         `json:"status"`
	RecordCount int            `json:"record_count"`
	Record      []SCListRecord `json:"record"`
}

// SCList queries sc.php?validbook=1 for the available commentary books.
func (c *Client) SCList(ctx context.Context, simplified bool) (*SCListResponse, error) {
	q := url.Values{}
	q.Set("validbook", "1")
	q.Set("gb", boolDigit(simplified))
	var out SCListResponse
	if err := c.getJSON(ctx, "sc.php", q, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

// SCRecord is one commentary entry of a verse. Title and ComText may be null upstream.
// (The response also carries top-level prev/next verse refs; no tool output uses them.)
type SCRecord struct {
	Title    *string `json:"title"`
	BookName string  `json:"book_name"`
	ComText  *string `json:"com_text"`
}

// SCResponse is the sc.php commentary response. A verse without commentary has no "record" key.
type SCResponse struct {
	Status      string     `json:"status"`
	RecordCount int        `json:"record_count"`
	Record      []SCRecord `json:"record"`
}

// SC queries sc.php for the commentaries of one verse. A nil verse leaves "sec" out; the upstream then answers with zero records.
func (c *Client) SC(ctx context.Context, bid, chapter int, verse *int, simplified bool) (*SCResponse, error) {
	q := url.Values{}
	q.Set("bid", strconv.Itoa(bid))
	q.Set("chap", strconv.Itoa(chapter))
	if verse != nil {
		q.Set("sec", strconv.Itoa(*verse))
	}
	q.Set("gb", boolDigit(simplified))
	var out SCResponse
	if err := c.getJSON(ctx, "sc.php", q, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

// ErrTooManyRecords is an upstream search that matched more records than it is willing to return
// (ssc.php answers status "Fail:record count > 500").
var ErrTooManyRecords = errors.New("upstream result has more than 500 records")

// SSCRecord is one hit of a commentary search. Chinesef keeps the upstream's space padding.
type SSCRecord struct {
	Title    *string `json:"title"`
	Tag      int     `json:"tag"`
	BookName string  `json:"book_name"`
	Chinesef string  `json:"chinesef"`
	BID      int     `json:"bid"`
	Engs     string  `json:"engs"`
	BChap    int     `json:"bchap"`
	BSec     int     `json:"bsec"`
	EChap    int     `json:"echap"`
	ESec     int     `json:"esec"`
}

// SSCResponse is the ssc.php response. No hits means no "record" key.
type SSCResponse struct {
	Status      string      `json:"status"`
	RecordCount int         `json:"record_count"`
	Record      []SSCRecord `json:"record"`
}

// SSC searches the commentaries (ssc.php). Too common a keyword yields ErrTooManyRecords.
func (c *Client) SSC(ctx context.Context, key string, simplified bool) (*SSCResponse, error) {
	q := url.Values{}
	q.Set("key", key)
	q.Set("gb", boolDigit(simplified))
	var out SSCResponse
	if err := c.getJSON(ctx, "ssc.php", q, &out); err != nil {
		return nil, err
	}
	if out.Status == "Fail:record count > 500" {
		return nil, &Error{Kind: ErrTooManyRecords, Message: "status " + strconv.Quote(out.Status)}
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

// STRecord is one topic entry of st.php. Book is the source: 0 Torrey (en), 1 Nave's (en), 2 Torrey (zh), 3 Nave's (zh).
type STRecord struct {
	Book  int     `json:"book"`
	ID    string  `json:"id"`
	Topic *string `json:"topic"`
	Text  *string `json:"text"`
}

// STResponse is the st.php response. RecordCount stays a pointer: Python falls back to the
// number of returned records when it is absent.
type STResponse struct {
	Status      string     `json:"status"`
	RecordCount *int       `json:"record_count"`
	Record      []STRecord `json:"record"`
}

// ST queries st.php (topical study). source is the upstream N value: 0..3 for one source, 4 for all.
func (c *Client) ST(ctx context.Context, source int, keyword string, countOnly, simplified bool) (*STResponse, error) {
	q := url.Values{}
	q.Set("N", strconv.Itoa(source))
	q.Set("count_only", boolDigit(countOnly))
	q.Set("keyword", keyword)
	q.Set("gb", boolDigit(simplified))
	var out STResponse
	if err := c.getJSON(ctx, "st.php", q, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, q url.Values, dst any) error {
	u := strings.TrimRight(c.BaseURL, "/") + "/" + endpoint + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return &Error{Kind: ErrNetwork, Message: err.Error()}
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return &Error{Kind: ErrNetwork, Message: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return &Error{Kind: ErrNetwork, Message: err.Error()}
	}
	if resp.StatusCode >= 400 {
		return &Error{Kind: ErrHTTP, Status: resp.StatusCode, Message: fmt.Sprintf("%s returned HTTP %d", endpoint, resp.StatusCode)}
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return &Error{Kind: ErrParse, Message: fmt.Sprintf("%s: %v", endpoint, err)}
	}
	return nil
}

func boolDigit(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
