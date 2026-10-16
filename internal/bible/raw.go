package bible

import (
	"bytes"
	"encoding/json"
)

// norm re-encodes an upstream JSON value the way Python's json.dumps(ensure_ascii=False) prints it: the
// upstream escapes non-ASCII text as \uXXXX, and a passthrough RawMessage would otherwise keep those escapes.
// Numbers keep their literal form. A value that does not parse is returned unchanged.
func norm(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return raw
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return raw
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return raw
	}
	return json.RawMessage(bytes.TrimRight(buf.Bytes(), "\n"))
}
