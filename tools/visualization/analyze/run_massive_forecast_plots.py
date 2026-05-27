import os
import subprocess
import glob
import shutil
import sys
from datetime import datetime
import yaml

def run_massive_batch_plots(run_dir=None):
    ddl_dir = "experiments/ddl"
    artifact_base = "/home/gyeongho/.gemini/antigravity-cli/brain/fc61e4a8-7c53-4959-8257-7d147633cdc3"
    config_dir = "experiments/configs/cases"
    
    # Resolve run_dir from arguments or fallback
    if not run_dir:
        if len(sys.argv) > 1:
            run_dir = sys.argv[1]
        else:
            # Fallback to dynamic timestamp generation if run standalone without args
            timestamp_str = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")
            run_dir = f"experiments/reports/batch_runs/{timestamp_str}"
            
    run_dir = os.path.abspath(run_dir)
    timestamp_str = os.path.basename(run_dir)
    
    # Scan experiments/configs/cases/*.yaml to load cases dynamically
    cases = {}
    config_files = glob.glob(os.path.join(config_dir, "*.yaml"))
    
    if not config_files:
        print(f"❌ [ERROR] No case configurations found in {config_dir}")
        return
        
    for conf_path in sorted(config_files):
        try:
            with open(conf_path, 'r', encoding='utf-8') as f:
                data = yaml.safe_load(f)
                if not data or 'metadata' not in data:
                    print(f"⚠️ Warning: Skip invalid configuration {conf_path} (missing metadata)")
                    continue
                metadata = data['metadata']
                case_name = metadata.get('case_name')
                
                # Overwrite DB path to use the local runs DB instance
                db_path = os.path.join(run_dir, "data", f"{case_name}.db")
                
                cases[case_name] = {
                    "db": db_path,
                    "config": conf_path
                }
        except Exception as e:
            print(f"⚠️ Error reading config {conf_path}: {e}")
            continue
            
    if not cases:
        print("❌ [ERROR] No valid cases loaded. Aborting.")
        return
        
    ddl_files = sorted(glob.glob(os.path.join(ddl_dir, "*.sql")))
    
    print("==================================================")
    print("🚀 Starting Massive Multi-Capacity DDL 24h Forecasting Plotter...")
    print(f"📂 Execution Run Path: {run_dir}/")
    print("==================================================")
    
    # Create target directories dynamically
    for case_name in cases.keys():
        os.makedirs(os.path.join(run_dir, case_name), exist_ok=True)
        # Mirror in artifact directory to ensure system integration
        os.makedirs(os.path.join(artifact_base, "batch_runs", timestamp_str, case_name), exist_ok=True)
        
    for case_name, settings in cases.items():
        print(f"\n🔥 Processing Case: [{case_name}] using DB: {settings['db']}...")
        
        # Ensure database actually exists in the run pack
        if not os.path.exists(settings['db']):
            print(f"   ⚠️ Skipping entire case [{case_name}] - database file missing in runs pack: {settings['db']}")
            continue
            
        for ddl_path in ddl_files:
            ddl_basename = os.path.basename(ddl_path).replace(".sql", "")
            
            # 1. Run MigraGuard forecast analyze (exit code 1 danger is handled)
            cmd_analyze = [
                "./build/migraguard", "analyze", ddl_path,
                "--sandbox", settings["db"],
                "--config", settings["config"],
                "--forecast", "-o", "console"
            ]
            subprocess.run(cmd_analyze, capture_output=True)
            
            csv_file = "predictive_forecast.csv"
            if not os.path.exists(csv_file):
                print(f"   ⚠️ Skipping {ddl_basename} (no forecast CSV generated)")
                continue
                
            # 2. Run plot tool on the generated CSV
            cmd_plot = [
                "python3", "tools/visualization/analyze/plot_predictive_heatmap.py",
                csv_file
            ]
            subprocess.run(cmd_plot, capture_output=True)
            
            default_png = "predictive_risk_heatmap.png"
            if os.path.exists(default_png):
                output_filename = f"{ddl_basename}_predictive_risk_heatmap.png"
                
                try:
                    # Move to official reports folder under run_dir/[case]
                    dest_report_path = os.path.join(run_dir, case_name, output_filename)
                    shutil.move(default_png, dest_report_path)
                    
                    # Copy to artifact folder for document integrity
                    dest_artifact_path = os.path.join(artifact_base, "batch_runs", timestamp_str, case_name, output_filename)
                    shutil.copy(dest_report_path, dest_artifact_path)
                    
                    print(f"   ✅ Saved: {case_name}/{output_filename}")
                except Exception as e:
                    print(f"   ⚠️ Error archiving plot for {ddl_basename}: {e}")
                
                # Clean up temporary CSV
                try:
                    if os.path.exists(csv_file):
                        os.remove(csv_file)
                except OSError:
                    pass
            else:
                print(f"   ❌ Plot failed for {ddl_basename} (congestion or calculation error)")
                
    print(f"\n✨ Massive Batch Run plotting complete.")
    print(f"📍 Archived Path: {run_dir}/")

if __name__ == "__main__":
    run_massive_batch_plots()
