#!/usr/bin/env python3
"""
Draw rough PowerPoint-friendly diagrams for the Korean final report.

These figures are intentionally simple: boxes, arrows, and short labels.
They are meant to be used as visual drafts before redrawing them in slides.
"""

from __future__ import annotations

import argparse
from pathlib import Path

import matplotlib.pyplot as plt
from matplotlib.patches import FancyArrowPatch, FancyBboxPatch


BLUE = "#2563eb"
LIGHT_BLUE = "#dbeafe"
GREEN = "#16a34a"
LIGHT_GREEN = "#dcfce7"
ORANGE = "#ea580c"
LIGHT_ORANGE = "#ffedd5"
PURPLE = "#7c3aed"
LIGHT_PURPLE = "#ede9fe"
GRAY = "#475569"
LIGHT_GRAY = "#f1f5f9"
RED = "#dc2626"
LIGHT_RED = "#fee2e2"


def ensure_output_dir(path: Path) -> None:
    path.mkdir(parents=True, exist_ok=True)


def box(ax, x, y, w, h, title, subtitle="", fc=LIGHT_GRAY, ec=GRAY, fontsize=12):
    patch = FancyBboxPatch(
        (x, y),
        w,
        h,
        boxstyle="round,pad=0.018,rounding_size=0.025",
        linewidth=1.6,
        edgecolor=ec,
        facecolor=fc,
    )
    ax.add_patch(patch)
    ax.text(x + w / 2, y + h * 0.62, title, ha="center", va="center", fontsize=fontsize, weight="bold", color="#111827")
    if subtitle:
        ax.text(x + w / 2, y + h * 0.28, subtitle, ha="center", va="center", fontsize=fontsize - 2, color="#334155")
    return patch


def arrow(ax, x1, y1, x2, y2, color=GRAY, style="-|>", lw=1.8, rad=0.0):
    arr = FancyArrowPatch(
        (x1, y1),
        (x2, y2),
        arrowstyle=style,
        mutation_scale=16,
        linewidth=lw,
        color=color,
        connectionstyle=f"arc3,rad={rad}",
    )
    ax.add_patch(arr)
    return arr


def setup_canvas(figsize=(15, 6)):
    fig, ax = plt.subplots(figsize=figsize)
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax.axis("off")
    return fig, ax


def save(fig, path: Path):
    fig.savefig(path, dpi=220, bbox_inches="tight", facecolor="white")
    plt.close(fig)
    print(f"[OK] saved {path}")


def draw_figure1(output_dir: Path):
    fig, ax = setup_canvas((15.5, 6.8))
    ax.text(0.5, 0.95, "Figure 1 Draft. MigraGuard analysis and decision flow", ha="center", va="center", fontsize=17, weight="bold")

    # Inputs
    box(ax, 0.05, 0.63, 0.16, 0.18, "DDL Input", "ALTER TABLE\nCREATE INDEX", LIGHT_BLUE, BLUE)
    box(ax, 0.05, 0.31, 0.16, 0.18, "Runtime Metrics", "Table size / TPS\nP99 / Connections / Lag", LIGHT_GREEN, GREEN)

    # Middle processing
    box(ax, 0.31, 0.63, 0.17, 0.18, "Static Analysis", "Operation / Lock\nRewrite / Index", LIGHT_BLUE, BLUE)
    box(ax, 0.31, 0.31, 0.17, 0.18, "Metric Snapshot", "Current or forecast\nworkload state", LIGHT_GREEN, GREEN)

    box(ax, 0.57, 0.47, 0.18, 0.22, "5-Step Risk\nCalculation", "T_ddl → T_block\nC_peak → T_rec\nRiskScore", LIGHT_PURPLE, PURPLE)

    # Outputs
    box(ax, 0.82, 0.62, 0.13, 0.14, "Risk Level", "Safe / Warning\nDanger", LIGHT_ORANGE, ORANGE)
    box(ax, 0.82, 0.38, 0.13, 0.14, "Decision", "Run now / Delay\nSplit / Online change", LIGHT_RED, RED)

    arrow(ax, 0.21, 0.72, 0.31, 0.72, BLUE)
    arrow(ax, 0.21, 0.40, 0.31, 0.40, GREEN)
    arrow(ax, 0.48, 0.72, 0.57, 0.60, PURPLE)
    arrow(ax, 0.48, 0.40, 0.57, 0.55, PURPLE)
    arrow(ax, 0.75, 0.59, 0.82, 0.69, ORANGE)
    arrow(ax, 0.75, 0.52, 0.82, 0.45, RED)
    arrow(ax, 0.885, 0.62, 0.885, 0.52, GRAY)

    ax.text(0.50, 0.10, "Purpose: combine DDL static impact and pre-deployment workload state before execution", ha="center", fontsize=12, color="#334155")
    save(fig, output_dir / "01_migraguard_analysis_flow_draft.png")


