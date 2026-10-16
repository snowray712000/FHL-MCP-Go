#!/bin/bash
# 雙擊啟動 FHL MCP 服務；終端機視窗會保持開著，按 Ctrl+C 或另開視窗執行 ./fhl-mcp stop 即停止。
cd "$(dirname "$0")" || exit 1
if [ ! -x ./fhl-mcp ]; then
  echo "找不到可執行的 fhl-mcp，請確認本檔案與 fhl-mcp 在同一個資料夾，且有執行權限 (chmod +x fhl-mcp)。"
  read -r -n 1 -p "按任意鍵關閉…"
  exit 1
fi
echo "啟動 FHL MCP 服務 (停止請按 Ctrl+C，或另開視窗執行 ./fhl-mcp stop)"
echo
./fhl-mcp serve
code=$?
echo
echo "服務已結束 (結束碼 $code)。"
read -r -n 1 -p "按任意鍵關閉…"
