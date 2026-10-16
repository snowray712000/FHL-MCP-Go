# Oracle for M2-f (audio): runs the Python reference's own tools/audio.py on recorded au.php fixtures.
# Assemble a scratch package: fhl_bible_mcp/{__init__,tools/__init__,utils/__init__,api/__init__}.py (empty), copy
# tools/audio.py, utils/booknames.py, utils/errors.py from the reference, and put audio_stub_endpoints.py at
# fhl_bible_mcp/api/endpoints.py. Then: python -I audio_oracle.py <fixtures-dir> <out-dir>
import asyncio, json, os, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from fhl_bible_mcp.api import endpoints
from fhl_bible_mcp.tools import audio as A

FIX, OUT = sys.argv[1], sys.argv[2]
TABLE = {
    (43, 3, 0): "au_john_3.json", (57, 1, 0): "au_phm_1.json", (1, 1, 0): "au_gen_1.json",
    (66, 22, 0): "au_rev_22.json", (43, 3, 1): "au_john_3_taiwanese.json", (43, 3, 7): "au_john_3_hebrew.json",
    (43, 99, 0): "au_john_99.json", (101, 1, 0): "au_tob_1.json",
}


def router(bid, chap, ver):
    with open(os.path.join(FIX, TABLE[(bid, chap, ver)]), encoding="utf-8") as f:
        return json.load(f)


endpoints.ROUTER = router
cases = {
    "get_audio_bible_john3.json": lambda: A.get_audio_bible("John", 3),
    "get_audio_bible_gen1.json": lambda: A.get_audio_bible("Gen", 1),
    "get_audio_bible_rev22.json": lambda: A.get_audio_bible("Rev", 22),
    "get_audio_bible_john3_taiwanese.json": lambda: A.get_audio_bible("John", 3, "taiwanese"),
    "get_audio_bible_john3_hebrew.json": lambda: A.get_audio_bible("約", 3, "hebrew"),
    "get_audio_bible_john99.json": lambda: A.get_audio_bible("John", 99),
    "get_audio_bible_tob1.json": lambda: A.get_audio_bible("101", 1),
    "get_audio_bible_phm1.json": lambda: A.get_audio_bible("門", 1),
    "list_audio_versions.json": lambda: A.list_audio_versions(),
}
os.makedirs(OUT, exist_ok=True)
for name, fn in cases.items():
    res = asyncio.run(fn())
    with open(os.path.join(OUT, name), "wb") as f:
        f.write(json.dumps(res, ensure_ascii=False, indent=2).encode("utf-8"))
    print("wrote", name)
