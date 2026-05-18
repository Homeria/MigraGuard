import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.patches as mpatches
import os
import sys

def plot_predictive_heatmap(csv_path, output_path="predictive_risk_heatmap.png"):
    if not os.path.exists(csv_path):
        print(f"[ERROR] Forecast data not found at {csv_path}")
        return

    # Load data
    df = pd.read_csv(csv_path)
    
    # Setup plot
    fig, ax1 = plt.subplots(figsize=(12, 7))
    plt.title("🛡️ MigraGuard: 24-Hour Predictive Risk Forecast", fontsize=16, fontweight='bold', pad=20)
    
    # Ax1: Expected TPS (The workload curve)
    color_tps = '#2c3e50'
    ax1.set_xlabel("Hour of the Day (24h Forecast)", fontsize=12)
    ax1.set_ylabel("Expected Transactions Per Second (TPS)", fontsize=12, color=color_tps)
    line_tps = ax1.plot(df['Hour'], df['ExpectedTPS'], color=color_tps, linewidth=2.5, label='Predicted TPS (Average)', marker='o', markersize=4, zorder=5)
    ax1.tick_params(axis='y', labelcolor=color_tps)
    
    # Plot Variance Cloud (Min/Max range)
    if 'MinTPS' in df.columns and 'MaxTPS' in df.columns:
        ax1.fill_between(df['Hour'], df['MinTPS'], df['MaxTPS'], color='#34495e', alpha=0.1, label='Traffic Variance (Min-Max)', zorder=4)
    
    # Ax2: P99 Latency (Secondary Axis)
    ax2 = ax1.twinx()
    color_p99 = '#3498db'
    ax2.set_ylabel("Predicted P99 Latency (ms)", fontsize=12, color=color_p99)
    line_p99 = ax2.plot(df['Hour'], df['ExpectedP99'], color=color_p99, linewidth=1.5, linestyle='--', label='Predicted P99 Latency', alpha=0.7, zorder=3)
    ax2.tick_params(axis='y', labelcolor=color_p99)
    
    # Overlay Risk Heatmap (Background coloring on Ax1)
    for i in range(len(df)):
        hour = df.iloc[i]['Hour']
        level = df.iloc[i]['RiskLevel']
        is_safe = df.iloc[i]['IsSafeWindow']
        
        color = 'none'
        alpha = 0.15
        
        if level == "Danger":
            color = '#e74c3c' # Red
        elif level == "Warning":
            color = '#f1c40f' # Yellow
        elif is_safe:
            color = '#2ecc71' # Green
            
        if color != 'none':
            ax1.axvspan(hour - 0.5, hour + 0.5, color=color, alpha=alpha, lw=0)

    # Highlight the Best Hour (Golden Window)
    best_row = df[df['IsBestHour'] == True]
    if not best_row.empty:
        best_hour = best_row.iloc[0]['Hour']
        best_tps = best_row.iloc[0]['ExpectedTPS']
        best_score = best_row.iloc[0]['RiskScore']
        
        ax1.annotate(f"★ Recommended Golden Window\nHour: {best_hour:02d}:00 | Risk: {best_score:.1f}%",
                     xy=(best_hour, best_tps), xytext=(best_hour, best_tps + (max(df['ExpectedTPS']) * 0.15)),
                     arrowprops=dict(facecolor='#27ae60', shrink=0.05, headwidth=10, width=2),
                     fontsize=11, fontweight='bold', color='#1e8449', ha='center',
                     bbox=dict(boxstyle="round,pad=0.5", fc="#e9f7ef", ec="#27ae60", alpha=0.9))

    # Formatting
    ax1.set_xticks(range(0, 24))
    ax1.grid(axis='both', linestyle='--', alpha=0.3)
    ax1.set_xlim(-0.5, 23.5)
    ax1.set_ylim(0, max(df['ExpectedTPS']) * 1.4)
    ax2.set_ylim(0, max(df['ExpectedP99']) * 1.5)
    
    # Legends
    safe_patch = mpatches.Patch(color='#2ecc71', alpha=0.3, label='Safe Zone (Low Risk)')
    warn_patch = mpatches.Patch(color='#f1c40f', alpha=0.3, label='Warning Zone (Moderate Risk)')
    danger_patch = mpatches.Patch(color='#e74c3c', alpha=0.3, label='Danger Zone (High Risk)')
    
    lines = line_tps + line_p99
    labels = [l.get_label() for l in lines]
    ax1.legend(lines + [safe_patch, warn_patch, danger_patch], 
               labels + ['Safe Zone', 'Warning Zone', 'Danger Zone'], 
               loc='upper right', frameon=True, shadow=True, fontsize=9)

    # Information box
    info_text = f"Analyzed Table: {os.path.basename(csv_path)}\nModel: MigraGuard v3.8 (Enhanced Visualization)"
    ax1.text(0.02, 0.02, info_text, transform=ax1.transAxes, fontsize=9, verticalalignment='bottom', 
             bbox=dict(boxstyle="round,pad=0.3", fc="white", ec="gray", alpha=0.8))

    plt.tight_layout()
    plt.savefig(output_path, dpi=300)
    print(f"[SUCCESS] Enhanced heatmap saved to: {output_path}")

if __name__ == "__main__":
    target_csv = "predictive_forecast.csv"
    if len(sys.argv) > 1:
        target_csv = sys.argv[1]
    
    plot_predictive_heatmap(target_csv)
