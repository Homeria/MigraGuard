import argparse
import csv
import json
from pathlib import Path


def read_json(path):
    return json.loads(path.read_text(encoding="utf-8-sig"))


def read_prediction(path):
    if not path.exists() or path.stat().st_size == 0:
        return {}
    with path.open(encoding="utf-8-sig", newline="") as handle:
        rows = list(csv.DictReader(handle))
    return rows[-1] if rows else {}


def md_table(headers, rows):
    lines = []
    lines.append("| " + " | ".join(headers) + " |")
    lines.append("| " + " | ".join("---" for _ in headers) + " |")
    for row in rows:
        lines.append("| " + " | ".join(str(row.get(h, "")) for h in headers) + " |")
    return "\n".join(lines)


def fmt(value, digits=1):
    if value is None or value == "":
        return ""
    try:
        return f"{float(value):.{digits}f}"
    except (TypeError, ValueError):
        return str(value)


def classify_observed(row, baseline):
    failed = float(row.get("failed_transactions") or 0)
    ddl_exit = int(row.get("ddl_exit_code") or 0)
    p99 = float(row.get("p99_latency_ms") or 0)
    baseline_p99 = float(baseline.get("p99_latency_ms") or 0)
    ratio = p99 / baseline_p99 if baseline_p99 > 0 else 0

    if ddl_exit != 0 or failed > 0:
        return "Danger"
    if ratio >= 5.0 and p99 - baseline_p99 >= 50:
        return "Danger"
    if ratio >= 2.0 and p99 - baseline_p99 >= 20:
        return "Warning"
    return "Safe"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--tag", required=True)
    parser.add_argument("--results-dir", default="experiments/benchmark/results")
    parser.add_argument("--output-md", required=True)
    parser.add_argument("--output-csv", required=True)
    args = parser.parse_args()

    results_dir = Path(args.results_dir)
    run_dirs = sorted([p for p in results_dir.glob(args.tag + "*") if p.is_dir()])

    baselines = {}
    cases = []
    for run_dir in run_dirs:
        metadata_path = run_dir / "metadata.json"
        summary_path = run_dir / "summary_metrics.json"
        if not metadata_path.exists() or not summary_path.exists():
            continue
        meta = read_json(metadata_path)
        metrics = read_json(summary_path)
        prediction = read_prediction(run_dir / "migraguard_prediction.csv")
        row = {
            "run_name": meta.get("run_name"),
            "ddl": Path(meta.get("ddl", "")).name if meta.get("ddl") else "baseline",
            "target_tps": meta.get("target_tps"),
            "clients": meta.get("clients"),
            "threads": meta.get("threads"),
            "extra_orders": meta.get("extra_orders", 0),
            "actual_tps": metrics.get("actual_tps"),
            "avg_latency_ms": metrics.get("average_latency_ms"),
            "p99_latency_ms": metrics.get("p99_latency_ms"),
            "failed_transactions": metrics.get("failed_transactions"),
            "ddl_duration_ms": meta.get("ddl_duration_ms"),
            "ddl_exit_code": meta.get("ddl_exit_code"),
            "squawk_exit_code": meta.get("squawk_exit_code"),
            "migraguard_level": prediction.get("RiskLevel", ""),
            "migraguard_score": prediction.get("RiskScore", ""),
            "migraguard_c_peak": prediction.get("C_peak", ""),
            "migraguard_t_block": prediction.get("T_block", ""),
            "migraguard_base_tps": prediction.get("BaseTPS", ""),
        }
        if row["ddl"] == "baseline":
            baselines.setdefault(str(row["target_tps"]), []).append(row)
        else:
            cases.append(row)

    # Use the latest baseline for each TPS. The generated Markdown still lists
    # all baseline runs so outliers remain visible.
    latest_baseline = {
        tps: sorted(rows, key=lambda r: r["run_name"])[-1]
        for tps, rows in baselines.items()
    }

    for row in cases:
        baseline = latest_baseline.get(str(row["target_tps"]), {})
        row["baseline_p99_ms"] = baseline.get("p99_latency_ms", "")
        row["p99_ratio_vs_baseline"] = (
            float(row["p99_latency_ms"]) / float(baseline["p99_latency_ms"])
            if baseline.get("p99_latency_ms") else ""
        )
        row["observed_label"] = classify_observed(row, baseline) if baseline else ""

    csv_headers = [
        "run_name",
        "ddl",
        "target_tps",
        "clients",
        "extra_orders",
        "actual_tps",
        "p99_latency_ms",
        "baseline_p99_ms",
        "p99_ratio_vs_baseline",
        "failed_transactions",
        "ddl_duration_ms",
        "ddl_exit_code",
        "squawk_exit_code",
        "migraguard_level",
        "migraguard_score",
        "migraguard_c_peak",
        "observed_label",
    ]
    output_csv = Path(args.output_csv)
    output_csv.parent.mkdir(parents=True, exist_ok=True)
    with output_csv.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, fieldnames=csv_headers)
        writer.writeheader()
        for row in cases:
            writer.writerow({k: row.get(k, "") for k in csv_headers})

    baseline_rows = []
    for rows in baselines.values():
        for row in rows:
            baseline_rows.append(
                {
                    "Target TPS": row["target_tps"],
                    "Clients": row["clients"],
                    "Actual TPS": fmt(row["actual_tps"]),
                    "Avg Latency(ms)": fmt(row["avg_latency_ms"]),
                    "P99(ms)": fmt(row["p99_latency_ms"]),
                    "Run": row["run_name"],
                }
            )
    baseline_rows.sort(key=lambda r: (int(r["Target TPS"]), r["Run"]))

    case_rows = []
    for row in sorted(cases, key=lambda r: (str(r["ddl"]), int(r["target_tps"]), str(r["run_name"]))):
        case_rows.append(
            {
                "DDL": row["ddl"],
                "TPS": row["target_tps"],
                "Squawk": "Pass" if row.get("squawk_exit_code") == 0 else "Warning",
                "MigraGuard": f"{row.get('migraguard_level')} ({row.get('migraguard_score')})",
                "Actual TPS": fmt(row["actual_tps"]),
                "P99(ms)": fmt(row["p99_latency_ms"]),
                "Baseline P99(ms)": fmt(row.get("baseline_p99_ms")),
                "P99 Ratio": fmt(row.get("p99_ratio_vs_baseline"), 2),
                "Observed": row.get("observed_label"),
                "DDL ms": fmt(row.get("ddl_duration_ms")),
            }
        )

    md = []
    md.append(f"# 벤치마크 결과 요약 - {args.tag}")
    md.append("")
    md.append("## 기준선 실행 결과")
    md.append("")
    md.append(md_table(["Target TPS", "Clients", "Actual TPS", "Avg Latency(ms)", "P99(ms)", "Run"], baseline_rows))
    md.append("")
    md.append("## DDL 실행 결과")
    md.append("")
    md.append(md_table(["DDL", "TPS", "Squawk", "MigraGuard", "Actual TPS", "P99(ms)", "Baseline P99(ms)", "P99 Ratio", "Observed", "DDL ms"], case_rows))
    md.append("")
    md.append("## 해석 시 주의사항")
    md.append("")
    md.append("- `Observed`는 실측 결과에서 임시로 부여한 실험용 판정이며, 보편적인 장애 정답 라벨은 아니다.")
    md.append("- 판정에는 DDL 실행 실패, 실패 트랜잭션 수, 동일 Target TPS 기준선 대비 P99 응답 지연 증가를 사용하였다.")
    md.append("- MigraGuard는 설정된 커넥션 한계와 작업 부하 추정값을 사용한다. 반면 pgbench는 지정한 client 수만큼만 동시 요청을 만들기 때문에, client 수가 낮으면 실제 커넥션 고갈은 충분히 재현되지 않을 수 있다.")
    md.append("- 따라서 본 결과는 방향성과 실패 사례를 확인하기 위한 예비 비교 실험으로 해석해야 하며, 통계적으로 검증된 정확도 실험으로 해석하기에는 아직 한계가 있다.")

    output_md = Path(args.output_md)
    output_md.parent.mkdir(parents=True, exist_ok=True)
    output_md.write_text("\n".join(md) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
