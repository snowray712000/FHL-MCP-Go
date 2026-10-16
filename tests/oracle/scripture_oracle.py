# Oracle for M2-g/h (Apocrypha, Apostolic Fathers): runs the six http_* functions of the Python reference's
# http_tools.py (extracted verbatim by AST) against recorded qsub/qaf/sesub/seaf fixtures. The only thing
# replaced is the book table: the reference numbered the books differently from the upstream (D24), so the
# oracle uses internal/bible/scripture_books.json (the Go table, upstream numbering) in the reference's
# dict shape. Everything else - field order, defaults, the book/bid lookups - is the reference's own code.
# Usage: python -I scripture_oracle.py <http_tools.py> <books.json> <fixtures-dir> <out-dir>
import ast, asyncio, json, os, sys
from typing import Any, Optional

HTTP_TOOLS, BOOKS, FIX, OUT = sys.argv[1:5]
WANT = {"http_get_apocrypha_verse", "http_search_apocrypha", "http_list_apocrypha_books",
        "http_get_apostolic_fathers_verse", "http_search_apostolic_fathers", "http_list_apostolic_fathers_books"}


def table(key):
    out = {}
    for b in json.load(open(BOOKS, encoding="utf-8"))[key]["books"]:
        for a in b["aliases"]:
            out[a] = {"id": b["id"], "name_zh": b["name_zh"], "name_en": b["name_en"]}
    return out


def fixture(name):
    with open(os.path.join(FIX, name), encoding="utf-8") as f:
        return json.load(f)


class FHLAPIEndpoints:
    ROUTES = {}

    async def __aenter__(self):
        return self

    async def __aexit__(self, *a):
        return False

    def _go(self, *key):
        return fixture(self.ROUTES[key])

    async def get_apocrypha_verse(self, book, chapter, verse=None):
        return self._go("qsub", book, chapter, verse)

    async def get_apostolic_fathers_verse(self, book, chapter, verse=None):
        return self._go("qaf", book, chapter, verse)

    async def search_apocrypha(self, query, limit=None, offset=0):
        return self._go("sesub", query, limit, offset)

    async def search_apostolic_fathers(self, query, limit=None, offset=0):
        return self._go("seaf", query, limit, offset)


FHLAPIEndpoints.ROUTES = {
    ("qsub", "多", 1, "1"): "qsub_tob_1_1.json",
    ("qsub", "Tobit", 1, "1-3"): "qsub_tob_1_1-3.json",
    ("qsub", "加上", 1, "1,3"): "qsub_1mac_1_1_3.json",
    ("qsub", "多俾亞傳", 99, None): "qsub_tob_99.json",
    ("qaf", "革利免前書", 1, "1"): "qaf_1clem_1_1.json",
    ("qaf", "Didache", 1, "1-2"): "qaf_did_1_1-2.json",
    ("qaf", "1Clem", 99, None): "qaf_1clem_99.json",
    ("sesub", "智慧", 3, 0): "sesub_wisdom_limit3.json",
    ("sesub", "智慧", 2, 2): "sesub_wisdom_off2_limit2.json",
    ("sesub", "zzqqxx", 10, 0): "sesub_none.json",
    ("seaf", "教會", 3, 0): "seaf_church_limit3.json",
    ("seaf", "zzqqxx", 10, 0): "seaf_none.json",
}

ns = {"Optional": Optional, "Any": Any, "FHLAPIEndpoints": FHLAPIEndpoints,
      "APOCRYPHA_BOOKS": table("apocrypha"), "APOSTOLIC_FATHERS_BOOKS": table("apostolic_fathers")}
tree = ast.parse(open(HTTP_TOOLS, encoding="utf-8").read())
for node in tree.body:
    if isinstance(node, ast.AsyncFunctionDef) and node.name in WANT:
        exec(compile(ast.Module([node], []), HTTP_TOOLS, "exec"), ns)
missing = WANT - set(ns)
assert not missing, missing

cases = {
    "get_apocrypha_verse_tob1_1.json": lambda: ns["http_get_apocrypha_verse"]("多", 1, "1"),
    "get_apocrypha_verse_tob1_1-3.json": lambda: ns["http_get_apocrypha_verse"]("Tobit", 1, "1-3"),
    "get_apocrypha_verse_1mac1_1_3.json": lambda: ns["http_get_apocrypha_verse"]("加上", 1, "1,3"),
    "get_apocrypha_verse_tob99_empty.json": lambda: ns["http_get_apocrypha_verse"]("多俾亞傳", 99),
    "search_apocrypha_wisdom.json": lambda: ns["http_search_apocrypha"]("智慧", 3, 0),
    "search_apocrypha_wisdom_off2.json": lambda: ns["http_search_apocrypha"]("智慧", 2, 2),
    "search_apocrypha_none.json": lambda: ns["http_search_apocrypha"]("zzqqxx", 10, 0),
    "list_apocrypha_books.json": lambda: ns["http_list_apocrypha_books"](),
    "get_apostolic_fathers_verse_1clem1_1.json": lambda: ns["http_get_apostolic_fathers_verse"]("革利免前書", 1, "1"),
    "get_apostolic_fathers_verse_did1_1-2.json": lambda: ns["http_get_apostolic_fathers_verse"]("Didache", 1, "1-2"),
    "get_apostolic_fathers_verse_1clem99_empty.json": lambda: ns["http_get_apostolic_fathers_verse"]("1Clem", 99),
    "search_apostolic_fathers_church.json": lambda: ns["http_search_apostolic_fathers"]("教會", 3, 0),
    "search_apostolic_fathers_none.json": lambda: ns["http_search_apostolic_fathers"]("zzqqxx", 10, 0),
    "list_apostolic_fathers_books.json": lambda: ns["http_list_apostolic_fathers_books"](),
}
os.makedirs(OUT, exist_ok=True)
for name, fn in cases.items():
    res = asyncio.run(fn())
    with open(os.path.join(OUT, name), "wb") as f:
        f.write(json.dumps(res, ensure_ascii=False, indent=2).encode("utf-8"))
    print("wrote", name)
