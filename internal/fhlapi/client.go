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

	"github.com/snowray712000/FHL-MCP-Go/internal/cache"
	"github.com/snowray712000/FHL-MCP-Go/internal/hanzi"
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
	// Cache, when set, stores successful responses on disk (see endpointTTL) and merges concurrent
	// identical requests. nil disables caching.
	Cache *cache.Cache
	// ArticlesBaseURL is the server of the article search (a different site from BaseURL).
	// Empty means DefaultArticlesBaseURL.
	ArticlesBaseURL string
}

// endpointTTL lists the endpoints whose successful responses are cached, with their lifetime
// (0 = never expires, only the disk cap removes it). Verses, original-language analysis,
// commentary, topics and the version list change rarely (7 days); searches may follow new
// data sooner (1 day); the Strong's dictionary is fixed. Python cached only qb/se/qp/sd/ab; the
// commentary and topic endpoints follow the plan (commentary ~7 days).
var endpointTTL = map[string]time.Duration{
	"qb.php":    7 * 24 * time.Hour,
	"qp.php":    7 * 24 * time.Hour,
	"sc.php":    7 * 24 * time.Hour,
	"st.php":    7 * 24 * time.Hour,
	"au.php":    7 * 24 * time.Hour,
	"ab.php":    7 * 24 * time.Hour,
	"uiabv.php": 7 * 24 * time.Hour,
	"se.php":    24 * time.Hour,
	"ssc.php":   24 * time.Hour,
	"sd.php":    0,
	// TWCB dictionaries (浸宣): fixed text.
	"sbdag.php":     0,
	"stwcbhdic.php": 0,
	// Apocrypha and Apostolic Fathers (qsub/qaf = text, sesub/seaf = search).
	"qsub.php":   7 * 24 * time.Hour,
	"qaf.php":    7 * 24 * time.Hour,
	"sesub.php":  24 * time.Hour,
	"seaf.php":   24 * time.Hour,
	"rt.php":     7 * 24 * time.Hour,
	"rt-nui.php": 7 * 24 * time.Hour,
	// Articles (www.fhl.net/api/json.php): the site updates weekly.
	"json.php": 24 * time.Hour,
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

// UIABV queries uiabv.php (bible.fhl.net/json), the version list FHL's own web UI uses. It is a superset of
// ab.php: it also lists versions ab.php omits (esv, lxx, cbol, tru, thv12h ...) and fills candownload / version.
// An empty list is an ErrUpstream error so callers can fall back to AB.
func (c *Client) UIABV(ctx context.Context, simplified bool) (*ABResponse, error) {
	q := url.Values{}
	q.Set("gb", boolDigit(simplified))
	var out ABResponse
	if err := c.getJSONAt(ctx, c.jsonBase(), "uiabv.php", q, isSuccess, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" || len(out.Record) == 0 || out.Record[0].Book == "" {
		return nil, &Error{Kind: ErrUpstream, Message: "uiabv.php: no version list"}
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
	q.Set("q", hanzi.ForMode(r.Query, r.Simplified))
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

// TWCBRecord is one entry of the TWCB dictionaries (sbdag.php for Greek, stwcbhdic.php for Hebrew).
type TWCBRecord struct {
	SN      string `json:"sn"`
	DicText string `json:"dic_text"`
}

// TWCBResponse is the sbdag.php / stwcbhdic.php response.
type TWCBResponse struct {
	Status      string       `json:"status"`
	RecordCount int          `json:"record_count"`
	Record      []TWCBRecord `json:"record"`
}

// TWCB queries the 浸宣 (TWCB) dictionary: sbdag.php for the Greek NT (nt), stwcbhdic.php for the Hebrew OT.
func (c *Client) TWCB(ctx context.Context, nt bool, number int, simplified bool) (*TWCBResponse, error) {
	endpoint := "stwcbhdic.php"
	if nt {
		endpoint = "sbdag.php"
	}
	q := url.Values{}
	q.Set("k", strconv.Itoa(number))
	q.Set("gb", boolDigit(simplified))
	var out TWCBResponse
	if err := c.getJSON(ctx, endpoint, q, &out); err != nil {
		// A number without an entry makes these endpoints answer HTTP 500 with a cut-off body
		// ({"status":"success","record_count":0,"record":[ ), measured 2026-10-19: treat it as "no entry".
		var ue *Error
		if errors.As(err, &ue) && ue.Kind == ErrHTTP && ue.Status == 500 {
			return &TWCBResponse{Status: "success"}, nil
		}
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
	q.Set("key", hanzi.ForMode(key, simplified))
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
	q.Set("keyword", hanzi.ForMode(keyword, simplified))
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

// AUResponse is the au.php (audio Bible) response. Every field is kept as raw JSON because the
// upstream sends numbers as strings ("3") and the Python output passes them through unchanged.
type AUResponse struct {
	Status    string          `json:"status"`
	Name      json.RawMessage `json:"name"`
	Chinesef  json.RawMessage `json:"chinesef"`
	Engf      json.RawMessage `json:"engf"`
	Chap      json.RawMessage `json:"chap"`
	PBID      json.RawMessage `json:"pbid"`
	PChinesef json.RawMessage `json:"pchinesef"`
	PChap     json.RawMessage `json:"pchap"`
	NBID      json.RawMessage `json:"nbid"`
	NChinesef json.RawMessage `json:"nchinesef"`
	NChap     json.RawMessage `json:"nchap"`
	MP3       json.RawMessage `json:"mp3"`
	OGG       json.RawMessage `json:"ogg"`
}

// AU queries au.php for the audio files of one chapter. version is the upstream audio-version id.
// (The upstream answers success even for a chapter that does not exist.)
func (c *Client) AU(ctx context.Context, version, bid, chapter int, simplified bool) (*AUResponse, error) {
	q := url.Values{}
	q.Set("version", strconv.Itoa(version))
	q.Set("bid", strconv.Itoa(bid))
	q.Set("chap", strconv.Itoa(chapter))
	q.Set("gb", boolDigit(simplified))
	var out AUResponse
	if err := c.getJSON(ctx, "au.php", q, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

// SubRecord is one verse of qsub / qaf / sesub / seaf.php. The numbers are kept raw because the
// Python output passes them through unchanged.
type SubRecord struct {
	Engs      *string         `json:"engs"`
	Chineses  *string         `json:"chineses"`
	Bid       json.RawMessage `json:"bid"`
	Chap      json.RawMessage `json:"chap"`
	Sec       json.RawMessage `json:"sec"`
	BibleText *string         `json:"bible_text"`
}

// SubVerseResponse is the qsub.php / qaf.php response.
type SubVerseResponse struct {
	Status      string          `json:"status"`
	RecordCount json.RawMessage `json:"record_count"`
	VName       *string         `json:"v_name"`
	Version     *string         `json:"version"`
	Record      []SubRecord     `json:"record"`
}

// SubSearchResponse is the sesub.php / seaf.php response.
type SubSearchResponse struct {
	Status      string          `json:"status"`
	Key         *string         `json:"key"`
	RecordCount json.RawMessage `json:"record_count"`
	Record      []SubRecord     `json:"record"`
}

// QSub queries qsub.php (Apocrypha text, upstream book ids 101-115). verse is sent as given
// ("1", "1-5", "1,3,5"); nil asks for the whole chapter.
func (c *Client) QSub(ctx context.Context, bid, chapter int, verse *string) (*SubVerseResponse, error) {
	return c.subVerse(ctx, "qsub.php", bid, chapter, verse)
}

// QAF queries qaf.php (Apostolic Fathers text, upstream book ids 201-217).
func (c *Client) QAF(ctx context.Context, bid, chapter int, verse *string) (*SubVerseResponse, error) {
	return c.subVerse(ctx, "qaf.php", bid, chapter, verse)
}

// SESub searches the Apocrypha (sesub.php). The upstream treats limit 0 as unlimited.
func (c *Client) SESub(ctx context.Context, q string, limit, offset int) (*SubSearchResponse, error) {
	return c.subSearch(ctx, "sesub.php", q, limit, offset)
}

// SEAF searches the Apostolic Fathers (seaf.php).
func (c *Client) SEAF(ctx context.Context, q string, limit, offset int) (*SubSearchResponse, error) {
	return c.subSearch(ctx, "seaf.php", q, limit, offset)
}

func (c *Client) subVerse(ctx context.Context, endpoint string, bid, chapter int, verse *string) (*SubVerseResponse, error) {
	q := url.Values{}
	q.Set("bid", strconv.Itoa(bid))
	q.Set("chap", strconv.Itoa(chapter))
	q.Set("strong", "0")
	if verse != nil {
		q.Set("sec", *verse)
	}
	var out SubVerseResponse
	if err := c.getJSON(ctx, endpoint, q, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

func (c *Client) subSearch(ctx context.Context, endpoint, keyword string, limit, offset int) (*SubSearchResponse, error) {
	q := url.Values{}
	q.Set("q", hanzi.ToTraditional(keyword))
	q.Set("offset", strconv.Itoa(offset))
	q.Set("limit", strconv.Itoa(limit))
	var out SubSearchResponse
	if err := c.getJSON(ctx, endpoint, q, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

// RTRecord is one footnote of rt.php.
type RTRecord struct {
	ID   json.RawMessage `json:"id"`
	Text *string         `json:"text"`
}

// RTResponse is the rt.php (TCV footnote) response. engs is the book abbreviation, empty when nothing was found.
type RTResponse struct {
	Status      string     `json:"status"`
	RecordCount int        `json:"record_count"`
	Version     string     `json:"version"`
	Engs        *string    `json:"engs"`
	Scope       string     `json:"scope"` // rt-nui.php only: "chapter" (lcc/csb) or "book" (other versions)
	Record      []RTRecord `json:"record"`
}

// RT queries rt.php for one footnote of the TCV version (the only version that has footnotes).
func (c *Client) RT(ctx context.Context, bid, id int, simplified bool) (*RTResponse, error) {
	q := url.Values{}
	q.Set("bid", strconv.Itoa(bid))
	q.Set("id", strconv.Itoa(id))
	q.Set("version", "tcv")
	q.Set("gb", boolDigit(simplified))
	var out RTResponse
	if err := c.getJSON(ctx, "rt.php", q, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "status " + strconv.Quote(out.Status)}
	}
	return &out, nil
}

// RTQuery asks for footnotes of any version by upstream book abbreviation (engs, e.g. "1 Sam"). With ID >= 0 it is
// one footnote (rt.php); with ID < 0 it is every footnote of the chapter (lcc, csb) or the whole book (other
// versions), through rt-nui.php. Chap is only sent when > 0.
type RTQuery struct {
	Engs, Version string
	Chap, ID      int
	Simplified    bool
}

// jsonBase is the /json/ twin of the /api/ base: only /json/ has rt-nui.php and honours engs on rt.php
// (measured 2026-10-19; /api/rt.php ignores engs and answers an empty result for non-TCV versions).
func (c *Client) jsonBase() string {
	if strings.HasSuffix(c.BaseURL, "/api") {
		return strings.TrimSuffix(c.BaseURL, "/api") + "/json"
	}
	return c.BaseURL
}

// RTMulti queries rt.php / rt-nui.php with an RTQuery. An upstream "fail" status (unknown version, a version that
// is not public) is an ErrUpstream error with a fixed message (rt.php would echo its SQL otherwise).
func (c *Client) RTMulti(ctx context.Context, rq RTQuery) (*RTResponse, error) {
	q := url.Values{}
	q.Set("engs", rq.Engs)
	q.Set("version", rq.Version)
	q.Set("gb", boolDigit(rq.Simplified))
	if rq.Chap > 0 {
		q.Set("chap", strconv.Itoa(rq.Chap))
	}
	endpoint := "rt-nui.php"
	if rq.ID >= 0 {
		endpoint = "rt.php"
		q.Set("id", strconv.Itoa(rq.ID))
	}
	var out RTResponse
	if err := c.getJSONAt(ctx, c.jsonBase(), endpoint, q, isSuccess, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, &Error{Kind: ErrUpstream, Message: "footnote query failed (version " + strconv.Quote(rq.Version) + " unknown or not public)"}
	}
	return &out, nil
}

// DefaultArticlesBaseURL is the server of the 信望愛 article search.
const DefaultArticlesBaseURL = "https://www.fhl.net/api"

// ArticleQuery holds the article-search conditions; empty ones are not sent.
type ArticleQuery struct {
	Title, Author, Content, Abstract, Column, PubDate string
	Simplified                                        bool
}

// ArticleRecord is one article of json.php. The fields are kept raw so the output passes them through as sent.
type ArticleRecord struct {
	ID      json.RawMessage `json:"id"`
	Column  json.RawMessage `json:"column"`
	Title   json.RawMessage `json:"title"`
	Author  json.RawMessage `json:"author"`
	PubTime json.RawMessage `json:"pubtime"`
	Abst    json.RawMessage `json:"abst"`
	Txt     json.RawMessage `json:"txt"`
}

// ArticlesResponse is the json.php response. Status is 1 for success and 0 for a failure whose reason
// is in Result ("no data", "data too much", ...); json.php is the only endpoint with a numeric status.
type ArticlesResponse struct {
	Status      int             `json:"status"`
	Result      string          `json:"result"`
	RecordCount int             `json:"record_count"`
	Record      []ArticleRecord `json:"record"`
}

// Articles searches the 信望愛 articles. Failure statuses are returned, not raised: the caller decides
// what "no data" means.
func (c *Client) Articles(ctx context.Context, a ArticleQuery) (*ArticlesResponse, error) {
	q := url.Values{}
	q.Set("gb", boolDigit(a.Simplified))
	for k, v := range map[string]string{"title": a.Title, "author": a.Author, "txt": a.Content, "abst": a.Abstract, "ptab": a.Column, "pubtime": a.PubDate} {
		if v != "" {
			if k == "title" || k == "author" || k == "txt" || k == "abst" {
				v = hanzi.ForMode(v, a.Simplified)
			}
			q.Set(k, v)
		}
	}
	base := c.ArticlesBaseURL
	if base == "" {
		base = DefaultArticlesBaseURL
	}
	var out ArticlesResponse
	ok := func(b []byte) bool {
		var s struct {
			Status int `json:"status"`
		}
		return json.Unmarshal(b, &s) == nil && s.Status == 1
	}
	if err := c.getJSONAt(ctx, base, "json.php", q, ok, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// getJSON fetches endpoint (through the cache, when the endpoint is cacheable and a Cache is
// configured) and decodes the JSON body into dst. Only complete, valid responses with
// status "success" are stored; failures and malformed bodies are always asked for again.
func (c *Client) getJSON(ctx context.Context, endpoint string, q url.Values, dst any) error {
	return c.getJSONAt(ctx, c.BaseURL, endpoint, q, isSuccess, dst)
}

// getJSONAt is getJSON for a given server; ok says whether a response may be cached.
func (c *Client) getJSONAt(ctx context.Context, base, endpoint string, q url.Values, ok func([]byte) bool, dst any) error {
	var body []byte
	var err error
	ttl, cacheable := endpointTTL[endpoint]
	if c.Cache != nil && cacheable {
		// The key holds everything that shapes the answer: server, endpoint, every parameter (gb included).
		key := base + "/" + endpoint + "?" + q.Encode()
		body, err = c.Cache.Do(ctx, key, ttl, func(ctx context.Context) ([]byte, bool, error) {
			b, err := c.fetch(ctx, base, endpoint, q)
			return b, err == nil && ok(b), err
		})
	} else {
		body, err = c.fetch(ctx, base, endpoint, q)
	}
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return &Error{Kind: ErrParse, Message: fmt.Sprintf("%s: %v", endpoint, err)}
	}
	return nil
}

// isSuccess reports whether body is valid JSON whose "status" is "success".
func isSuccess(body []byte) bool {
	var s struct {
		Status string `json:"status"`
	}
	return json.Unmarshal(body, &s) == nil && s.Status == "success"
}

// fetch performs one upstream GET and returns the body.
func (c *Client) fetch(ctx context.Context, base, endpoint string, q url.Values) ([]byte, error) {
	u := strings.TrimRight(base, "/") + "/" + endpoint + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, &Error{Kind: ErrNetwork, Message: err.Error()}
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, &Error{Kind: ErrNetwork, Message: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, &Error{Kind: ErrNetwork, Message: err.Error()}
	}
	if resp.StatusCode >= 400 {
		return nil, &Error{Kind: ErrHTTP, Status: resp.StatusCode, Message: fmt.Sprintf("%s returned HTTP %d", endpoint, resp.StatusCode)}
	}
	return body, nil
}

func boolDigit(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
