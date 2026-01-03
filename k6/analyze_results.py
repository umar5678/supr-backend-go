import json
import sys
from pathlib import Path
from datetime import datetime
from collections import defaultdict

def analyze_k6_results(json_file):
    """Analyze k6 JSON output and provide insights"""
    
    try:
        with open(json_file, 'r') as f:
            data = json.load(f)
    except Exception as e:
        print(f"❌ Error reading file: {e}")
        return
    
    # Extract metrics
    samples = data.get('metrics', {})
    
    print("\n" + "="*60)
    print(f"📊 K6 Load Test Analysis")
    print(f"File: {json_file}")
    print("="*60 + "\n")
    
    # Parse metadata
    if 'metadata' in data:
        meta = data['metadata']
        print(f"🏷️  Test Metadata:")
        print(f"   Environment: {meta.get('env', {}).get('BASE_URL', 'Unknown')}")
        print(f"   Test Type: {meta.get('testType', 'Unknown')}")
        print()
    
    # Collect all metric samples
    metrics = defaultdict(list)
    if 'metrics' in data:
        for metric_name, metric_data in data['metrics'].items():
            if 'values' in metric_data:
                metrics[metric_name] = metric_data['values']
    
    # Display key metrics
    print("📈 Key Performance Indicators:")
    print("-" * 60)
    
    # HTTP Request Duration
    if 'http_req_duration' in samples:
        values = samples['http_req_duration'].get('values', {})
        print(f"\n⏱️  HTTP Request Duration:")
        if isinstance(values, dict):
            for stat, val in values.items():
                if stat in ['p(95)', 'p(99)', 'avg', 'min', 'max']:
                    status = "✅" if float(val) < 500 else "⚠️ " if float(val) < 1000 else "❌"
                    print(f"   {stat:8} {val:>10}  {status}")
        elif isinstance(values, (int, float)):
            print(f"   Value: {values} ms")
    
    # HTTP Requests
    if 'http_requests' in samples:
        req_count = samples['http_requests'].get('value', 0)
        print(f"\n📨 HTTP Requests:")
        print(f"   Total: {req_count}")
    
    # HTTP Request Failed
    if 'http_req_failed' in samples:
        failed = samples['http_req_failed'].get('value', 0)
        error_rate = samples['http_req_failed'].get('rate', 0)
        status = "✅" if error_rate < 0.01 else "⚠️ " if error_rate < 0.05 else "❌"
        print(f"\n❌ Error Rate:")
        print(f"   Rate: {error_rate*100:.2f}%  {status}")
        print(f"   Failed Requests: {failed}")
    
    # VUs
    print(f"\n👥 Virtual Users:")
    if 'vus' in samples:
        vus_val = samples['vus'].get('value', 0)
        print(f"   Current: {vus_val}")
    if 'vus_max' in samples:
        vus_max_val = samples['vus_max'].get('value', 0)
        print(f"   Max: {vus_max_val}")
    
    # Data transferred
    print(f"\n💾 Data Transfer:")
    if 'data_received' in samples:
        received = samples['data_received'].get('value', 0)
        print(f"   Received: {received/1024/1024:.2f} MB")
    if 'data_sent' in samples:
        sent = samples['data_sent'].get('value', 0)
        print(f"   Sent: {sent/1024/1024:.2f} MB")
    
    # Recommendations
    print("\n" + "="*60)
    print("💡 Recommendations:")
    print("="*60)
    
    recommendations = []
    
    # Check error rate
    if 'http_req_failed' in samples:
        error_rate = samples['http_req_failed'].get('rate', 0)
        if error_rate > 0.05:
            recommendations.append("⚠️  High error rate detected. Check backend logs.")
        elif error_rate > 0.01:
            recommendations.append("⚠️  Error rate above 1%. Investigate failed endpoints.")
    
    # Check response times
    if 'http_req_duration' in samples:
        values = samples['http_req_duration'].get('values', {})
        if isinstance(values, dict):
            p95 = values.get('p(95)', 0)
            if isinstance(p95, str):
                p95 = float(p95)
            if p95 > 1000:
                recommendations.append("⚠️  P95 latency > 1s. Consider optimization.")
            elif p95 > 500:
                recommendations.append("ℹ️  P95 latency acceptable but slightly elevated.")
            else:
                recommendations.append("✅ Latency is excellent!")
    
    # Check overall health
    if 'http_req_failed' in samples:
        error_rate = samples['http_req_failed'].get('rate', 0)
        if error_rate < 0.01 and 'http_req_duration' in samples:
            values = samples['http_req_duration'].get('values', {})
            if isinstance(values, dict):
                p95 = values.get('p(95)', 0)
                if isinstance(p95, str):
                    p95 = float(p95)
                if p95 < 500:
                    recommendations.append("✅ Load test passed! System is performing well.")
    
    if not recommendations:
        recommendations.append("ℹ️  Test completed successfully.")
    
    for rec in recommendations:
        print(f"   {rec}")
    
    print("\n" + "="*60)

def main():
    if len(sys.argv) < 2:
        print("Usage: python analyze_results.py <results.json>")
        print("\nExample:")
        print("  python analyze_results.py k6-results/basic-load-test_20240103_120000.json")
        sys.exit(1)
    
    json_file = sys.argv[1]
    
    if not Path(json_file).exists():
        print(f"❌ File not found: {json_file}")
        sys.exit(1)
    
    analyze_k6_results(json_file)

if __name__ == '__main__':
    main()
