// Package version holds the product version. It must match `last` in app_versions.md.
package version

// Version is the product version.
const Version = "1.1.2.16"

// Name is the product name reported in serverInfo and logs.
const Name = "FHL-MCP-Go"

// Commit and Date are stamped by tools/release with -ldflags "-X ...". They are
// empty in a plain `go build`; Describe then reports only the version.
var (
	Commit string
	Date   string
)

// Describe returns "commit abc1234, built 2026-10-16" (or "" when not stamped).
func Describe() string {
	switch {
	case Commit != "" && Date != "":
		return "commit " + Commit + ", built " + Date
	case Commit != "":
		return "commit " + Commit
	}
	return ""
}
