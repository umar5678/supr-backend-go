#!/bin/bash
# Quick fix implementation script - copy paste this!

set -e

echo "🔧 K6 FIX IMPLEMENTATION"
echo "================================"
echo ""
echo "Step 1: Pull latest code..."
git pull origin main
echo "✅ Code pulled"
echo ""

echo "Step 2: Rebuild backend..."
go build -o bin/api ./cmd/api
echo "✅ Backend rebuilt"
echo ""

echo "Step 3: Stop old backend..."
pkill -f "go run" || true
pkill -f "./bin/api" || true
sleep 1
echo "✅ Old backend stopped"
echo ""

echo "Step 4: Start new backend..."
nohup ./bin/api > api.log 2>&1 &
sleep 2
echo "✅ New backend started"
echo ""

echo "Step 5: Test health endpoint..."
RESPONSE=$(curl -s http://localhost:8080/health)
echo "Response: $RESPONSE"
echo ""

if [ -z "$RESPONSE" ]; then
    echo "❌ Backend not responding! Check logs:"
    tail -20 api.log
else
    echo "✅ Backend is responding!"
    echo ""
    echo "Step 6: Run k6 test..."
    cd k6
    ./run-k6-tests.sh basic
fi
