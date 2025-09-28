@echo off
REM Orca Build Script for Windows

echo Building Orca Orchestrator...

REM Create bin directory if it doesn't exist
if not exist "bin" mkdir bin

REM Build orchestrator
echo Building orchestrator...
go build -o bin\orchestrator.exe .\cmd\orchestrator
if %errorlevel% neq 0 (
    echo Failed to build orchestrator
    exit /b 1
)

REM Build CLI
echo Building CLI...
go build -o bin\orcacli.exe .\cmd\orcacli
if %errorlevel% neq 0 (
    echo Failed to build CLI
    exit /b 1
)

echo Build completed successfully!
echo.
echo Binaries created:
echo   - bin\orchestrator.exe
echo   - bin\orcacli.exe
echo.
echo To install CLI globally (requires admin privileges):
echo   copy bin\orcacli.exe C:\Windows\System32\orca.exe
echo.
echo To run orchestrator:
echo   .\bin\orchestrator.exe
echo.
echo To use CLI:
echo   .\bin\orcacli.exe --help