import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt
import sys
import os

def plot_research_results(csv_path):
    print(f"🔥 [Analyze] Plotting research report: {csv_path}")
    try:
        df = pd.read_csv(csv_path)
    except Exception as e:
        print(f"❌ Error reading CSV: {e}")
        return

    if df.empty or 'RiskScore' not in df.columns:
        print("❌ CSV is empty or malformed.")
        return

    plt.figure(figsize=(15, 12))
    sns.set_theme(style="whitegrid")

    # 1. Clean labels
    df['ScenarioShort'] = df['Scenario'].apply(lambda x: os.path.basename(x).replace('.db', ''))
    df['SQLFileShort'] = df['SQLFile'].apply(lambda x: os.path.basename(x).replace('.sql', '')[:25])

    # 2. Heatmap
    plt.subplot(2, 1, 1)
    pivot_df = df.pivot_table(index='SQLFileShort', columns='ScenarioShort', values='RiskScore', aggfunc='mean')
    sns.heatmap(pivot_df, annot=True, cmap='RdYlGn_r', fmt=".1f", cbar_kws={'label': 'Risk Score (%)'})
    plt.title("Migration Risk Heatmap (Scenario vs SQL Case)", fontsize=14)

    # 3. Distribution
    plt.subplot(2, 2, 3)
    level_counts = df['RiskLevel'].value_counts()
    colors = {'Safe': '#2ca02c', 'Warning': '#ff7f0e', 'Danger': '#d62728'}
    level_counts.plot(kind='bar', color=[colors.get(x, '#1f77b4') for x in level_counts.index])
    plt.title("Risk Level Distribution", fontsize=14)
    plt.xticks(rotation=0)

    # 4. Correlation
    plt.subplot(2, 2, 4)
    df['TableSizeMB'] = df['TableSize'] / (1024 * 1024)
    sns.scatterplot(data=df, x='TableSizeMB', y='T_ddl', hue='RiskLevel', palette=colors, style='RewriteRequired', s=100)
    plt.title("Impact: DDL Time vs Table Size", fontsize=14)
    plt.xlabel("Table Size (MB)")

    plt.tight_layout()
    output_png = csv_path.replace('.csv', '_analysis.png')
    plt.savefig(output_png, dpi=200)
    print(f"✨ Saved analysis plot to: {output_png}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python plot_research_report.py <csv_path>")
        sys.exit(1)
    plot_research_results(sys.argv[1])
