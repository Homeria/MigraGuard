import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import sys
import os

def save_individual_plots(csv_path):
    print(f"📊 [Sandbox] Plotting scenario metrics: {csv_path}")
    try:
        df = pd.read_csv(csv_path)
    except Exception as e:
        print(f"❌ Error reading CSV: {e}")
        return

    df['Timestamp'] = pd.to_datetime(df['Timestamp'])
    tables = df['TableName'].unique()
    colors = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728', '#9467bd']

    base_name = os.path.basename(csv_path).replace('.csv', '')
    output_dir = os.path.join(os.path.dirname(csv_path), f"{base_name}_plots")
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)

    plt.style.use('seaborn-v0_8-whitegrid')

    # 1. Load Profile
    plt.figure(figsize=(12, 6))
    ax1 = plt.gca()
    ax1_twin = ax1.twinx()
    for i, table in enumerate(tables):
        t_df = df[df['TableName'] == table]
        ax1.plot(t_df['Timestamp'], t_df['TPS'], label=f'{table} TPS', alpha=0.8, color=colors[i % len(colors)], linewidth=2)
    ax1_twin.fill_between(df['Timestamp'], 0, df['ActiveConnections'], color='gray', alpha=0.1, label='Connections')
    ax1.set_title(f"Load Profile ({base_name})"); ax1.set_ylabel('TPS'); ax1_twin.set_ylabel('Active Conns')
    ax1.legend(loc='upper left'); plt.tight_layout(); plt.savefig(os.path.join(output_dir, "01_load_profile.png")); plt.close()

    # 2. Performance (P99)
    plt.figure(figsize=(12, 6))
    for i, table in enumerate(tables):
        t_df = df[df['TableName'] == table]
        plt.plot(t_df['Timestamp'], t_df['P99Time'], label=f'{table} P99', alpha=0.8, color=colors[i % len(colors)])
    plt.title(f"P99 Response Time ({base_name})"); plt.ylabel('ms'); plt.legend(); plt.tight_layout()
    plt.savefig(os.path.join(output_dir, "02_latency_p99.png")); plt.close()

    # 3. Replication Lag
    plt.figure(figsize=(12, 6))
    for i, table in enumerate(tables):
        t_df = df[df['TableName'] == table]
        plt.plot(t_df['Timestamp'], t_df['ReplicationLag'], label=f'{table} Lag', alpha=0.8, color=colors[i % len(colors)])
    plt.title(f"Replication Lag ({base_name})"); plt.ylabel('sec'); plt.legend(); plt.tight_layout()
    plt.savefig(os.path.join(output_dir, "03_replication_lag.png")); plt.close()

    # 4. Storage Growth
    plt.figure(figsize=(12, 6))
    for i, table in enumerate(tables):
        t_df = df[df['TableName'] == table]
        plt.plot(t_df['Timestamp'], t_df['TableSize']/(1024*1024), label=f'{table} Size', alpha=0.8, color=colors[i % len(colors)])
    plt.title(f"Storage Growth ({base_name})"); plt.ylabel('MB'); plt.legend(); plt.tight_layout()
    plt.savefig(os.path.join(output_dir, "04_storage_growth.png")); plt.close()

    # 5. I/O Efficiency
    if 'SharedBlksHit' in df.columns:
        plt.figure(figsize=(12, 6))
        for i, table in enumerate(tables):
            tdf = df[df['TableName'] == table].copy()
            tdf['HitRatio'] = tdf['SharedBlksHit'] / (tdf['SharedBlksHit'] + tdf['SharedBlksRead'])
            plt.plot(tdf['Timestamp'], tdf['HitRatio'] * 100, label=f'{table} Hit %', alpha=0.8, color=colors[i % len(colors)])
        plt.title(f"Buffer Cache Hit Ratio ({base_name})"); plt.ylabel('%'); plt.ylim(80, 101); plt.legend(); plt.tight_layout()
        plt.savefig(os.path.join(output_dir, "05_cache_hit_ratio.png")); plt.close()

    # 6. Growth Rate
    plt.figure(figsize=(12, 6))
    for i, table in enumerate(tables):
        tdf = df[df['TableName'] == table].sort_values('Timestamp').copy()
        tdf['GrowthMB'] = tdf['TableSize'].diff() / (1024 * 1024)
        plt.plot(tdf['Timestamp'], tdf['GrowthMB'], label=f'{table} Growth', alpha=0.6, color=colors[i % len(colors)])
    plt.title(f"Data Growth Rate ({base_name})"); plt.ylabel('MB/Interval'); plt.legend(); plt.tight_layout()
    plt.savefig(os.path.join(output_dir, "06_storage_growth_rate.png")); plt.close()

    # 7. Saturation Correlation
    plt.figure(figsize=(10, 8))
    sns.scatterplot(data=df, x='TPS', y='P99Time', hue='TableName', alpha=0.5, palette='viridis')
    plt.title(f"Load vs Latency Saturation ({base_name})"); plt.xlabel('TPS'); plt.ylabel('P99 (ms)'); plt.tight_layout()
    plt.savefig(os.path.join(output_dir, "07_load_latency_correlation.png")); plt.close()

    print(f"✨ Saved 7 plots to: {output_dir}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python plot_scenario_metrics.py <csv_path>")
        sys.exit(1)
    save_individual_plots(sys.argv[1])
