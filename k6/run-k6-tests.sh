#!/bin/bash
# Helper script to run various k6 tests with common options

set -e

BASE_URL="${BASE_URL:-https://api.pittapizzahusrev.be/go}"
AUTH_TOKEN="${AUTH_TOKEN:-}"
RESULTS_DIR="./k6-results"

# Create results directory
mkdir -p "$RESULTS_DIR"

# Get timestamp for unique filenames
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "🚀 k6 Load Testing Helper"
echo "================================"
echo "Base URL: $BASE_URL"
echo "Results Directory: $RESULTS_DIR"
echo ""

# Function to run a test
run_test() {
    local test_name=$1
    local test_file=$2
    local description=$3

    echo "▶️  Running: $description"
    echo "   File: $test_file"
    echo ""

    if [ -z "$AUTH_TOKEN" ]; then
        k6 run \
            -e BASE_URL="$BASE_URL" \
            -o json="$RESULTS_DIR/${test_name}_${TIMESTAMP}.json" \
            "$test_file"
    else
        k6 run \
            -e BASE_URL="$BASE_URL" \
            -e AUTH_TOKEN="$AUTH_TOKEN" \
            -o json="$RESULTS_DIR/${test_name}_${TIMESTAMP}.json" \
            "$test_file"
    fi

    echo ""
    echo "✅ Test completed: $test_name"
    echo ""
}

# Show usage
if [ $# -eq 0 ]; then
    echo "Usage: ./run-k6-tests.sh [command]"
    echo ""
    echo "Commands:"
    echo "  diagnose    - Run endpoint diagnostics (START HERE!)"
    echo "  basic       - Run basic load test (recommended first)"
    echo "  realistic   - Run realistic user journey test"
    echo "  spike       - Run spike test"
    echo "  stress      - Run stress test (will likely crash API!)"
    echo "  endurance   - Run 30-minute endurance test"
    echo "  ramp        - Run ramp-up test"
    echo "  all         - Run all tests sequentially"
    echo ""
    echo "Examples:"
    echo "  ./run-k6-tests.sh basic"
    echo "  BASE_URL=http://api.example.com ./run-k6-tests.sh realistic"
    echo "  AUTH_TOKEN=your_token BASE_URL=http://localhost:8080 ./run-k6-tests.sh all"
    echo ""
    exit 1
fi

case "$1" in
    diagnose)
        echo "🔍 Running endpoint diagnostics..."
        run_test "diagnose" "diagnose-endpoints.js" "Endpoint Diagnostics (1 VU)"
        ;;
    basic)
        run_test "basic-load-test" "basic-load-test.js" "Basic Load Test (50-100 VUs, 9 min)"
        ;;
    realistic)
        run_test "realistic-journey" "realistic-user-journey.js" "Realistic User Journey (50 VUs, 10 min)"
        ;;
    spike)
        run_test "spike-test" "spike-test.js" "Spike Test (30->200->150 VUs, 8 min)"
        ;;
    stress)
        echo "⚠️  WARNING: Stress test will attempt to crash your API!"
        echo "   This will help identify maximum capacity."
        read -p "Continue? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            run_test "stress-test" "stress-test.js" "Stress Test (100->500 VUs, 30 min)"
        else
            echo "Aborted."
        fi
        ;;
    endurance)
        echo "⏱️  Starting 30+ minute endurance test..."
        run_test "endurance-test" "endurance-test.js" "Endurance Test (50 VUs, 40 min)"
        ;;
    ramp)
        run_test "ramp-up-test" "ramp-up-test.js" "Ramp-Up Test (10->100 VUs, 6 min)"
        ;;
    all)
        echo "🔄 Running all tests sequentially..."
        run_test "1-basic" "basic-load-test.js" "1. Basic Load Test"
        run_test "2-realistic" "realistic-user-journey.js" "2. Realistic User Journey"
        run_test "3-ramp" "ramp-up-test.js" "3. Ramp-Up Test"
        run_test "4-spike" "spike-test.js" "4. Spike Test"
        echo "⏭️  Skipping endurance test (too long). Run separately: ./run-k6-tests.sh endurance"
        ;;
    *)
        echo "Unknown command: $1"
        echo "Use: diagnose, basic, realistic, spike, stress, endurance, ramp, or all"
        exit 1
        ;;
esac

echo ""
echo "📊 Results saved to: $RESULTS_DIR/"
echo "📈 View latest results:"
ls -lh "$RESULTS_DIR" | tail -5
