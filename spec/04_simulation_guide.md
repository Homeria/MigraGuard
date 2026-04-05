# 🧪 MigraGuard v3.2 Simulation & Validation Guide

This document describes the procedures for setting up a production-like PostgreSQL environment using Docker, generating traffic, and validating the risk analysis engine.

---

## 1. Environment Setup

### [1-1] Docker Infrastructure
MigraGuard uses `docker-compose.yml` to spin up a target database and a background metrics collector.

```bash
docker-compose up -d
```

- **PostgreSQL (`db`):** `localhost:5432` (user: `user`, password: `pass`, db: `target_db`).
- **Agent (`agent`):** Continuously collects metrics from the `db` and stores them in a shared SQLite database.
- **Shared Volume:** `/app/data` is shared between the agent and the CLI.

### [1-2] Database Initialization
The target database is automatically initialized with the `pg_stat_statements` extension via `init-db.sql`.

```sql
-- init-db.sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
```

### [1-3] Test Data Generation
Use `pgbench` to create an initial dataset (e.g., 1 million rows).

```bash
docker exec -it postgres pgbench -i -s 10 target_db
```

---

## 2. Traffic Simulation

### [2-1] Baseline Load (Regular Traffic)
Simulate a normal workload to establish a traffic baseline in the SQLite repository.

```bash
# 10 users for 5 minutes
docker exec -it postgres pgbench -c 10 -T 300 target_db
```

### [2-2] Peak Load Simulation
Generate a high traffic spike to test the engine's sensitivity to peak concurrency.

```bash
# 50 users for 1 minute
docker exec -it postgres pgbench -c 50 -j 4 -T 60 target_db
```

---

## 3. Risk Analysis Validation

Execute the `analyze` command while traffic is being generated or shortly after a peak period.

### [3-1] Write a Migration SQL (`test.sql`)
```sql
ALTER TABLE pgbench_accounts ALTER COLUMN abalance TYPE BIGINT;
```

### [3-2] Run Analysis
```bash
# Run the analysis CLI (e.g., within the analyze-shell container)
./migraguard analyze test.sql --db "postgres://user:pass@db:5432/target_db" --sqlite "/app/data/migraguard.db"
```

---

## 4. Verification Checklist (Success Criteria)

1.  **Metric Continuity:** Verify that `table_metrics` in `migraguard.db` is populated every 10 seconds.
2.  **Peak Detection:** Ensure that `Analyze` reports include the `PeakTPS24h` captured during your peak simulation.
3.  **Gatekeeping:** Confirm that the process returns `Exit Code 1` when the `RiskScore` exceeds the critical threshold (e.g., 90%).
4.  **Data Retention:** Observe the SQLite file size over time to ensure the `PurgeOldSnapshots` logic is active.

## 5. Summary Table: Load Scenarios

| Scenario | Concurrency | Expected Outcome | Risk Level |
| :--- | :--- | :--- | :--- |
| **Idle** | 0 | Minimum $C_{peak}$, fast $T_{ddl}$ | **Safe** |
| **Moderate** | 10 | $RiskScore$ < 50% | **Safe** |
| **High** | 50 | $RiskScore$ > 80% | **Warning** |
| **Peak+Rewrite** | 100+ | $RiskScore$ > 95% | **Danger** |
