package bible

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// uiabvAPI adds the optional uiabv.php list on top of the plain fake (whose AB always fails).
type uiabvAPI struct {
	fakeAPI
	fail bool
}

func (f *uiabvAPI) AB(context.Context, bool) (*fhlapi.ABResponse, error) {
	return &fhlapi.ABResponse{Status: "success", Record: []fhlapi.ABRecord{{Book: "unv", CName: "ab", Strong: 1}}}, nil
}

func (f *uiabvAPI) UIABV(context.Context, bool) (*fhlapi.ABResponse, error) {
	if f.fail {
		return nil, errors.New("down")
	}
	b, err := os.ReadFile(filepath.Join("..", "..", "tests", "fixtures", "upstream", "uiabv.json"))
	if err != nil {
		return nil, err
	}
	var out fhlapi.ABResponse
	return &out, json.Unmarshal(b, &out)
}

func TestListVersionsPrefersUIABVAndFallsBack(t *testing.T) {
	res, err := (&Service{API: &uiabvAPI{}}).ListVersions(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]VersionItem{}
	for _, v := range res.Versions {
		by[v.Code] = v
	}
	if res.TotalCount != 6 || by["esv"].Code == "" || by["lxx"].Testament != "ot_only" || by["gnt6"].Testament != "nt_only" || !by["unv"].CanDownload || by["unv"].Version == "" {
		t.Errorf("uiabv list = %+v", res.Versions)
	}
	res, err = (&Service{API: &uiabvAPI{fail: true}}).ListVersions(context.Background(), false)
	if err != nil || res.TotalCount != 1 || res.Versions[0].Name != "ab" {
		t.Errorf("fallback to ab.php: %v %+v", err, res)
	}
}
