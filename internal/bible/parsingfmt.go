package bible

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// This file is a Go port of the Python reference's utils/parsing_format.py (itself a port of the
// FHLNUI ai_parsing_gen_tp1 JS modules): it turns qp.php rows into the three-section Markdown
// used by get_parsing_reference. Output must stay byte-identical to the reference.

const (
	lrm   = "‎" // Left-to-Right Mark
	rli   = "⁧" // Right-to-Left Isolate
	pdi   = "⁩" // Pop Directional Isolate
	maqaf = "־" // Hebrew maqaf
)

var (
	tokenRE   = regexp.MustCompile(`[^ \n\r\x{05BE},.]+`)
	newlineRE = regexp.MustCompile(`\r?\n\r?`)
	rtlRE     = regexp.MustCompile(`[\x{0590}-\x{05FF}\x{0600}-\x{06FF}\x{0750}-\x{077F}]+`)
)

// address is (book, chapter, verse).
type address struct{ book, chapter, verse int }

// computeAddressTP decides how short token ids may be: 0 w only, 1 v+w, 2 c+v+w, 3 b+c+v+w.
func computeAddressTP(addrs []address) int {
	books, chapters, verses := map[int]bool{}, map[int]bool{}, map[int]bool{}
	for _, a := range addrs {
		books[a.book], chapters[a.chapter], verses[a.verse] = true, true, true
	}
	switch {
	case len(books) > 1:
		return 3
	case len(chapters) > 1:
		return 2
	case len(verses) > 1:
		return 1
	}
	return 0
}

// formatBCVW builds b1c1v1w12 style token ids; wid -1 omits the word part.
func formatBCVW(a address, wid, tp int) string {
	w := ""
	if wid != -1 {
		w = "w" + strconv.Itoa(wid)
	}
	switch {
	case tp <= 0:
		return w
	case tp == 1:
		return "v" + strconv.Itoa(a.verse) + w
	case tp == 2:
		return "c" + strconv.Itoa(a.chapter) + "v" + strconv.Itoa(a.verse) + w
	}
	return "b" + strconv.Itoa(a.book) + "c" + strconv.Itoa(a.chapter) + "v" + strconv.Itoa(a.verse) + w
}

// boundWord is one token (wid > 0) or one run of delimiters (wid 0) of the verse text.
type boundWord struct {
	w   string
	wid int
	wu  byte // 'w' Westcott-Hort variant, 'u' UBS4 variant, 0 outside a +...+ group
}

// bindWords splits the verse text into tokens numbered like the upstream wid column and marks
// the +w+u+ variant groups.
func bindWords(text string) []boundWord {
	var out []boundWord
	pos, wid := 0, 1
	for _, loc := range tokenRE.FindAllStringIndex(text, -1) {
		if loc[0] > pos {
			out = append(out, boundWord{w: text[pos:loc[0]]})
		}
		out = append(out, boundWord{w: text[loc[0]:loc[1]], wid: wid})
		wid++
		pos = loc[1]
	}
	if pos < len(text) {
		out = append(out, boundWord{w: text[pos:]})
	}
	i, n := 0, len(out)
	for i < n {
		if out[i].w != "+" {
			i++
			continue
		}
		i++
		for i < n && out[i].w != "+" {
			out[i].wu = 'w'
			i++
		}
		i++
		for i < n && out[i].w != "+" {
			out[i].wu = 'u'
			i++
		}
		i++
	}
	return out
}

// normalizeOTWordText drops \r and reverses the line order (upstream lists Old Testament lines
// in the opposite order to the translation).
func normalizeOTWordText(word string) string {
	lines := strings.Split(strings.ReplaceAll(word, "\r", ""), "\n")
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return strings.Join(lines, "\n")
}

func formatTranslation(exp string) string {
	return newlineRE.ReplaceAllString(strings.ReplaceAll(exp, "\r", ""), " ↩ ")
}

func buildOriginalTextSection(words []boundWord, a address, tp int, hebrew bool) string {
	newline := " ↩ "
	if hebrew {
		newline = " ↪ "
	}
	var b strings.Builder
	if hebrew {
		b.WriteString(rli)
	}
	for _, it := range words {
		if it.w == "+" || it.wu == 'w' {
			continue
		}
		if it.wid != 0 {
			b.WriteString(formatBCVW(a, it.wid, tp) + " `" + it.w + "`")
			continue
		}
		w := it.w
		if hebrew && strings.Contains(w, maqaf) {
			w = strings.ReplaceAll(w, maqaf, " "+maqaf+" ")
		}
		b.WriteString(newlineRE.ReplaceAllString(w, newline))
	}
	if hebrew {
		b.WriteString(pdi)
	}
	return b.String()
}

