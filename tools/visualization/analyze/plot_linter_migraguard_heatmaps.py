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
PASS_GROUPS = {
    "Safe": range(22, 27),
    "Warning": range(27, 32),
    "Danger": range(32, 37),
}
LINTER_WARNING_RULES = {
    37: "non-concurrent index",
    38: "column type change",
    39: "direct NOT NULL",
    40: "immediate constraint validation",
    41: "column rename",
}


def parse_args():
    parser = argparse.ArgumentParser(
        description="Render the two Squawk/MigraGuard comparison heatmaps."
    )
    parser.add_argument("csv_path", type=Path, help="Path to matrix_results.csv")
    parser.add_argument("--config", default="default")
    parser.add_argument("--output-dir", type=Path)
    return parser.parse_args()


def ddl_number(value):
    prefix = str(value).split("_", 1)[0]
    if not prefix.isdigit():
        raise ValueError(f"DDL name does not start with a number: {value}")
    return int(prefix)


def short_ddl_label(value, max_length=43):
    value = str(value)
    return value if len(value) <= max_length else f"{value[: max_length - 3]}..."


def display_ddl_label(value):
    value = str(value)
    parts = value.split("_", 2)
    if len(parts) == 3 and parts[1] in {"safe", "warning", "danger"}:
        return f"{parts[0]}  {parts[2]}"
    if value.startswith(tuple(f"{number:03d}_linter_warning_" for number in range(37, 42))):
        number, remainder = value.split("_linter_warning_", 1)
        return f"{number}  {remainder}"
    return value


def short_scenario_label(value):
    value = str(value)
    if len(value) > 3 and value[:2].isdigit() and value[2] == "_":
        return f"{value[:2]} {value[3:].replace('_', ' ')}"
    return value.replace("_", " ")


def load_data(csv_path, config_name):
    df = pd.read_csv(csv_path)
    required = {"Scenario", "Config", "DDL", "IntendedLevel", "ActualLevel"}
    missing = required.difference(df.columns)
    if missing:
        raise ValueError(f"Missing columns: {', '.join(sorted(missing))}")

    df = df[df["Config"].str.casefold() == config_name.casefold()].copy()
    df["DDLNumber"] = df["DDL"].map(ddl_number)
    df = df[df["DDLNumber"].between(22, 41)].copy()
    if df.empty:
        raise ValueError("No DDL 022-041 rows found for the selected config")

    expected_numbers = set(range(22, 42))
    actual_numbers = set(df["DDLNumber"].unique())
    if actual_numbers != expected_numbers:
        missing_numbers = sorted(expected_numbers - actual_numbers)
        raise ValueError(f"Incomplete DDL set; missing: {missing_numbers}")

    if set(df["ActualLevel"].dropna()) - set(LEVELS):
        raise ValueError("ActualLevel contains values outside Safe/Warning/Danger")

    scenario_count = df["Scenario"].nunique()
    if scenario_count != 20:
        raise ValueError(f"Expected 20 scenarios, found {scenario_count}")
    if df.duplicated(["DDL", "Scenario"]).any():
        raise ValueError("Duplicate DDL/scenario rows found")

    counts = df[["DDLNumber", "IntendedLevel"]].drop_duplicates()[
        "IntendedLevel"
    ].value_counts()
    expected_counts = {"Safe": 5, "Warning": 5, "Danger": 5, "LinterWarning": 5}
    for level, expected in expected_counts.items():
        if counts.get(level, 0) != expected:
            raise ValueError(
                f"Expected {expected} {level} DDLs, found {counts.get(level, 0)}"
            )
    return df


def pivot(df, value_column, ddl_numbers):
    subset = df[df["DDLNumber"].isin(ddl_numbers)].copy()
    ddl_order = (
        subset[["DDL", "DDLNumber"]]
        .drop_duplicates()
        .sort_values("DDLNumber")["DDL"]
        .tolist()
    )
    scenario_order = sorted(subset["Scenario"].unique())
    matrix = subset.pivot(index="DDL", columns="Scenario", values=value_column)
    return matrix.reindex(index=ddl_order, columns=scenario_order)


def draw_level_heatmap(ax, matrix, title, show_ylabels=True, ylabels=None):
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
        linewidths=0.45,
        linecolor="white",
        annot_kws={"fontsize": 7, "fontweight": "bold", "color": "white"},
    )
    ax.set_title(title, fontsize=14, fontweight="bold", pad=14)
    ax.set_xlabel("Traffic Scenario", fontsize=10, fontweight="bold", labelpad=10)
    ax.set_ylabel("DDL Case" if show_ylabels else "", fontsize=10, fontweight="bold")
    ax.set_xticklabels(
        [short_scenario_label(value) for value in matrix.columns],
        rotation=55,
        ha="right",
        fontsize=7,
    )
    if show_ylabels:
        labels = ylabels or [short_ddl_label(display_ddl_label(value)) for value in matrix.index]
        ax.set_yticklabels(labels, rotation=0, fontsize=8)
    else:
        ax.tick_params(axis="y", left=False, labelleft=False)


