import argparse
from pathlib import Path

import matplotlib.pyplot as plt
from matplotlib.colors import BoundaryNorm, ListedColormap
import matplotlib.patches as mpatches
import pandas as pd
import seaborn as sns


LEVEL_TO_VALUE = {"Safe": 0, "Warning": 1, "Danger": 2}
LEVEL_TO_MARKER = {"Safe": "S", "Warning": "W", "Danger": "D"}
LEVEL_COLORS = ["#2E8B57", "#F2B134", "#C83E4D"]


def parse_args():
    parser = argparse.ArgumentParser(
        description=(
            "Compare fixed DDL baseline labels with workload-aware MigraGuard "
            "outcomes for one capacity configuration."
        )
    )
    parser.add_argument("csv_path", type=Path, help="Path to matrix_results.csv")
    parser.add_argument(
        "--config",
        default="default",
        help="Config value to filter (default: default)",
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        help="Output directory (default: <csv-dir>/figures/default_level_comparison)",
    )
    return parser.parse_args()


def shorten_label(value, max_length):
    value = str(value)
    return value if len(value) <= max_length else f"{value[: max_length - 3]}..."


def load_matrices(csv_path, config_name):
    df = pd.read_csv(csv_path)
    required = {"Scenario", "Config", "DDL", "IntendedLevel", "ActualLevel"}
    missing = required.difference(df.columns)
    if missing:
        raise ValueError(f"Missing columns: {', '.join(sorted(missing))}")

    filtered = df[df["Config"].str.casefold() == config_name.casefold()].copy()
    if filtered.empty:
        raise ValueError(f"No rows found for config '{config_name}'")

    for column in ("IntendedLevel", "ActualLevel"):
        unknown = sorted(set(filtered[column].dropna()) - set(LEVEL_TO_VALUE))
        if unknown:
            raise ValueError(f"Unknown values in {column}: {unknown}")

    duplicates = filtered.duplicated(["DDL", "Scenario"], keep=False)
    if duplicates.any():
        raise ValueError("Duplicate DDL/Scenario rows exist after config filtering")

    ddl_order = sorted(filtered["DDL"].unique())
    scenario_order = sorted(filtered["Scenario"].unique())

    intended = filtered.pivot(index="DDL", columns="Scenario", values="IntendedLevel")
    actual = filtered.pivot(index="DDL", columns="Scenario", values="ActualLevel")
    intended = intended.reindex(index=ddl_order, columns=scenario_order)
    actual = actual.reindex(index=ddl_order, columns=scenario_order)

    if intended.isna().any().any() or actual.isna().any().any():
        raise ValueError("The selected config does not contain a complete DDL/scenario matrix")

    return filtered, intended, actual


def encode(matrix):
    return matrix.map(LEVEL_TO_VALUE.__getitem__).astype(int)


def annotate(matrix):
    return matrix.map(LEVEL_TO_MARKER.__getitem__)


def add_legend(fig, anchor_y=0.01):
    patches = [
        mpatches.Patch(color=color, label=level)
        for level, color in zip(("Safe", "Warning", "Danger"), LEVEL_COLORS)
    ]
    fig.legend(
        handles=patches,
        loc="lower center",
        ncol=3,
        frameon=False,
        bbox_to_anchor=(0.5, anchor_y),
    )


def draw_heatmap(ax, matrix, title, subtitle=None):
    cmap = ListedColormap(LEVEL_COLORS)
    norm = BoundaryNorm([-0.5, 0.5, 1.5, 2.5], cmap.N)
    encoded = encode(matrix)
    markers = annotate(matrix)

    sns.heatmap(
        encoded,
        ax=ax,
        cmap=cmap,
        norm=norm,
        cbar=False,
        annot=markers,
        fmt="",
        linewidths=0.45,
        linecolor="#FFFFFF",
        annot_kws={"fontsize": 7, "fontweight": "bold", "color": "#FFFFFF"},
    )
    ax.set_title(title, fontsize=14, fontweight="bold", pad=18)
    if subtitle:
        ax.text(
            0.5,
            1.01,
            subtitle,
            transform=ax.transAxes,
            ha="center",
            va="bottom",
            fontsize=9,
            color="#555555",
        )
    ax.set_xlabel("Traffic Scenario", fontsize=10, fontweight="bold", labelpad=10)
    ax.set_ylabel("DDL Case", fontsize=10, fontweight="bold", labelpad=10)
    ax.set_xticklabels(
        [shorten_label(value, 18) for value in matrix.columns],
        rotation=55,
        ha="right",
        fontsize=7,
    )
    ax.set_yticklabels(
        [shorten_label(value, 34) for value in matrix.index],
        rotation=0,
        fontsize=8,
    )


