#!/usr/bin/env python3
"""
Generate report figures for docs/kr/09_report/3 - 주요 추진내용.md.

Outputs:
  - Figure 3: P99 latency ratio vs baseline by DDL and TPS
  - Figure 4: Squawk / MigraGuard / observed result comparison heatmap
  - Figure 6: Forecast recommendation vs actual operation rank by hour
  - Figure 7: RiskScore sensitivity by capacity config
"""

from __future__ import annotations

import argparse
from pathlib import Path

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import seaborn as sns
from matplotlib.colors import ListedColormap, BoundaryNorm
from matplotlib.lines import Line2D


LEVEL_TO_VALUE = {
    "Pass": 0,
    "Safe": 0,
    "Warning": 1,
    "Danger": 2,
}

LEVEL_CMAP = ListedColormap(["#2ca25f", "#ffd166", "#d95f02"])
LEVEL_NORM = BoundaryNorm([-0.5, 0.5, 1.5, 2.5], LEVEL_CMAP.N)

DDL_LABELS = {
    "DDL-1": "DDL-1\nNullable column",
    "DDL-2": "DDL-2\nConcurrent index",
    "DDL-3": "DDL-3\nStandard index",
    "DDL-4": "DDL-4\nType change",
    "DDL-5": "DDL-5\nSet NOT NULL",
    "DDL-6": "DDL-6\nImmediate CHECK",
}

ONLINE_LABELS = {
    "ONLINE-1": "ONLINE-1\nCREATE INDEX\nCONCURRENTLY",
    "ONLINE-2": "ONLINE-2\nCREATE UNIQUE INDEX\nCONCURRENTLY",
    "ONLINE-3": "ONLINE-3\nADD CONSTRAINT\nNOT VALID",
    "ONLINE-4": "ONLINE-4\nVALIDATE\nCONSTRAINT",
}


def ensure_output_dir(path: Path) -> None:
    path.mkdir(parents=True, exist_ok=True)


def save_figure(fig: plt.Figure, path: Path) -> None:
    fig.savefig(path, dpi=220, bbox_inches="tight")
    plt.close(fig)
    print(f"[OK] saved {path}")


def load_experiment1(path: Path) -> pd.DataFrame:
    df = pd.read_csv(path)
    expected = {
        "ID",
        "TargetTPS",
        "Squawk",
        "MigraGuard",
        "Actual",
        "P99Ratio",
        "P99",
        "BaselineP99",
    }
    missing = expected - set(df.columns)
    if missing:
        raise ValueError(f"{path} is missing columns: {sorted(missing)}")

    df["TargetTPS"] = pd.to_numeric(df["TargetTPS"], errors="coerce").astype("Int64")
    df["P99Ratio"] = pd.to_numeric(df["P99Ratio"], errors="coerce")
    df["P99"] = pd.to_numeric(df["P99"], errors="coerce")
    df["BaselineP99"] = pd.to_numeric(df["BaselineP99"], errors="coerce")
    return df.sort_values(["ID", "TargetTPS"])


def plot_figure3(df: pd.DataFrame, output_path: Path) -> None:
    sns.set_theme(style="whitegrid", font_scale=0.95)

    fig, ax = plt.subplots(figsize=(11.5, 5.8))
    palette = {
        100: "#8ecae6",
        1000: "#219ebc",
        5000: "#023047",
    }

    plot_df = df.copy()
    plot_df["DDL"] = plot_df["ID"].map(DDL_LABELS).fillna(plot_df["ID"])

    sns.barplot(
        data=plot_df,
        x="DDL",
        y="P99Ratio",
        hue="TargetTPS",
        palette=palette,
        ax=ax,
        edgecolor="white",
        linewidth=0.8,
    )

    ax.axhline(1, color="#555555", linestyle="-", linewidth=1.0, alpha=0.8)
    ax.axhline(2, color="#f4a261", linestyle="--", linewidth=1.0, alpha=0.8)
    ax.axhline(10, color="#d62828", linestyle="--", linewidth=1.0, alpha=0.8)
    ax.set_yscale("log")
    ax.set_ylim(0.15, max(plot_df["P99Ratio"].max() * 2.2, 12))
    ax.set_xlabel("")
    ax.set_ylabel("P99 ratio vs baseline (log scale)")
    ax.set_title("Figure 3. P99 latency change by DDL and target TPS", pad=16, weight="bold")
    ax.legend(title="Target TPS", ncol=1, loc="upper left", bbox_to_anchor=(0.02, 0.98), frameon=True)

    for container in ax.containers:
        labels = []
        for bar in container:
            height = bar.get_height()
            if not np.isfinite(height) or height <= 0:
                labels.append("")
            elif height >= 10:
                labels.append(f"{height:.1f}x")
            elif height >= 1:
                labels.append(f"{height:.2f}x")
            else:
                labels.append(f"{height:.2f}x")
        ax.bar_label(container, labels=labels, fontsize=8, padding=2, rotation=90)

    ax.text(
        0.995,
        0.98,
        "Guide lines: 1x baseline / 2x warning / 10x danger",
        transform=ax.transAxes,
        ha="right",
        va="top",
        fontsize=9,
        color="#555555",
        bbox={"facecolor": "white", "edgecolor": "#dddddd", "alpha": 0.9},
    )
    fig.tight_layout()
    save_figure(fig, output_path)


