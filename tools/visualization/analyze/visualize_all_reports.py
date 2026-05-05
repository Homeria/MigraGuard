import os
import glob
import subprocess
import sys

def run_batch_analyze_viz():
    # Setup paths relative to script location
    script_dir = os.path.dirname(os.path.abspath(__file__))
    base_dir = os.path.dirname(os.path.dirname(os.path.dirname(script_dir))) # Project root
    reports_dir = os.path.join(base_dir, "experiments", "reports")
    plot_script = os.path.join(script_dir, "plot_research_report.py")

    print("==================================================")
    print("🛡️  MigraGuard: Batch Analyze Visualization")
    print("==================================================")

    # Look for research_results.csv or any other report CSVs in the root of reports/
    csv_files = glob.glob(os.path.join(reports_dir, "*.csv"))
    
    if not csv_files:
        print(f"❓ No report CSVs found in {reports_dir}")
        return

    for f in csv_files:
        print(f"\n🔥 Processing: {os.path.basename(f)}")
        subprocess.run([sys.executable, plot_script, f])

    print("\n✅ Batch visualization complete.")

if __name__ == "__main__":
    run_batch_analyze_viz()
