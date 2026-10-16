// Package tools adapts MCP tool calls (raw JSON arguments) to the Bible domain service.
package tools

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
)

// args is a decoded tool-call argument object. Validation follows the approved
// differences: optional null == absent (D5), integer parameters accept numeric strings (D6),
// messages are short and custom, not Pydantic text (D2).
type args map[string]json.RawMessage

func parseArgs(raw json.RawMessage) (args, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return args{}, nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, &bible.InvalidParameterError{Name: "arguments", Reason: "must be a JSON object"}
	}
	return args(m), nil
}

// get returns the raw value, treating an absent key and an explicit null alike.
func (a args) get(name string) (json.RawMessage, bool) {
	v, ok := a[name]
	if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
		return nil, false
	}
	return v, true
}

func shown(v json.RawMessage) string {
	s := strings.TrimSpace(string(v))
	if len(s) > 80 {
		s = s[:80] + "..."
	}
	return s
}

func (a args) requiredString(name string) (string, error) {
	v, ok := a.get(name)
	if !ok {
		return "", &bible.InvalidParameterError{Name: name, Reason: "required parameter is missing"}
	}
	return decodeString(name, v)
}

func (a args) optString(name, def string) (string, error) {
	v, ok := a.get(name)
	if !ok {
		return def, nil
	}
	return decodeString(name, v)
}

func decodeString(name string, v json.RawMessage) (string, error) {
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return "", &bible.InvalidParameterError{Name: name, Value: shown(v), Reason: "must be a string"}
	}
	return s, nil
}

func (a args) requiredInt(name string) (int, error) {
	v, ok := a.get(name)
	if !ok {
		return 0, &bible.InvalidParameterError{Name: name, Reason: "required parameter is missing"}
	}
	return decodeInt(name, v)
}

func decodeInt(name string, v json.RawMessage) (int, error) {
	bad := &bible.InvalidParameterError{Name: name, Value: shown(v), Reason: "must be an integer"}
	var text string
	if bytes.HasPrefix(bytes.TrimSpace(v), []byte(`"`)) {
		if err := json.Unmarshal(v, &text); err != nil {
			return 0, bad
		}
		text = strings.TrimSpace(text)
	} else {
		text = strings.TrimSpace(string(v))
	}
	if n, err := strconv.Atoi(text); err == nil {
		return n, nil
	}
	// 3.0 style values: accept only when exactly integral.
	if f, err := strconv.ParseFloat(text, 64); err == nil && f == math.Trunc(f) && math.Abs(f) < 1e9 && !strings.ContainsAny(text, "nNiI") {
		return int(f), nil
	}
	return 0, bad
}

func (a args) optBool(name string, def bool) (bool, error) {
	v, ok := a.get(name)
	if !ok {
		return def, nil
	}
	var b bool
	if err := json.Unmarshal(v, &b); err == nil {
		return b, nil
	}
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off":
			return false, nil
		}
	}
	return false, &bible.InvalidParameterError{Name: name, Value: shown(v), Reason: "must be true or false"}
}

// optStringPtr returns nil when the parameter is absent or null (D5).
func (a args) optStringPtr(name string) (*string, error) {
	v, ok := a.get(name)
	if !ok {
		return nil, nil
	}
	s, err := decodeString(name, v)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// optBoolPtr returns nil when the parameter is absent or null (D5).
func (a args) optBoolPtr(name string) (*bool, error) {
	if _, ok := a.get(name); !ok {
		return nil, nil
	}
	b, err := a.optBool(name, false)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
