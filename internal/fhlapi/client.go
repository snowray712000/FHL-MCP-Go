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
