package bible

import "strings"

// s2tBookChars maps the simplified forms of the characters used in the book-name tables
// (canon, Apocrypha, Apostolic Fathers) to the traditional forms the tables use.
// Only characters that differ are listed; a name that fails to resolve as typed is retried
// after this folding, so traditional input is never altered.
var s2tBookChars = map[rune]rune{
	'丢': '丟', '亚': '亞', '来': '來', '传': '傳', '内': '內', '创': '創', '参': '參', '叁': '參',
	'启': '啟', '弥': '彌', '后': '後', '数': '數', '书': '書', '历': '歷', '残': '殘', '犹': '猶',
	'玛': '瑪', '祷': '禱', '约': '約', '纳': '納', '结': '結', '统': '統', '罗': '羅', '与': '與',
	'萨': '薩', '苏': '蘇', '众': '衆', '眾': '衆', '补': '補', '训': '訓', '词': '詞', '诗': '詩',
	'该': '該', '赛': '賽', '遗': '遺', '录': '錄', '铁': '鐵', '长': '長', '门': '門', '马': '馬',
	'鸿': '鴻', '龙': '龍', '纪': '紀', '记': '記', '师': '師', '贰': '貳',
}

// foldSimplifiedBook converts simplified characters of a book name to traditional ones.
// It reports whether anything changed.
func foldSimplifiedBook(name string) (string, bool) {
	changed := false
	var b strings.Builder
	for _, r := range name {
		if t, ok := s2tBookChars[r]; ok && t != r {
			r, changed = t, true
		}
		b.WriteRune(r)
	}
	return b.String(), changed
}