def plot_figure4(df: pd.DataFrame, output_path: Path) -> None:
    sns.set_theme(style="white", font_scale=0.95)

    columns = []
    values_by_method = {
        "Squawk": [],
        "MigraGuard": [],
        "Observed": [],
    }
    annotations_by_method = {
        "Squawk": [],
        "MigraGuard": [],
        "Observed": [],
    }
    for _, row in df.sort_values(["ID", "TargetTPS"]).iterrows():
        columns.append(f"{row['ID']}\n{int(row['TargetTPS'])} TPS")
        for method, source_col in [
            ("Squawk", "Squawk"),
            ("MigraGuard", "MigraGuard"),
            ("Observed", "Actual"),
        ]:
            label = str(row[source_col])
            values_by_method[method].append(LEVEL_TO_VALUE.get(label, np.nan))
            annotations_by_method[method].append(label)

    matrix = pd.DataFrame(values_by_method, index=columns).T
    annot = pd.DataFrame(annotations_by_method, index=columns).T

    fig, ax = plt.subplots(figsize=(16.2, 4.2))
    sns.heatmap(
        matrix,
        annot=annot,
        fmt="",
        cmap=LEVEL_CMAP,
        norm=LEVEL_NORM,
        linewidths=0.8,
        linecolor="white",
        cbar=False,
        ax=ax,
    )
    ax.set_xlabel("")
    ax.set_ylabel("")
    ax.set_title("Figure 4. Static check, MigraGuard, and observed result", pad=16, weight="bold")
    ax.set_xticklabels(ax.get_xticklabels(), rotation=50, ha="right", rotation_mode="anchor")
    ax.tick_params(axis="y", rotation=0)

    legend_handles = [
        Line2D([0], [0], marker="s", color="w", label="Pass / Safe", markerfacecolor="#2ca25f", markersize=12),
        Line2D([0], [0], marker="s", color="w", label="Warning", markerfacecolor="#ffd166", markersize=12),
        Line2D([0], [0], marker="s", color="w", label="Danger", markerfacecolor="#d95f02", markersize=12),
    ]
    ax.legend(handles=legend_handles, loc="center left", bbox_to_anchor=(1.01, 0.5), ncol=1, frameon=False)
    fig.tight_layout(rect=(0, 0, 0.9, 1))
    save_figure(fig, output_path)


def load_forecast_ranking(path: Path) -> pd.DataFrame:
    df = pd.read_csv(path)
    expected = {
        "kind",
        "ddl_id",
        "hour",
        "target_tps",
        "forecast_is_best_hour",
        "latency_p99_ms",
        "p99_ratio_to_baseline",
        "ddl_duration_ms",
        "actual_operation_score",
        "actual_operation_rank",
        "is_operation_top3",
    }
    missing = expected - set(df.columns)
    if missing:
        raise ValueError(f"{path} is missing columns: {sorted(missing)}")

    df = df[df["kind"] == "ddl"].copy()
    df["hour_num"] = df["hour"].str.slice(0, 2).astype(int)
    for col in [
        "target_tps",
        "latency_p99_ms",
        "p99_ratio_to_baseline",
        "ddl_duration_ms",
        "actual_operation_score",
        "actual_operation_rank",
    ]:
        df[col] = pd.to_numeric(df[col], errors="coerce")
    df["forecast_is_best_hour"] = df["forecast_is_best_hour"].astype(str).str.lower().eq("true")
    df["is_operation_top3"] = df["is_operation_top3"].astype(str).str.lower().eq("true")
    return df.sort_values(["ddl_id", "hour_num"])


