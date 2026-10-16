package bible

import (
	"context"
	"strings"
)

// VersionItem is one Bible version in list_bible_versions (field order is part of the contract).
type VersionItem struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	HasStrongs  bool   `json:"has_strongs"`
	Testament   string `json:"testament"` // both | nt_only | ot_only
	SpecialFont string `json:"special_font"`
	CanDownload bool   `json:"can_download"`
	Version     string `json:"version"`
}

// VersionListResult is the list_bible_versions result object.
type VersionListResult struct {
	TotalCount int           `json:"total_count"`
	Versions   []VersionItem `json:"versions"`
}

// ListVersions lists every upstream Bible version (ab.php).
func (s *Service) ListVersions(ctx context.Context, simplified bool) (*VersionListResult, error) {
	resp, err := s.API.AB(ctx, simplified)
	if err != nil {
		return nil, err
	}
	items := make([]VersionItem, 0, len(resp.Record))
	for _, r := range resp.Record {
		testament := "both"
		if r.NTOnly == 1 {
			testament = "nt_only"
		} else if r.OTOnly == 1 {
			testament = "ot_only"
		}
		font := "none"
		switch r.Proc {
		case 1:
			font = "greek"
		case 2:
			font = "hebrew"
		case 3:
			font = "roman"
		case 4:
			font = "openhan"
		}
		items = append(items, VersionItem{
			Code: r.Book, Name: r.CName, HasStrongs: r.Strong == 1, Testament: testament,
			SpecialFont: font, CanDownload: r.CanDownload == 1, Version: r.Version,
		})
	}
	return &VersionListResult{TotalCount: len(items), Versions: items}, nil
}

// VersionFilters echoes the filters as given (null when absent).
type VersionFilters struct {
	HasStrongs *bool   `json:"has_strongs"`
	Testament  *string `json:"testament"`
}

// VersionSearchResult is the search_available_versions result object.
type VersionSearchResult struct {
	TotalCount int            `json:"total_count"`
	Filters    VersionFilters `json:"filters"`
	Versions   []VersionItem  `json:"versions"`
}

// SearchVersions filters the version list locally. Like the Python reference:
// has_strongs must match exactly; testament "OT"/"NT" (case-insensitive) keeps
// versions that are "both" or only that testament, any other value (including
// "both" and "") applies no filter; the filters are echoed unchanged.
func (s *Service) SearchVersions(ctx context.Context, hasStrongs *bool, testament *string, simplified bool) (*VersionSearchResult, error) {
	all, err := s.ListVersions(ctx, simplified)
	if err != nil {
		return nil, err
	}
	only := ""
	if testament != nil {
		switch strings.ToUpper(*testament) {
		case "OT":
			only = "ot_only"
		case "NT":
			only = "nt_only"
		}
	}
	out := make([]VersionItem, 0, len(all.Versions))
	for _, v := range all.Versions {
		if hasStrongs != nil && v.HasStrongs != *hasStrongs {
			continue
		}
		if only != "" && v.Testament != "both" && v.Testament != only {
			continue
		}
		out = append(out, v)
	}
	return &VersionSearchResult{
		TotalCount: len(out),
		Filters:    VersionFilters{HasStrongs: hasStrongs, Testament: testament},
		Versions:   out,
	}, nil
}
