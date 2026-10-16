package bible

// scriptureReferenceLegend is copied verbatim from the Python reference (utils/parsing_format.py).
const scriptureReferenceLegend = "- 標記說明: `b` `c` `v` `w` (用於標示聖經原文的章節位置與詞索引)\n\t- b = book (書卷)\n\t- c = chapter (章)\n\t- v = verse (節)\n\t- w = word (詞索引，用於標示該節中的原文字，編號為唯一ID，通常遞增，但不一定連續)\n\t- 範例: b1c1v1w1 = 創世記 1:1 的第一個原文詞\n\t- 若內容都在同一節中，則可省略 v，例如: w1 w2 w3\n\t- 若內容都在同一章中，則可省略 c，例如: v1w1 v2w1\n\t- 若內容都在同一書中，則可省略 b，例如: c1v31w1 c2v1w1"
