# M0 來源 manifest（Python 原專案快照）

- 盤點日期：2026-10-16
- 來源：`D:\coding\信望愛\FHL-MCP-Server`（唯讀，未修改）
- Git HEAD：`03e9cb067bea5800934c6f281c05b8ecdac2e4f6`（2025-12-30 01:23:06 +0800）
- `pyproject.toml` 版本：`0.1.2`
- 清單：[source-manifest.tsv](source-manifest.tsv)（欄位：`path` / `git_state` / `bytes` / `sha256`，UTF-8、LF、tab 分隔）

## 範圍

收錄 `git ls-files --cached --others --exclude-standard` 的全部檔案，共 **254** 個：已追蹤未修改 237、已修改 7、未追蹤 10。
被 `.gitignore` 排除者（`.venv/`、`.cache/`、`.pytest_cache/`、`.coverage`、`__pycache__/`、`*.egg-info/`）不收錄。
清單檔名比對未發現 `.env`、金鑰、憑證類檔案。

### 已修改（7）

`README.md`、`pyproject.toml`、`src/fhl_bible_mcp/api/endpoints.py`、`src/fhl_bible_mcp/http_server.py`、
`src/fhl_bible_mcp/server.py`、`src/fhl_bible_mcp/smithery_server.py`、`src/fhl_bible_mcp/tools/__init__.py`

### 未追蹤（10，HEAD 不含）

- `src/fhl_bible_mcp/tools/parsing_reference.py`、`src/fhl_bible_mcp/utils/parsing_format.py`（原文分析實作）
- `tests/fixtures/parsing_reference_james_1_1.md`、`tests/test_tools/test_parsing_reference.py`、`tests/test_utils/test_parsing_format.py`
- `doc-ai/index.md`、`doc-ai/2026/s3/0820a_parsing-reference-design.md`、`doc-ai/2026/s3/0820b_parsing-reference-implementation.md`
- `start-fhl-mcp.bat`、`uv.lock`

## 已核對的事實

- `http_server.py` 有 29 個 `@mcp.tool`；函式名稱與計畫書 §4 的 29 項逐一相同。`server.py`、`smithery_server.py` 的 `@mcp.tool` 為 0。
- 此快照是「工作目錄狀態」，不等於 HEAD；對照 HEAD 需另看 `git diff`（7 個檔案，+143/−4 行）。

## 未涵蓋 / 後續

- 尚未執行原版測試、尚未啟動原版匯出 `tools/list` schema（M0 下一步）。
- 原專案若再修改，需重跑本清單並比對 SHA-256 以確認基準漂移。
- Prompts 19 項、Resources 7 類尚未盤點。
