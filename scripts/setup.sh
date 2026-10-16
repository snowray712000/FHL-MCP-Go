#!/bin/bash
# 首次設定 (Ubuntu / Linux)：把本服務登錄到 Claude Code / Codex (使用者層級，只需做一次)。
# 在終端機執行： ./setup.sh   (參數會原樣傳給 fhl-mcp setup，例如 ./setup.sh --client codex)
cd "$(dirname "$0")" || exit 1
if [ ! -x ./fhl-mcp ]; then
  echo "找不到可執行的 fhl-mcp，請確認本檔案與 fhl-mcp 在同一個資料夾，且有執行權限 (chmod +x fhl-mcp)。"
  exit 1
fi
exec ./fhl-mcp setup "$@"
