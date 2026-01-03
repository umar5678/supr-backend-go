@echo off
REM Helper script to run k6 tests on Windows PowerShell

setlocal enabledelayedexpansion

set BASE_URL=http://localhost:8080
set AUTH_TOKEN=
set RESULTS_DIR=.\k6-results

REM Create results directory
if not exist "%RESULTS_DIR%" mkdir "%RESULTS_DIR%"

REM Get timestamp
for /f "tokens=2-4 delims=/ " %%a in ('date /t') do (set mydate=%%c%%a%%b)
for /f "tokens=1-2 delims=/:" %%a in ('time /t') do (set mytime=%%a%%b)
set TIMESTAMP=%mydate%_%mytime%

echo.
echo 🚀 k6 Load Testing Helper for Windows
echo =====================================
echo Base URL: %BASE_URL%
echo Results Directory: %RESULTS_DIR%
echo.

if "%1"=="" (
    echo Usage: run-k6-tests.bat [command]
    echo.
    echo Commands:
    echo   basic       - Run basic load test (recommended first^)
    echo   realistic   - Run realistic user journey test
    echo   spike       - Run spike test
    echo   stress      - Run stress test (will likely crash API!^)
    echo   endurance   - Run 30-minute endurance test
    echo   ramp        - Run ramp-up test
    echo.
    echo Examples:
    echo   run-k6-tests.bat basic
    echo   set BASE_URL=http://api.example.com && run-k6-tests.bat realistic
    echo.
    exit /b 1
)

if "%1"=="basic" (
    echo ▶️  Running Basic Load Test (50-100 VUs, 9 min^)
    if "%AUTH_TOKEN%"=="" (
        k6 run -e BASE_URL="%BASE_URL%" -o json="%RESULTS_DIR%\basic-load-test_%TIMESTAMP%.json" basic-load-test.js
    ) else (
        k6 run -e BASE_URL="%BASE_URL%" -e AUTH_TOKEN="%AUTH_TOKEN%" -o json="%RESULTS_DIR%\basic-load-test_%TIMESTAMP%.json" basic-load-test.js
    )
    echo ✅ Test completed!
    goto :end
)

if "%1"=="realistic" (
    echo ▶️  Running Realistic User Journey Test (50 VUs, 10 min^)
    if "%AUTH_TOKEN%"=="" (
        k6 run -e BASE_URL="%BASE_URL%" -o json="%RESULTS_DIR%\realistic-journey_%TIMESTAMP%.json" realistic-user-journey.js
    ) else (
        k6 run -e BASE_URL="%BASE_URL%" -e AUTH_TOKEN="%AUTH_TOKEN%" -o json="%RESULTS_DIR%\realistic-journey_%TIMESTAMP%.json" realistic-user-journey.js
    )
    echo ✅ Test completed!
    goto :end
)

if "%1"=="spike" (
    echo ▶️  Running Spike Test (30-^>200-^>150 VUs, 8 min^)
    if "%AUTH_TOKEN%"=="" (
        k6 run -e BASE_URL="%BASE_URL%" -o json="%RESULTS_DIR%\spike-test_%TIMESTAMP%.json" spike-test.js
    ) else (
        k6 run -e BASE_URL="%BASE_URL%" -e AUTH_TOKEN="%AUTH_TOKEN%" -o json="%RESULTS_DIR%\spike-test_%TIMESTAMP%.json" spike-test.js
    )
    echo ✅ Test completed!
    goto :end
)

if "%1"=="stress" (
    echo.
    echo ⚠️  WARNING: Stress test will attempt to crash your API!
    echo    This will help identify maximum capacity.
    echo.
    set /p confirm="Continue? (y/n): "
    if /i "%confirm%"=="y" (
        echo ▶️  Running Stress Test (100-^>500 VUs, 30 min^)
        if "%AUTH_TOKEN%"=="" (
            k6 run -e BASE_URL="%BASE_URL%" -o json="%RESULTS_DIR%\stress-test_%TIMESTAMP%.json" stress-test.js
        ) else (
            k6 run -e BASE_URL="%BASE_URL%" -e AUTH_TOKEN="%AUTH_TOKEN%" -o json="%RESULTS_DIR%\stress-test_%TIMESTAMP%.json" stress-test.js
        )
        echo ✅ Test completed!
    ) else (
        echo Aborted.
    )
    goto :end
)

if "%1"=="endurance" (
    echo ⏱️  Starting 30+ minute endurance test...
    if "%AUTH_TOKEN%"=="" (
        k6 run -e BASE_URL="%BASE_URL%" -o json="%RESULTS_DIR%\endurance-test_%TIMESTAMP%.json" endurance-test.js
    ) else (
        k6 run -e BASE_URL="%BASE_URL%" -e AUTH_TOKEN="%AUTH_TOKEN%" -o json="%RESULTS_DIR%\endurance-test_%TIMESTAMP%.json" endurance-test.js
    )
    echo ✅ Test completed!
    goto :end
)

if "%1"=="ramp" (
    echo ▶️  Running Ramp-Up Test (10-^>100 VUs, 6 min^)
    if "%AUTH_TOKEN%"=="" (
        k6 run -e BASE_URL="%BASE_URL%" -o json="%RESULTS_DIR%\ramp-up-test_%TIMESTAMP%.json" ramp-up-test.js
    ) else (
        k6 run -e BASE_URL="%BASE_URL%" -e AUTH_TOKEN="%AUTH_TOKEN%" -o json="%RESULTS_DIR%\ramp-up-test_%TIMESTAMP%.json" ramp-up-test.js
    )
    echo ✅ Test completed!
    goto :end
)

echo Unknown command: %1
echo Use: basic, realistic, spike, stress, endurance, or ramp

:end
echo.
echo 📊 Results saved to: %RESULTS_DIR%
echo.
dir "%RESULTS_DIR%" /o:-d
echo.