def save_single(matrix, output_path, title, subtitle):
    sns.set_theme(style="white", font_scale=0.9)
    fig, ax = plt.subplots(figsize=(18, 12))
    draw_heatmap(ax, matrix, title, subtitle)
    add_legend(fig, anchor_y=0.012)
    fig.subplots_adjust(left=0.23, right=0.98, top=0.88, bottom=0.24)
    fig.savefig(
        output_path,
        dpi=240,
        facecolor="white",
        bbox_inches="tight",
        pad_inches=0.2,
    )
    plt.close(fig)


def save_comparison(intended, actual, output_path, config_name):
    sns.set_theme(style="white", font_scale=0.85)
    fig, axes = plt.subplots(1, 2, figsize=(32, 13), sharey=True)
    draw_heatmap(
        axes[0],
        intended,
        "A. Fixed DDL Baseline Labels",
        "Labels assigned before workload-aware evaluation",
    )
    draw_heatmap(
        axes[1],
        actual,
        f"B. MigraGuard Outcomes ({config_name} config)",
        "Risk levels after applying each traffic scenario",
    )
    axes[1].set_ylabel("")
    add_legend(fig, anchor_y=0.018)
    fig.suptitle(
        "Fixed DDL Labels vs Workload-Aware Risk Outcomes",
        fontsize=19,
        fontweight="bold",
        y=0.965,
    )
    fig.subplots_adjust(left=0.13, right=0.99, top=0.87, bottom=0.25, wspace=0.08)
    fig.savefig(
        output_path,
        dpi=240,
        facecolor="white",
        bbox_inches="tight",
        pad_inches=0.2,
    )
    plt.close(fig)


def print_summary(filtered):
    comparison = filtered["IntendedLevel"] == filtered["ActualLevel"]
    print(f"Rows: {len(filtered)}")
    print(f"Baseline label matches: {int(comparison.sum())}")
    print(f"Changed by workload-aware evaluation: {int((~comparison).sum())}")
    print("\nActual level distribution:")
    print(filtered["ActualLevel"].value_counts().reindex(["Safe", "Warning", "Danger"], fill_value=0))


def main():
    args = parse_args()
    csv_path = args.csv_path.resolve()
    output_dir = (
        args.output_dir.resolve()
        if args.output_dir
        else csv_path.parent / "figures" / "default_level_comparison"
    )
    output_dir.mkdir(parents=True, exist_ok=True)

    filtered, intended, actual = load_matrices(csv_path, args.config)

    intended_path = output_dir / "01_fixed_ddl_baseline_labels.png"
    actual_path = output_dir / "02_workload_aware_actual_levels.png"
    comparison_path = output_dir / "03_baseline_vs_actual_comparison.png"

    save_single(
        intended,
        intended_path,
        "Fixed DDL Baseline Labels",
        "The same baseline label is repeated across scenarios before workload evaluation",
    )
    save_single(
        actual,
        actual_path,
        f"MigraGuard Workload-Aware Outcomes ({args.config} config)",
        "Each cell reflects the DDL and traffic scenario combination",
    )
    save_comparison(intended, actual, comparison_path, args.config)

    print_summary(filtered)
    print(f"\nSaved: {intended_path}")
    print(f"Saved: {actual_path}")
    print(f"Saved: {comparison_path}")


if __name__ == "__main__":
    main()
