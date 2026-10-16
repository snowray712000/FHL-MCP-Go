"""Replica of tools/strongs.py get_word_analysis (formatting part only) for offline goldens."""
import json
import sys

FX, OUT = sys.argv[1], sys.argv[2]


def word_analysis(response, book, chapter, verse):
    testament = "OT" if response["N"] == 1 else "NT"
    original_text = ""
    translation = ""
    remark = ""
    if response["record"] and len(response["record"]) > 0:
        first_record = response["record"][0]
        if first_record.get("wid") == 0:
            original_text = first_record.get("word", "")
            translation = first_record.get("exp", "")
            remark = first_record.get("remark", "")
    words = []
    for record in response["record"]:
        if record.get("wid") and record["wid"] > 0:
            word_info = {
                "position": record["wid"],
                "word": record.get("word", ""),
                "strongs_number": record.get("sn", ""),
                "part_of_speech": record.get("pro", ""),
                "morphology": record.get("wform", ""),
                "lemma": record.get("orig", ""),
                "gloss": record.get("exp", ""),
            }
            if record.get("remark"):
                word_info["remark"] = record["remark"]
            words.append(word_info)
    result = {
        "testament": testament,
        "book": book,
        "chapter": chapter,
        "verse": verse,
        "original_text": original_text,
        "translation": translation,
        "word_count": len(words),
        "words": words,
    }
    if remark:
        result["remark"] = remark
    return result


cases = {
    "gen_1_1": ("qp_gen_1_1", "Gen", 1, 1),
    "john_1_1": ("qp_john_1_1", "John", 1, 1),
    "ruth_1_22": ("qp_ruth_1_22", "路得記", 1, 22),
}
for key, (fx, book, c, v) in cases.items():
    resp = json.load(open(f"{FX}/{fx}.json", encoding="utf-8"))
    text = json.dumps(word_analysis(resp, book, c, v), ensure_ascii=False, indent=2)
    open(f"{OUT}/get_word_analysis_{key}.json", "w", encoding="utf-8", newline="").write(text)
    print(key, len(text))
