package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// twcbAPI answers sd.php (CBOL) and the TWCB dictionaries from recorded fixtures.
type twcbAPI struct {
	footnoteAPI
	twcb     string // fixture of TWCB; "" = no entry
	twcbCall []bool
}

func (f *twcbAPI) SD(context.Context, bool, int, bool) (*fhlapi.SDResponse, error) {
	var out fhlapi.SDResponse
	return &out, json.Unmarshal(fixture(f.t, "upstream", "sd_nt_25.json"), &out)
}

func (f *twcbAPI) TWCB(_ context.Context, nt bool, _ int, _ bool) (*fhlapi.TWCBResponse, error) {
	f.twcbCall = append(f.twcbCall, nt)
	out := fhlapi.TWCBResponse{Status: "success"}
	if f.twcb == "" {
		return &out, nil
	}
	return &out, json.Unmarshal(fixture(f.t, "upstream", f.twcb), &out)
}

func TestLookupStrongsTWCB(t *testing.T) {
	api := &twcbAPI{footnoteAPI: footnoteAPI{t: t}, twcb: "sbdag_3303.json"}
	got, err := callOn(t, api, "lookup_strongs", `{"number":"G3303","dictionary":"TWCB"}`)
	if err != nil {
		t.Fatal(err)
	}
	var res struct {
		StrongsNumber, Dictionary, Testament, Definition string
	}
	if err := json.Unmarshal([]byte(strings.NewReplacer("strongs_number", "strongsnumber").Replace(got)), &res); err != nil {
		t.Fatal(err)
	}
	if res.StrongsNumber != "03303" || res.Dictionary != "TWCB" || res.Testament != "NT" || !strings.Contains(strings.SplitN(res.Definition, "\n", 2)[0], " 連接詞") {
		t.Errorf("result = %.200s", got)
	}
	for _, bad := range []string{"<span", "<div", "\r", "#可9:12", "|"} {
		if strings.Contains(res.Definition, bad) {
			t.Errorf("definition still contains %q", bad)
		}
	}
	if len(api.twcbCall) != 1 || !api.twcbCall[0] {
		t.Errorf("twcb calls = %v", api.twcbCall)
	}
}

func TestLookupStrongsBothAndCBOLUnchanged(t *testing.T) {
	api := &twcbAPI{footnoteAPI: footnoteAPI{t: t}, twcb: "stwcbhdic_128.json"}
	both, err := callOn(t, api, "lookup_strongs", `{"number":"25","testament":"NT","dictionary":"both"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(both, `"chinese_definition"`) || !strings.Contains(both, `"twcb_definition"`) || !strings.Contains(both, "亞大瑪") {
		t.Errorf("both = %.300s", both)
	}
	plain, err := callOn(t, api, "lookup_strongs", `{"number":"25","testament":"NT"}`)
	if err != nil {
		t.Fatal(err)
	}
	explicit, err := callOn(t, api, "lookup_strongs", `{"number":"25","testament":"NT","dictionary":"cbol"}`)
	if err != nil {
		t.Fatal(err)
	}
	if plain != explicit || strings.Contains(plain, "twcb") || len(api.twcbCall) != 1 {
		t.Errorf("default must not touch TWCB; calls = %v", api.twcbCall)
	}
	// no TWCB entry: "both" still answers with the CBOL entry, "twcb" alone reports not found
	api = &twcbAPI{footnoteAPI: footnoteAPI{t: t}}
	out, err := callOn(t, api, "lookup_strongs", `{"number":"G25","dictionary":"both"}`)
	if err != nil || strings.Contains(out, "twcb_definition") {
		t.Errorf("both without TWCB entry: %v %.200s", err, out)
	}
	if _, err = callOn(t, api, "lookup_strongs", `{"number":"G25","dictionary":"twcb"}`); err == nil || !strings.Contains(err.Error(), "not found in the TWCB dictionary") {
		t.Errorf("twcb not found: %v", err)
	}
}

func TestLookupStrongsDictionaryValidation(t *testing.T) {
	_, err := callOn(t, &twcbAPI{footnoteAPI: footnoteAPI{t: t}}, "lookup_strongs", `{"number":"G25","dictionary":"x"}`)
	if err == nil || err.Error() != "Invalid parameter 'dictionary': x - must be 'cbol', 'twcb' or 'both'" {
		t.Errorf("err = %v", err)
	}
	_, err = callOn(t, &footnoteAPI{t: t}, "lookup_strongs", `{"number":"G25","dictionary":"twcb"}`)
	if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Errorf("unsupported client: %v", err)
	}
}
