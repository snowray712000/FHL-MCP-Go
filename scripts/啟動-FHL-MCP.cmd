@echo off
chcp 65001 >nul
rem 雙擊啟動 FHL MCP 服務；視窗會保持開著，關閉視窗或按 Ctrl+C 即停止。
cd /d "%~dp0"
if not exist "fhl-mcp.exe" (
  echo 找不到 fhl-mcp.exe，請確認本檔案與 fhl-mcp.exe 在同一個資料夾。
  pause
  exit /b 1
)
echo 啟動 FHL MCP 服務 (停止請按 Ctrl+C，或另開視窗執行 fhl-mcp.exe stop)
echo.
"%~dp0fhl-mcp.exe" serve
echo.
echo 服務已結束 (結束碼 %errorlevel%)。
pause