// pyStr renders an optional JSON field the way a Python f-string would: absent -> def,
// null -> "None", string -> itself, number/bool -> its JSON text.
func pyStr(raw json.RawMessage, def string) string {
	switch v := rawOr(raw, def).(type) {
	case nil:
		return "None"
	case string:
		return v
	case json.Number:
		return v.String()
	case bool:
		if v {
			return "True"
		}
		return "False"
	default:
		return fmt.Sprint(v)
	}
}

// pyOrEmpty mirrors `(r.get(k) or "")`: absent, null, "" and 0 all give "".
func pyOrEmpty(raw json.RawMessage) string {
	switch v := rawOr(raw, "").(type) {
	case string:
		return v
	case json.Number:
		if f, err := v.Float64(); err == nil && f != 0 {
			return v.String()
		}
	case bool:
		if v {
			return "True"
		}
	}
	return ""
}

// widOf reads the integer wid; ok is false when it is absent or not an integer.
func widOf(raw json.RawMessage) (int, bool) {
	n, ok := rawOr(raw, nil).(json.Number)
	if !ok {
		return 0, false
	}
	i, err := strconv.Atoi(n.String())
	return i, err == nil
}

func fixHebrew(s string) string {
	return rtlRE.ReplaceAllStringFunc(s, func(m string) string { return m + lrm })
}

func orSpace(s string) string {
	if s == "" {
		return " "
	}
	return s
}

func snShorter(raw json.RawMessage) string {
	return strings.TrimLeft(pyOrEmpty(raw), "0")
}

func buildGreekTable(rows []fhlapi.QPRecord, words []boundWord, a address, tp int) (string, error) {
	skip := map[int]bool{}
	for _, w := range words {
		if w.wid != 0 && w.wu == 'w' {
			skip[w.wid] = true
		}
	}
	var lines []string
	for _, r := range rows[1:] {
		if s, ok := rawOr(r.Word, nil).(string); ok && s == "+" {
			continue
		}
		wid, ok := widOf(r.WID)
		if !ok {
			return "", fmt.Errorf("upstream word row has no usable wid")
		}
		if skip[wid] {
			continue
		}
		lines = append(lines, "- "+strings.Join([]string{
			"詞索引: `" + formatBCVW(a, wid, tp) + "`",
			"SN: `G" + snShorter(r.SN) + "`",
			"原文字: `" + pyStr(r.Word, "") + "`",
			"詞性: `" + orSpace(strings.TrimSpace(pyOrEmpty(r.Pro))) + "`",
			"字彙分析: `" + orSpace(strings.TrimSpace(pyOrEmpty(r.WForm))) + "`",
			"原型: `" + pyStr(r.Orig, "") + "`",
			"原型簡義: `" + pyStr(r.Exp, "") + "`",
			"備註: `" + orSpace(strings.TrimSpace(pyOrEmpty(r.Remark))) + "`",
		}, " | "))
	}
	return strings.Join(lines, "\n"), nil
}

func buildHebrewTable(rows []fhlapi.QPRecord, a address, tp int) (string, error) {
	var lines []string
	for _, r := range rows[1:] {
		wid, ok := widOf(r.WID)
		if !ok {
			return "", fmt.Errorf("upstream word row has no usable wid")
		}
		lines = append(lines, "- "+strings.Join([]string{
			"詞索引: `" + formatBCVW(a, wid, tp) + "`",
			"SN: `H" + snShorter(r.SN) + "`",
			"原文字: `" + pyStr(r.Word, "") + "`",
			"字彙分析: `" + orSpace(fixHebrew(strings.TrimSpace(pyOrEmpty(r.WForm)))) + "`",
			"原型: `" + pyStr(r.Orig, "") + "`",
			"原型簡義: `" + pyStr(r.Exp, "") + "`",
			"備註: `" + orSpace(fixHebrew(strings.TrimSpace(pyOrEmpty(r.Remark)))) + "`",
		}, " | "))
	}
	return strings.Join(lines, "\n"), nil
}
