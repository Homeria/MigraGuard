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
    plt.figure(figsize=(12, 7))
    plt.title("🛡️ MigraGuard: 24-Hour Predictive Risk Forecast", fontsize=16, fontweight='bold', pad=20)
    
    # Plot Expected TPS (The workload curve)
    plt.plot(df['Hour'], df['ExpectedTPS'], color='#2c3e50', linewidth=2.5, label='Predicted TPS (Lambda)', marker='o', markersize=4)
    plt.fill_between(df['Hour'], df['ExpectedTPS'], color='#bdc3c7', alpha=0.2)
    
    # Overlay Risk Heatmap (Background coloring)
    for i in range(len(df)):
        hour = df.iloc[i]['Hour']
        score = df.iloc[i]['RiskScore']
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
            plt.axvspan(hour - 0.5, hour + 0.5, color=color, alpha=alpha, lw=0)

    # Highlight the Best Hour (Golden Window)
    best_row = df[df['IsBestHour'] == True]
    if not best_row.empty:
        best_hour = best_row.iloc[0]['Hour']
        best_tps = best_row.iloc[0]['ExpectedTPS']
        best_score = best_row.iloc[0]['RiskScore']
        
        plt.annotate(f"★ Recommended Golden Window\nHour: {best_hour:02d}:00 | Risk: {best_score:.1f}%",
                     xy=(best_hour, best_tps), xytext=(best_hour, best_tps + (max(df['ExpectedTPS']) * 0.15)),
                     arrowprops=dict(facecolor='#27ae60', shrink=0.05, headwidth=10, width=2),
                     fontsize=11, fontweight='bold', color='#1e8449', ha='center',
                     bbox=dict(boxstyle="round,pad=0.5", fc="#e9f7ef", ec="#27ae60", alpha=0.9))

    # Formatting
    plt.xlabel("Hour of the Day (24h Forecast)", fontsize=12)
    plt.ylabel("Expected Transactions Per Second (TPS)", fontsize=12)
    plt.xticks(range(0, 24))
    plt.grid(axis='y', linestyle='--', alpha=0.5)
    plt.xlim(-0.5, 23.5)
    plt.ylim(0, max(df['ExpectedTPS']) * 1.4) # Add space for annotation
    
    # Legends
    tps_patch = mpatches.Patch(color='#2c3e50', label='Expected Workload')
    safe_patch = mpatches.Patch(color='#2ecc71', alpha=0.3, label='Safe Zone (Low Risk)')
    warn_patch = mpatches.Patch(color='#f1c40f', alpha=0.3, label='Warning Zone (Moderate Risk)')
    danger_patch = mpatches.Patch(color='#e74c3c', alpha=0.3, label='Danger Zone (High Risk)')
    plt.legend(handles=[tps_patch, safe_patch, warn_patch, danger_patch], loc='upper right', frameon=True, shadow=True)

    # Information box
    info_text = f"Analyzed Table: {os.path.basename(csv_path)}\nModel: MigraGuard v3.8 (Predictive Engine)"
    plt.text(0.02, 0.02, info_text, transform=plt.gca().transAxes, fontsize=9, verticalalignment='bottom', 
             bbox=dict(boxstyle="round,pad=0.3", fc="white", ec="gray", alpha=0.8))

    plt.tight_layout()
    plt.savefig(output_path, dpi=300)
    print(f"[SUCCESS] Predictive heatmap saved to: {output_path}")

if __name__ == "__main__":
    target_csv = "predictive_forecast.csv"
    if len(sys.argv) > 1:
        target_csv = sys.argv[1]
    
    plot_predictive_heatmap(target_csv)