def load_config_sensitivity(path: Path) -> pd.DataFrame:
    df = pd.read_csv(path)
    expected = {
        "ID",
        "Type",
        "Config",
        "RiskScore",
        "RiskLevel",
        "CMax",
        "MuMax",
        "DiskIO",
    }
    missing = expected - set(df.columns)
    if missing:
        raise ValueError(f"{path} is missing columns: {sorted(missing)}")

    df["RiskScore"] = pd.to_numeric(df["RiskScore"], errors="coerce")
    config_order = ["Low", "Default", "High"]
    df["Config"] = pd.Categorical(df["Config"], categories=config_order, ordered=True)
    return df.sort_values(["ID", "Config"])


def plot_figure6(df: pd.DataFrame, output_path: Path) -> None:
    sns.set_theme(style="white", font_scale=0.9)

    ddl_order = sorted(df["ddl_id"].unique())
    hours = list(range(24))
    rank_matrix = (
        df.pivot_table(index="ddl_id", columns="hour_num", values="actual_operation_rank", aggfunc="first")
        .reindex(index=ddl_order, columns=hours)
    )
    rank_matrix.index = [ONLINE_LABELS.get(x, x) for x in rank_matrix.index]

    fig, ax = plt.subplots(figsize=(14.5, 4.8))
    sns.heatmap(
        rank_matrix,
        cmap="RdYlGn_r",
        vmin=1,
        vmax=24,
        linewidths=0.35,
        linecolor="white",
        cbar_kws={"label": "Actual operation rank (1 = best)"},
        ax=ax,
    )

    # Mark MigraGuard's recommended hour, actual top-3, and actual worst hour.
    y_positions = {ddl: i for i, ddl in enumerate(ddl_order)}
    for ddl in ddl_order:
        ddl_df = df[df["ddl_id"] == ddl]
        y = y_positions[ddl] + 0.5

        rec = ddl_df[ddl_df["forecast_is_best_hour"]]
        for _, row in rec.iterrows():
            ax.scatter(
                row["hour_num"] + 0.5,
                y,
                marker="*",
                s=230,
                color="#1d4ed8",
                edgecolor="white",
                linewidth=0.9,
                zorder=5,
            )

        top3 = ddl_df[ddl_df["is_operation_top3"]]
        for _, row in top3.iterrows():
            ax.text(
                row["hour_num"] + 0.5,
                y,
                f"{int(row['actual_operation_rank'])}",
                ha="center",
                va="center",
                fontsize=9,
                weight="bold",
                color="#073b4c",
                zorder=6,
            )

        worst = ddl_df.loc[ddl_df["actual_operation_rank"].idxmax()]
        ax.scatter(
            worst["hour_num"] + 0.5,
            y,
            marker="X",
            s=130,
            color="#7f1d1d",
            edgecolor="white",
            linewidth=0.8,
            zorder=5,
        )

    ax.set_xticklabels([f"{h:02d}:00" for h in hours], rotation=45, ha="right")
    ax.set_xlabel("Hour")
    ax.set_ylabel("")
    ax.set_title("Figure 6. MigraGuard recommendation vs actual best/worst hours", pad=16, weight="bold")

    legend_handles = [
        Line2D([0], [0], marker="*", color="w", label="MigraGuard recommended hour", markerfacecolor="#1d4ed8", markeredgecolor="white", markersize=15),
        Line2D([0], [0], marker="$1$", color="w", label="Actual top-3 rank label", markerfacecolor="#073b4c", markeredgecolor="#073b4c", markersize=11),
        Line2D([0], [0], marker="X", color="w", label="Actual worst hour", markerfacecolor="#7f1d1d", markeredgecolor="white", markersize=10),
    ]
    ax.legend(handles=legend_handles, loc="upper center", bbox_to_anchor=(0.5, -0.22), ncol=3, frameon=False)
    fig.tight_layout()
    save_figure(fig, output_path)


