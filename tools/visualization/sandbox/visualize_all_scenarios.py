import os
import glob
import subprocess
import sys

def run_batch_sandbox_viz():
    # Setup paths relative to script location
    script_dir = os.path.dirname(os.path.abspath(__file__))
    base_dir = os.path.dirname(os.path.dirname(os.path.dirname(script_dir))) # Project root
    metrics_dir = os.path.join(base_dir, "experiments", "reports", "metrics")
    plot_script = os.path.join(script_dir, "plot_scenario_metrics.py")

    print("==================================================")
    print("🛡️  MigraGuard: Batch Sandbox Visualization")
    print("==================================================")

    csv_files = glob.glob(os.path.join(metrics_dir, "*.csv"))
    
    if not csv_files:
        print(f"❓ No metrics found in {metrics_dir}")
        return

    for f in csv_files:
        print(f"\n🔄 Processing: {os.path.basename(f)}")
        subprocess.run([sys.executable, plot_script, f])

    print("\n✅ Batch visualization complete.")

if __name__ == "__main__":
    run_batch_sandbox_viz()
