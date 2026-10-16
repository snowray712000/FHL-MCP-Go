@echo off
chcp 65001 >nul
rem 雙擊執行首次設定：把本服務登錄到 Claude Code / Codex (使用者層級，只需做一次)。
cd /d "%~dp0"
if not exist "fhl-mcp.exe" (
  echo 找不到 fhl-mcp.exe，請確認本檔案與 fhl-mcp.exe 在同一個資料夾。
  pause
  exit /b 1
)
"%~dp0fhl-mcp.exe" setup
echo.
pause
