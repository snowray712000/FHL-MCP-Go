# Oracle for M2-e: runs the Python reference's own tools/commentary.py on recorded upstream fixtures.
# Assemble a scratch package first: fhl_bible_mcp/{__init__,tools/__init__,utils/__init__,api/__init__}.py (empty),
# copy tools/commentary.py, utils/booknames.py, utils/errors.py from the reference, and put
# commentary_stub_endpoints.py at fhl_bible_mcp/api/endpoints.py. Then: python -I commentary_oracle.py <fixtures-dir> <out-dir>
import asyncio, json, sys, os
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__))))
from fhl_bible_mcp.api import endpoints
from fhl_bible_mcp.tools import commentary as C

FIX = sys.argv[1]
OUT = sys.argv[2]


def load(name):
    with open(os.path.join(FIX, name), encoding="utf-8") as f:
        return json.load(f)


def router(method, **kw):
    if method == "list_commentaries":
        return load("sc_list.json")
    if method == "get_commentary":
        if kw["verse"] is None:
            return load("sc_john_3_nosec.json")
        if kw["chapter"] == 99:
            return load("sc_john_99_1.json")
        return load("sc_john_3_16.json")
    if method == "search_commentary":
        return {"尼哥底母": load("ssc_nicodemus.json"), "zzqqxx": load("ssc_none.json")}[kw["keyword"]]
    if method == "get_topic_study":
        k, s, c = kw["keyword"], kw["source"], kw["count_only"]
        table = {
            ("愛", "all", True): "st_love_count.json",
            ("愛", "all", False): "st_love_all.json",
            ("faith", "torrey_en", False): "st_faith_torrey_en.json",
            ("zzqqxx", "all", False): "st_none.json",
            ("zzqqxx", "all", True): "st_none_count.json",
        }
        return load(table[(k, s, c)])
    raise KeyError(method)


endpoints.ROUTER = router
cases = {
    "list_commentaries.json": lambda: C.list_commentaries(),
    "get_commentary_john316.json": lambda: C.get_commentary("John", 3, 16),
    "get_commentary_john99_empty.json": lambda: C.get_commentary("John", 99, 1),
    "get_commentary_john3_noverse.json": lambda: C.get_commentary("John", 3, None),
    "search_commentary_nicodemus.json": lambda: C.search_commentary("尼哥底母"),
    "search_commentary_none.json": lambda: C.search_commentary("zzqqxx"),
    "get_topic_study_love_count.json": lambda: C.get_topic_study("愛", count_only=True),
    "get_topic_study_love.json": lambda: C.get_topic_study("愛"),
    "get_topic_study_faith_torrey_en.json": lambda: C.get_topic_study("faith", source="torrey_en"),
    "get_topic_study_none.json": lambda: C.get_topic_study("zzqqxx"),
    "get_topic_study_none_count.json": lambda: C.get_topic_study("zzqqxx", count_only=True),
}
os.makedirs(OUT, exist_ok=True)
for name, fn in cases.items():
    res = asyncio.run(fn())
    with open(os.path.join(OUT, name), "wb") as f:
        f.write(json.dumps(res, ensure_ascii=False, indent=2).encode("utf-8"))
    print("wrote", name)
