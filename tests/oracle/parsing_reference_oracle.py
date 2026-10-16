import json, sys
sys.path.insert(0, sys.argv[1])
from parsing_format import *

FX = sys.argv[2]
def load(n): return json.load(open(f"{FX}/{n}.json", encoding="utf-8"))

def render(book_id, chinese_full_fallback, refs, qp_names):
    is_hebrew = book_id < 40
    addresses = [(book_id, c, v) for c, v in refs]
    tp = compute_address_tp(addresses)
    blocks = ["# 參考資料: 原文", SCRIPTURE_REFERENCE_LEGEND]
    for (c, v), name in zip(refs, qp_names):
        response = load(name)
        record = response["record"]
        summary = record[0]
        address = (book_id, c, v)
        word_text = summary.get("word", "")
        if is_hebrew:
            word_text = normalize_ot_word_text(word_text)
        bound = bind_words(word_text)
        chinese_full = summary.get("chinesef") or chinese_full_fallback
        blocks.append(f"## 節: {chinese_full} {c}:{v} {format_bcvw(address, -1, tp)}")
        blocks.append("### 類型: 原文")
        blocks.append(build_original_text_section(bound, address, tp, is_hebrew))
        blocks.append("### 類型: 直譯")
        blocks.append(format_translation(summary.get("exp", "")))
        blocks.append("### 類型: 原文分析")
        blocks.append(build_parsing_table_section(record, bound, address, tp, book_id, exclude=[]))
    return "\n\n".join(blocks)

out = sys.argv[3]
cases = {
  "james_1_1": (59, "雅各書", [(1, 1)], ["qp_james_1_1"]),
  "james_1_1_2": (59, "雅各書", [(1, 1), (1, 2)], ["qp_james_1_1", "qp_james_1_2"]),
  "gen_1_1": (1, "創世記", [(1, 1)], ["qp_gen_1_1"]),
  "ruth_1_22_to_2_1": (8, "路得記", [(1, 22), (2, 1)], ["qp_ruth_1_22", "qp_ruth_2_1"]),
}
for k, (b, fb, refs, names) in cases.items():
    open(f"{out}/get_parsing_reference_{k}.md", "w", encoding="utf-8", newline="").write(render(b, fb, refs, names))
    print(k, "ok")
