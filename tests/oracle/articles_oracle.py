# Oracle for M2-j (articles): runs the Python reference's own code - FHLAPIEndpoints.search_articles /
# list_article_columns (extracted from api/endpoints.py by AST, with `httpx` replaced by a replay stub) and
# http_search_articles / http_list_article_columns (from http_tools.py) - on recorded json.php fixtures.
# Only one change: the reference read the article text from the key "content", which json.php does not
# send (it sends "txt"); the oracle replaces those two lookups with "txt" (D25, user-approved fix).
# Usage: python -I articles_oracle.py <endpoints.py> <http_tools.py> <errors.py> <fixtures-dir> <out-dir>
import ast, asyncio, json, logging, os, sys, types
from typing import Any, Optional

ENDPOINTS, HTTP_TOOLS, ERRORS, FIX, OUT = sys.argv[1:6]
PATCHED = 0


def fixture(name):
    with open(os.path.join(FIX, name), encoding="utf-8") as f:
        return json.load(f)


ROUTES = {("愛", "0"): "articles_title_ai.json", ("愛", "1"): "articles_title_ai_gb1.json", ("zzqqxx", "0"): "articles_none.json"}


class _Response:
    def __init__(self, data):
        self._d = data

    def raise_for_status(self):
        pass

    def json(self):
        return self._d


class _Client:
    def __init__(self, *a, **k):
        pass

    async def __aenter__(self):
        return self

    async def __aexit__(self, *a):
        return False

    async def get(self, url, params=None):
        assert url == "https://www.fhl.net/api/json.php", url
        return _Response(fixture(ROUTES[(params.get("title"), str(params["gb"]))]))


httpx = types.SimpleNamespace(AsyncClient=_Client)

errs = {}
exec(compile(open(ERRORS, encoding="utf-8").read(), ERRORS, "exec"), errs)
ns = {"Any": Any, "Optional": Optional, "httpx": httpx, "logger": logging.getLogger("x"),
      "InvalidParameterError": errs["InvalidParameterError"]}

# FHLAPIEndpoints with just the two methods under test.
tree = ast.parse(open(ENDPOINTS, encoding="utf-8").read())
methods = []
for node in tree.body:
    if isinstance(node, ast.ClassDef) and node.name == "FHLAPIEndpoints":
        methods = [n for n in node.body if isinstance(n, (ast.AsyncFunctionDef, ast.FunctionDef)) and n.name in ("search_articles", "list_article_columns")]
cls = ast.ClassDef("FHLAPIEndpoints", [], [], methods, [])
ast.fix_missing_locations(cls)
exec(compile(ast.Module([cls], []), ENDPOINTS, "exec"), ns)
FHL = ns["FHLAPIEndpoints"]
FHL.timeout = 30
FHL.__aenter__ = lambda self: _aenter(self)
FHL.__aexit__ = lambda self, *a: _aexit()


async def _aenter(self):
    return self


async def _aexit():
    return False


ns["FHLAPIEndpoints"] = FHL

# The two http_* functions, with the "content" -> "txt" fix.
tree = ast.parse(open(HTTP_TOOLS, encoding="utf-8").read())
for node in tree.body:
    if isinstance(node, ast.AsyncFunctionDef) and node.name in ("http_search_articles", "http_list_article_columns"):
        for sub in ast.walk(node):
            if isinstance(sub, ast.Call) and isinstance(sub.func, ast.Attribute) and sub.func.attr == "get" and sub.args \
                    and isinstance(sub.args[0], ast.Constant) and sub.args[0].value == "content":
                sub.args[0].value = "txt"
                PATCHED += 1
        exec(compile(ast.Module([node], []), HTTP_TOOLS, "exec"), ns)
assert PATCHED == 2, PATCHED

cases = {
    "search_fhl_articles_preview.json": lambda: ns["http_search_articles"](title="愛", limit=2),
    "search_fhl_articles_full.json": lambda: ns["http_search_articles"](title="愛", limit=2, include_content=True),
    "search_fhl_articles_simplified.json": lambda: ns["http_search_articles"](title="愛", limit=1, use_simplified=True),
    "search_fhl_articles_all3.json": lambda: ns["http_search_articles"](title="愛", limit=10),
    "search_fhl_articles_none.json": lambda: ns["http_search_articles"](title="zzqqxx"),
    "list_fhl_article_columns.json": lambda: ns["http_list_article_columns"](),
}
os.makedirs(OUT, exist_ok=True)
for name, fn in cases.items():
    res = asyncio.run(fn())
    with open(os.path.join(OUT, name), "wb") as f:
        f.write(json.dumps(res, ensure_ascii=False, indent=2).encode("utf-8"))
    print("wrote", name)