def draw_figure2(output_dir: Path):
    fig, ax = setup_canvas((15.5, 5.8))
    ax.text(0.5, 0.93, "Figure 2 Draft. Experiment 1 benchmark procedure", ha="center", va="center", fontsize=17, weight="bold")

    steps = [
        ("1. Reset DB", "Clean benchmark DB\nbefore each DDL", LIGHT_GRAY, GRAY),
        ("2. Baseline", "Run pgbench\nwithout DDL", LIGHT_GREEN, GREEN),
        ("3. Pre-check", "Squawk + MigraGuard\nbefore execution", LIGHT_BLUE, BLUE),
        ("4. Run DDL", "Execute DDL while\nworkload is running", LIGHT_ORANGE, ORANGE),
        ("5. Measure", "Actual TPS / P99\nfailure / DDL time", LIGHT_PURPLE, PURPLE),
        ("6. Compare", "Pre-deployment\njudgment vs observed", LIGHT_RED, RED),
    ]

    x0 = 0.04
    y = 0.43
    w = 0.13
    gap = 0.03
    for i, (title, subtitle, fc, ec) in enumerate(steps):
        x = x0 + i * (w + gap)
        box(ax, x, y, w, 0.24, title, subtitle, fc, ec, fontsize=11)
        if i < len(steps) - 1:
            arrow(ax, x + w, y + 0.12, x + w + gap, y + 0.12, GRAY)

    # Inputs/outputs hints
    box(ax, 0.22, 0.13, 0.18, 0.12, "Input", "DDL-1~DDL-6\n100 / 1000 / 5000 TPS", "#eef2ff", "#4f46e5", fontsize=10)
    box(ax, 0.62, 0.13, 0.20, 0.12, "Output", "P99 ratio + observed label\nSafe / Warning / Danger", "#fff7ed", "#f97316", fontsize=10)
    arrow(ax, 0.31, 0.25, 0.31, 0.43, "#4f46e5", rad=0.0)
    arrow(ax, 0.72, 0.43, 0.72, 0.25, "#f97316", rad=0.0)

    ax.text(0.5, 0.04, "Question: Can MigraGuard warn before DDLs that actually increase latency or fail?", ha="center", fontsize=12, color="#334155")
    save(fig, output_dir / "02_experiment1_benchmark_procedure_draft.png")


def draw_figure5(output_dir: Path):
    fig, ax = setup_canvas((15.5, 6.3))
    ax.text(0.5, 0.94, "Figure 5 Draft. Forecast optimality validation procedure", ha="center", va="center", fontsize=17, weight="bold")

    top_steps = [
        ("1. Forecast", "Run MigraGuard\n24-hour forecast", LIGHT_BLUE, BLUE),
        ("2. Recommend", "Select best hour\nby RiskScore/level", LIGHT_BLUE, BLUE),
        ("3. Compress Hours", "Replay each hour's\nexpected TPS quickly", LIGHT_GREEN, GREEN),
        ("4. Execute DDL", "Run online/staged DDL\nunder each hourly load", LIGHT_ORANGE, ORANGE),
        ("5. Rank Hours", "P99 / TPS drop\nDDL time / failures", LIGHT_PURPLE, PURPLE),
        ("6. Validate", "Recommended hour\nvs actual rank", LIGHT_RED, RED),
    ]

    x0 = 0.04
    y = 0.53
    w = 0.13
    gap = 0.03
    for i, (title, subtitle, fc, ec) in enumerate(top_steps):
        x = x0 + i * (w + gap)
        box(ax, x, y, w, 0.23, title, subtitle, fc, ec, fontsize=10)
        if i < len(top_steps) - 1:
            arrow(ax, x + w, y + 0.115, x + w + gap, y + 0.115, GRAY)

    # Two lanes: MigraGuard prediction vs actual execution.
    ax.plot([0.05, 0.95], [0.43, 0.43], color="#cbd5e1", linewidth=1.4, linestyle="--")
    ax.text(0.06, 0.39, "Prediction lane", fontsize=11, weight="bold", color=BLUE)
    ax.text(0.06, 0.18, "Execution lane", fontsize=11, weight="bold", color=ORANGE)

    box(ax, 0.23, 0.32, 0.20, 0.11, "MigraGuard result", "Recommended: 02:00 Safe", "#eff6ff", BLUE, fontsize=10)
    box(ax, 0.57, 0.32, 0.20, 0.11, "Actual result", "Top hours / worst hours\nfrom benchmark rank", "#fff7ed", ORANGE, fontsize=10)
    arrow(ax, 0.43, 0.375, 0.57, 0.375, PURPLE)

    box(ax, 0.23, 0.10, 0.20, 0.12, "DDL Set", "CREATE INDEX CONCURRENTLY\nADD NOT VALID / VALIDATE", "#f8fafc", GRAY, fontsize=9)
    box(ax, 0.57, 0.10, 0.20, 0.12, "Judgment", "Exact optimum? No\nLow-risk candidate? Yes", "#fef2f2", RED, fontsize=9)
    arrow(ax, 0.33, 0.22, 0.33, 0.32, GRAY)
    arrow(ax, 0.67, 0.32, 0.67, 0.22, RED)

    ax.text(0.5, 0.035, "Question: Is MigraGuard's recommended hour actually the best hour under measured execution metrics?", ha="center", fontsize=12, color="#334155")
    save(fig, output_dir / "05_forecast_optimality_procedure_draft.png")


def parse_args():
    parser = argparse.ArgumentParser(description="Draw rough report diagrams for PowerPoint redraw.")
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path("experiments/reports/figures"),
        help="Directory where PNG diagrams will be written.",
    )
    return parser.parse_args()


def main():
    args = parse_args()
    ensure_output_dir(args.output_dir)
    draw_figure1(args.output_dir)
    draw_figure2(args.output_dir)
    draw_figure5(args.output_dir)


if __name__ == "__main__":
    main()
