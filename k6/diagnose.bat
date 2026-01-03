@echo off
REM Diagnostic script to test backend before running k6

setlocal enabledelayedexpansion

set BASE_URL=%1
if "!BASE_URL!"=="" set BASE_URL=http://localhost:8080

echo.
echo 🔍 Backend Diagnostic Test
echo ================================
echo Testing URL: !BASE_URL!
echo.

REM Test 1: Health Check
echo 1️⃣  Testing Health Endpoint...
for /f %%i in ('curl -s -o /dev/null -w "%%{http_code}" "!BASE_URL!/health"') do set HTTP_CODE=%%i

if "!HTTP_CODE!"=="200" (
    echo    ✅ Health check passed (HTTP !HTTP_CODE!)
) else (
    echo    ❌ Health check failed (HTTP !HTTP_CODE!)
)
echo.

REM Test 2: Categories Endpoint
echo 2️⃣  Testing Categories Endpoint...
for /f %%i in ('curl -s -o /dev/null -w "%%{http_code}" "!BASE_URL!/api/v1/homeservices/categories"') do set HTTP_CODE=%%i

if "!HTTP_CODE!"=="200" (
    echo    ✅ Categories endpoint passed (HTTP !HTTP_CODE!)
) else if "!HTTP_CODE!"=="404" (
    echo    ⚠️  Categories endpoint returned 404 (endpoint may not exist)
) else if "!HTTP_CODE!"=="401" (
    echo    ⚠️  Categories endpoint requires auth (HTTP !HTTP_CODE!)
) else (
    echo    ❌ Categories endpoint failed (HTTP !HTTP_CODE!)
)
echo.

REM Test 3: Service Providers Endpoint
echo 3️⃣  Testing Service Providers Endpoint...
for /f %%i in ('curl -s -o /dev/null -w "%%{http_code}" "!BASE_URL!/api/v1/serviceproviders"') do set HTTP_CODE=%%i

if "!HTTP_CODE!"=="200" (
    echo    ✅ Providers endpoint passed (HTTP !HTTP_CODE!)
) else if "!HTTP_CODE!"=="404" (
    echo    ⚠️  Providers endpoint returned 404 (endpoint may not exist)
) else if "!HTTP_CODE!"=="401" (
    echo    ⚠️  Providers endpoint requires auth (HTTP !HTTP_CODE!)
) else (
    echo    ❌ Providers endpoint failed (HTTP !HTTP_CODE!)
)
echo.

echo ================================
echo ✅ Diagnostic complete!
echo.
echo If all tests show HTTP codes like 200/401/404, run:
echo   k6 run k6/basic-load-test.js
echo.