def add_level_legend(fig, y=0.01):
    handles = [
        mpatches.Patch(color=color, label=level)
        for level, color in zip(LEVELS, LEVEL_COLORS)
    ]
    fig.legend(
        handles=handles,
        loc="lower center",
        ncol=3,
        frameon=False,
        bbox_to_anchor=(0.5, y),
    )


def save_pass_comparison(df, output_path, config_name):
    pass_numbers = list(range(22, 37))
    intended = pivot(df, "IntendedLevel", pass_numbers)
    actual = pivot(df, "ActualLevel", pass_numbers)

    sns.set_theme(style="white", font_scale=0.85)
    fig, axes = plt.subplots(1, 2, figsize=(31, 11), sharey=False)
    draw_level_heatmap(axes[0], intended, "A. Intended Risk Level")
    draw_level_heatmap(
        axes[1],
        actual,
        f"B. MigraGuard Result ({config_name} config)",
        show_ylabels=False,
    )
    for boundary in (5, 10):
        for ax in axes:
            ax.axhline(boundary, color="#202020", linewidth=2.0)
    fig.suptitle(
        "Squawk-Passing DDLs: Intended vs Workload-Aware Risk Levels",
        fontsize=19,
        fontweight="bold",
        y=0.97,
    )
    fig.text(
        0.5,
        0.925,
        "15 DDLs passed Squawk (5 Safe, 5 Warning, 5 Danger); 20 workload scenarios",
        ha="center",
        fontsize=10,
        color="#555555",
    )
    add_level_legend(fig, y=0.012)
    fig.subplots_adjust(left=0.14, right=0.99, top=0.86, bottom=0.25, wspace=0.08)
    fig.savefig(output_path, dpi=240, facecolor="white", bbox_inches="tight")
    plt.close(fig)

    matches = (df[df["DDLNumber"].between(22, 36)]["IntendedLevel"] ==
               df[df["DDLNumber"].between(22, 36)]["ActualLevel"])
    return int(matches.sum()), int((~matches).sum())


def save_linter_warning_heatmap(df, output_path, config_name):
    warning_numbers = list(range(37, 42))
    actual = pivot(df, "ActualLevel", warning_numbers)
    labels = []
    for ddl_name in actual.index:
        number = ddl_number(ddl_name)
        labels.append(
            f"{short_ddl_label(display_ddl_label(ddl_name), 38)}  "
            f"[{LINTER_WARNING_RULES[number]}]"
        )

    sns.set_theme(style="white", font_scale=0.9)
    fig, ax = plt.subplots(figsize=(18, 6.8))
    draw_level_heatmap(
        ax,
        actual,
        f"MigraGuard Results for Squawk-Warning DDLs ({config_name} config)",
        ylabels=labels,
    )
    fig.text(
        0.5,
        0.91,
        "5 DDLs flagged by Squawk; each cell shows MigraGuard's workload-aware result",
        ha="center",
        fontsize=10,
        color="#555555",
    )
    add_level_legend(fig, y=0.012)
    fig.subplots_adjust(left=0.3, right=0.98, top=0.82, bottom=0.31)
    fig.savefig(output_path, dpi=240, facecolor="white", bbox_inches="tight")
    plt.close(fig)

    subset = df[df["DDLNumber"].between(37, 41)]
    distribution = subset["ActualLevel"].value_counts().reindex(LEVELS, fill_value=0)
    return distribution


def main():
    args = parse_args()
    csv_path = args.csv_path.resolve()
    output_dir = (
        args.output_dir.resolve()
        if args.output_dir
        else csv_path.parent / "figures" / "linter_migraguard_comparison"
    )
    output_dir.mkdir(parents=True, exist_ok=True)

    df = load_data(csv_path, args.config)
    pass_path = output_dir / "01_squawk_pass_intended_vs_migraguard.png"
    warning_path = output_dir / "02_squawk_warning_migraguard_results.png"

    matches, changes = save_pass_comparison(df, pass_path, args.config)
    warning_distribution = save_linter_warning_heatmap(df, warning_path, args.config)

    print("Squawk-pass group (300 combinations)")
    print(f"  Intended-level matches: {matches}")
    print(f"  Workload-driven changes: {changes}")
    print("Squawk-warning group (100 combinations)")
    for level in LEVELS:
        print(f"  {level}: {int(warning_distribution[level])}")
    print(f"Saved: {pass_path}")
    print(f"Saved: {warning_path}")


if __name__ == "__main__":
    main()
