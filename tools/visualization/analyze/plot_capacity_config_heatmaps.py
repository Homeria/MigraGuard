import argparse
from pathlib import Path

import matplotlib.patches as mpatches
import matplotlib.pyplot as plt
from matplotlib.colors import BoundaryNorm, ListedColormap
import pandas as pd
import seaborn as sns


LEVELS = ("Safe", "Warning", "Danger")
LEVEL_TO_VALUE = {level: index for index, level in enumerate(LEVELS)}
LEVEL_TO_MARKER = {"Safe": "S", "Warning": "W", "Danger": "D"}
LEVEL_COLORS = ["#2E8B57", "#F2B134", "#C83E4D"]
CONFIGS = (
    ("lowcapacity", "A. Low Capacity", "C_max=100, MuMax=1,200 TPS"),
    ("default", "B. Default", "C_max=500, MuMax=5,000 TPS"),
    ("highcapacity", "C. High Capacity", "C_max=2,000, MuMax=20,000 TPS"),
)


def parse_args():
    parser = argparse.ArgumentParser(
        description="Render Low, Default, and High Capacity DDL/scenario heatmaps."
    )
    parser.add_argument("csv_path", type=Path, help="Path to matrix_results.csv")
    parser.add_argument("--output", type=Path)
    return parser.parse_args()


def ddl_number(value):
    prefix = str(value).split("_", 1)[0]
    if not prefix.isdigit():
        raise ValueError(f"DDL name does not start with a number: {value}")
    return int(prefix)


def display_ddl_label(value):
    value = str(value)
    for group in ("safe", "warning", "danger"):
        marker = f"_{group}_"
        if marker in value:
            number, remainder = value.split(marker, 1)
            return f"{number}  {remainder}"
    if "_linter_warning_" in value:
        number, remainder = value.split("_linter_warning_", 1)
        return f"{number}  {remainder}"
    return value


def display_scenario_label(value):
    value = str(value)
    if len(value) > 3 and value[:2].isdigit() and value[2] == "_":
        return f"{value[:2]} {value[3:].replace('_', ' ')}"
    return value.replace("_", " ")


def load_data(csv_path):
    df = pd.read_csv(csv_path)
    required = {"Scenario", "Config", "DDL", "ActualLevel"}
    missing = required.difference(df.columns)
    if missing:
        raise ValueError(f"Missing columns: {', '.join(sorted(missing))}")

    df["DDLNumber"] = df["DDL"].map(ddl_number)
    df = df[df["DDLNumber"].between(22, 41)].copy()
    df = df[df["Config"].str.casefold().isin(name for name, _, _ in CONFIGS)]

    expected_rows = 20 * 20 * len(CONFIGS)
    if len(df) != expected_rows:
        raise ValueError(f"Expected {expected_rows} rows, found {len(df)}")
    if df["Scenario"].nunique() != 20 or df["DDLNumber"].nunique() != 20:
        raise ValueError("Expected exactly 20 DDLs and 20 scenarios")
    if df.duplicated(["Config", "DDL", "Scenario"]).any():
        raise ValueError("Duplicate config/DDL/scenario rows found")
    unknown = set(df["ActualLevel"].dropna()) - set(LEVELS)
    if unknown:
        raise ValueError(f"Unknown ActualLevel values: {sorted(unknown)}")
    return df


def matrix_for_config(df, config_name):
    subset = df[df["Config"].str.casefold() == config_name.casefold()].copy()
    ddl_order = (
        subset[["DDL", "DDLNumber"]]
        .drop_duplicates()
        .sort_values("DDLNumber")["DDL"]
        .tolist()
    )
    scenario_order = sorted(subset["Scenario"].unique())
    matrix = subset.pivot(index="DDL", columns="Scenario", values="ActualLevel")
    matrix = matrix.reindex(index=ddl_order, columns=scenario_order)
    if matrix.isna().any().any():
        raise ValueError(f"Incomplete matrix for config '{config_name}'")
    return matrix


def draw_heatmap(ax, matrix, title, subtitle, show_ylabels):
    cmap = ListedColormap(LEVEL_COLORS)
    norm = BoundaryNorm([-0.5, 0.5, 1.5, 2.5], cmap.N)
    encoded = matrix.map(LEVEL_TO_VALUE.__getitem__).astype(int)
    markers = matrix.map(LEVEL_TO_MARKER.__getitem__)
    sns.heatmap(
        encoded,
        ax=ax,
        cmap=cmap,
        norm=norm,
        cbar=False,
        annot=markers,
        fmt="",
        linewidths=0.4,
        linecolor="white",
        annot_kws={"fontsize": 6.5, "fontweight": "bold", "color": "white"},
    )
    ax.set_title(title, fontsize=15, fontweight="bold", pad=25)
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
    ax.set_xticklabels(
        [display_scenario_label(value) for value in matrix.columns],
        rotation=57,
        ha="right",
        fontsize=6.5,
    )
    if show_ylabels:
        ax.set_ylabel("DDL Case", fontsize=10, fontweight="bold")
        ax.set_yticklabels(
            [display_ddl_label(value) for value in matrix.index],
            rotation=0,
            fontsize=7.2,
        )
    else:
        ax.set_ylabel("")
        ax.tick_params(axis="y", left=False, labelleft=False)
    for boundary in (5, 10, 15):
        ax.axhline(boundary, color="#202020", linewidth=1.8)


def save_figure(df, output_path):
    sns.set_theme(style="white", font_scale=0.85)
    fig, axes = plt.subplots(1, 3, figsize=(44, 12), sharey=False)

    for index, (config_name, title, subtitle) in enumerate(CONFIGS):
        matrix = matrix_for_config(df, config_name)
        draw_heatmap(
            axes[index],
            matrix,
            title,
            subtitle,
            show_ylabels=index == 0,
        )

    handles = [
        mpatches.Patch(color=color, label=level)
        for level, color in zip(LEVELS, LEVEL_COLORS)
    ]
    fig.legend(
        handles=handles,
        loc="lower center",
        ncol=3,
        frameon=False,
        bbox_to_anchor=(0.5, 0.008),
    )
    fig.suptitle(
        "DDL x Traffic Scenario Risk Levels by Capacity Configuration",
        fontsize=20,
        fontweight="bold",
        y=0.975,
    )
    fig.text(
        0.5,
        0.94,
        "Same 20 DDLs and 20 workload scenarios; only server capacity settings differ",
        ha="center",
        fontsize=10,
        color="#555555",
    )
    fig.subplots_adjust(left=0.1, right=0.995, top=0.86, bottom=0.25, wspace=0.08)
    fig.savefig(output_path, dpi=240, facecolor="white", bbox_inches="tight")
    plt.close(fig)


def print_summary(df):
    for config_name, title, _ in CONFIGS:
        counts = (
            df[df["Config"].str.casefold() == config_name]
            ["ActualLevel"]
            .value_counts()
            .reindex(LEVELS, fill_value=0)
        )
        print(
            f"{title[3:]}: "
            + ", ".join(f"{level}={int(counts[level])}" for level in LEVELS)
        )


def main():
    args = parse_args()
    csv_path = args.csv_path.resolve()
    output_path = (
        args.output.resolve()
        if args.output
        else csv_path.parent / "figures" / "capacity_config_comparison.png"
    )
    output_path.parent.mkdir(parents=True, exist_ok=True)

    df = load_data(csv_path)
    save_figure(df, output_path)
    print_summary(df)
    print(f"Saved: {output_path}")


if __name__ == "__main__":
    main()
