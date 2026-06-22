import json
import math
import re
import sys
from pathlib import Path


def percentile(values, probability):
    if not values:
        return None
    values = sorted(values)
    index = (len(values) - 1) * probability
    lower = math.floor(index)
    upper = math.ceil(index)
    if lower == upper:
        return values[lower]
    return values[lower] + (values[upper] - values[lower]) * (index - lower)


def parse_summary(path):
    text = path.read_text(encoding="utf-8", errors="replace")
    patterns = {
        "transactions": r"number of transactions actually processed: (\d+)",
        "failed_transactions": r"number of failed transactions: (\d+)",
        "average_latency_ms": r"latency average = ([\d.]+) ms",
        "actual_tps": r"tps = ([\d.]+)",
    }
    result = {}
    for name, pattern in patterns.items():
        match = re.search(pattern, text)
        result[name] = float(match.group(1)) if match else None
    return result


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: summarize_pgbench.py <result-directory>")

    result_dir = Path(sys.argv[1])
    latency_ms = []
    for log_path in result_dir.glob("pgbench_log.*"):
        for line in log_path.read_text(encoding="utf-8", errors="replace").splitlines():
            fields = line.split()
            if len(fields) < 3 or not fields[2].isdigit():
                continue
            latency_ms.append(int(fields[2]) / 1000.0)

    metrics = parse_summary(result_dir / "pgbench_summary.txt")
    metrics.update(
        {
            "logged_transactions": len(latency_ms),
            "p50_latency_ms": percentile(latency_ms, 0.50),
            "p95_latency_ms": percentile(latency_ms, 0.95),
            "p99_latency_ms": percentile(latency_ms, 0.99),
            "max_latency_ms": max(latency_ms) if latency_ms else None,
        }
    )
    output = result_dir / "summary_metrics.json"
    output.write_text(json.dumps(metrics, indent=2), encoding="utf-8")
    print(f"Wrote {output}")


if __name__ == "__main__":
    main()
