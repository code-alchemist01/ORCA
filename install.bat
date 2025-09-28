@echo off
REM Orca Install Script for Windows (requires admin privileges)

echo Installing Orca CLI...

REM Check if running as administrator
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo This script requires administrator privileges.
    echo Please run as administrator.
    pause
    exit /b 1
)

REM Build if binaries don't exist
if not exist "bin\orcacli.exe" (
    echo Building CLI first...
    call build.bat
    if %errorlevel% neq 0 (
        echo Build failed
        pause
        exit /b 1
    )
)

REM Copy CLI to system directory
echo Installing CLI to system directory...
copy bin\orcacli.exe C:\Windows\System32\orca.exe
if %errorlevel% neq 0 (
    echo Failed to install CLI
    pause
    exit /b 1
)

echo.
echo Orca CLI installed successfully!
echo You can now use 'orca' command from anywhere in the terminal.
echo.
echo Usage examples:
echo   orca --help
echo   orca version
echo   orca list
echo.
pause