@echo off
setlocal
cd /d "%~dp0"
echo TouchBridge Windows setup
echo.
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0Setup-TouchBridge-Windows.ps1"
set RESULT=%ERRORLEVEL%
echo.
if not "%RESULT%"=="0" echo Setup finished with errors. Review the messages above.
pause
exit /b %RESULT%