def plot_figure7(df: pd.DataFrame, output_path: Path) -> None:
    sns.set_theme(style="whitegrid", font_scale=0.95)

    plot_df = df.copy()
    plot_df["DDL"] = plot_df["ID"].map(DDL_LABELS).fillna(plot_df["ID"])

    fig, ax = plt.subplots(figsize=(12.5, 5.8))
    palette = {
        "Low": "#dc2626",
        "Default": "#f59e0b",
        "High": "#16a34a",
    }
    sns.barplot(
        data=plot_df,
        x="DDL",
        y="RiskScore",
        hue="Config",
        palette=palette,
        ax=ax,
        edgecolor="white",
        linewidth=0.8,
    )

    ax.set_yscale("log")
    ax.set_ylim(1, max(plot_df["RiskScore"].max() * 1.6, 100))
    ax.axhline(50, color="#f59e0b", linestyle="--", linewidth=1.2, alpha=0.85)
    ax.axhline(80, color="#dc2626", linestyle="--", linewidth=1.2, alpha=0.85)
    ax.text(
        0.995,
        0.96,
        "Guide lines: Safe < 50 / Danger >= 80",
        transform=ax.transAxes,
        ha="right",
        va="top",
        fontsize=9,
        color="#555555",
        bbox={"facecolor": "white", "edgecolor": "#dddddd", "alpha": 0.9},
    )

    for container in ax.containers:
        labels = []
        for bar in container:
            height = bar.get_height()
            if not np.isfinite(height) or height <= 0:
                labels.append("")
            elif height >= 1000:
                labels.append(f"{height:,.0f}")
            elif height >= 100:
                labels.append(f"{height:.0f}")
            else:
                labels.append(f"{height:.1f}")
        ax.bar_label(container, labels=labels, fontsize=8, padding=2, rotation=90)

    ax.set_xlabel("")
    ax.set_ylabel("RiskScore (log scale)")
    ax.set_title("Figure 7. RiskScore change by capacity config", pad=16, weight="bold")
    ax.legend(title="Config", loc="upper left", frameon=True)
    fig.tight_layout()
    save_figure(fig, output_path)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate benchmark figures for the Korean final report.")
    parser.add_argument(
        "--experiment1-csv",
        type=Path,
        default=Path("experiments/benchmark/results/expand_20260620_summary.csv"),
        help="CSV for DDL-by-DDL benchmark result comparison.",
    )
    parser.add_argument(
        "--forecast-ranking-csv",
        type=Path,
        default=Path("experiments/benchmark/results/forecast_hourly_optimality_20260621/hourly_optimality_operation_ranking.csv"),
        help="CSV for hourly forecast optimality ranking.",
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path("experiments/reports/figures"),
        help="Directory where PNG figures will be written.",
    )
    parser.add_argument(
        "--config-csv",
        type=Path,
        default=Path("experiments/benchmark/results/config_sensitivity_20260621/config_sensitivity_summary.csv"),
        help="CSV for capacity config sensitivity result.",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    ensure_output_dir(args.output_dir)

    exp1 = load_experiment1(args.experiment1_csv)
    if len(exp1) != 18:
        print(f"[WARN] expected 18 experiment-1 rows, got {len(exp1)} from {args.experiment1_csv}")

    forecast = load_forecast_ranking(args.forecast_ranking_csv)
    expected_forecast_rows = 4 * 24
    if len(forecast) != expected_forecast_rows:
        print(f"[WARN] expected {expected_forecast_rows} forecast rows, got {len(forecast)} from {args.forecast_ranking_csv}")

    config = load_config_sensitivity(args.config_csv)
    if len(config) != 18:
        print(f"[WARN] expected 18 config sensitivity rows, got {len(config)} from {args.config_csv}")

    plot_figure3(exp1, args.output_dir / "03_p99_latency_change_by_ddl.png")
    plot_figure4(exp1, args.output_dir / "04_squawk_migraguard_observed_heatmap.png")
    plot_figure6(forecast, args.output_dir / "06_forecast_recommendation_actual_rank.png")
    plot_figure7(config, args.output_dir / "07_config_riskscore_sensitivity.png")


if __name__ == "__main__":
    main()
