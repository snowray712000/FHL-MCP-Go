// Package assets holds data embedded into the binary.
package assets

import _ "embed"

// ToolsListJSON is the frozen tools/list result of the Python reference server (29 tools,
// inputSchema + outputSchema). It is a verbatim copy of doc/m0/contract/tools-list.json;
// assets_test.go keeps the two in sync.
//
//go:embed tools-list.json
var ToolsListJSON []byte
