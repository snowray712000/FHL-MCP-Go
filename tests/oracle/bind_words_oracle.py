import json
import sys

sys.path.insert(0, sys.argv[1])
from parsing_format import bind_words, build_original_text_section, compute_address_tp, format_bcvw

samples = ["a + b c + d e + f", "x + y + z + q", "no markers, here. ok\nnew line", "+ w1 + u1 + tail", "α־β γ"]
for s in samples:
    print(json.dumps(s, ensure_ascii=True), "=>", json.dumps(bind_words(s), ensure_ascii=True))
    print("  text(greek):", json.dumps(build_original_text_section(bind_words(s), (43, 1, 1), 2, False), ensure_ascii=True))
    print("  text(hebrew):", json.dumps(build_original_text_section(bind_words(s), (1, 1, 1), 3, True), ensure_ascii=True))
print(compute_address_tp([(1, 1, 1), (1, 1, 2)]), compute_address_tp([(1, 1, 1), (1, 2, 1)]), compute_address_tp([(1, 1, 1), (2, 1, 1)]), compute_address_tp([(1, 1, 1)]))
print(format_bcvw((1, 2, 3), -1, 3), format_bcvw((1, 2, 3), 4, 2), format_bcvw((1, 2, 3), 4, 1), format_bcvw((1, 2, 3), 4, 0))
