# 🛡️ MigraGuard: Expert Consultation Feedback & System Limitations Report

This document records the architectural and practical limitations (flaws) of the current MigraGuard v3.8 model, based on consultative feedback from senior backend and Site Reliability Engineering (SRE) experts, and presents defensive response scenarios to address them.

---

## 1. 📋 Summary of Expert Consultation Feedback

| Category | Consultation Feedback | Practical Evaluation |
| :--- | :--- | :--- |
| **Topic Selection** | Traffic-aware circuit breaking right before deployment (CI/CD Gate). | Excellent. A superb approach that fills the gap between static linters and post-deployment monitoring. |
| **Parser Reliability** | Adopting PostgreSQL's official AST parser via `pg_query_go`. | Excellent. Completely eliminates the risk of false positives inherent in regex-based parsers. |
| **Operational Overhead** | SQLite local caching and lightweight metrics querying at 1-minute intervals. | Good. The design minimizes read stress on the live production database. |
| **Practical Adoptability** | Can it be deployed immediately in large enterprise/fintech production environments? | **Impossible**. Barriers such as network segregation, uncertainty in heuristic formulas, and the existence of alternative technologies exist. |

---

## 2. ⚠️ 4 Core Architectural & Algorithmic Flaws of MigraGuard

### ① The Database Sync Problem (SQLite Sync & Network Segregation)
* **Flaw**: 
  - The metrics collected by the agent ([agent_service.go](file:///home/gyeongho/Github/MigraGuard/cmd/migraguard/agent_service.go)) from production servers are written to a local SQLite file (`migraguard.db`) on that machine.
  - However, CI/CD runners (like GitHub Actions) executing the actual deployment run in ephemeral, virtual container environments. 
  - There is currently no network synchronization strategy for how the deployment container will read the SQLite file residing on a physically isolated server. Exfiltrating files inside a secure network via SSH or S3 to an external container for every deploy will not pass security audits.
* **Practical Remedy (To-Do)**:
  - Decouple the local SQLite file dependency and build a centralized **HTTP Metrics API Server**.
  - Alternatively, integrate with industry-standard TSDBs like **Prometheus or VictoriaMetrics** so that the analyzer queries traffic trends remotely via standard HTTP queries.

### ② The Magic Number Problem (Lack of Mathematical Formula Credibility)
* **Flaw**:
  - The current risk scoring formulas ([risk_evaluator.go](file:///home/gyeongho/Github/MigraGuard/pkg/migraguard/internal/analyzer/risk_evaluator.go)) rely on arbitrary fixed constants (magic numbers) such as `avg_multiplier: 1.2` and `lockImpact: 0.5`.
  - In reality, database performance varies non-linearly due to **the number of indexes (every index is rebuilt during table rewrite), TOAST table storage, OS page cache hit ratios, and CPU scheduler states**, even for tables of identical sizes.
  - Relying on a tool's "Safe" prediction to deploy during the day can lead to massive lock contentions and outages, for which the tool cannot take responsibility.
* **Practical Remedy (To-Do)**:
  - Replace static magic numbers with an **Auto-Calibration Model** that continuously tunes parameters by tracing the historical DDL execution logs and `pg_stat_statements` patterns of the target database.

### ③ The "Why Bother?" Problem (Competition with Online Schema Change Tools)
* **Flaw**:
  - When a highly risky DDL (such as a table-rewrite-inducing column type change) is detected, the system simply halts deployment (circuit breaking) and suggests deploying during the midnight golden window.
  - However, in high-volume production, senior engineers completely bypass lock contentions for tables larger than dozens of GBs by using online schema change tools like `gh-ost` or `pt-online-schema-change`.
  - A tool that only blocks pipelines and warns developers is perceived as a productivity-hindering roadblock rather than a helpful assistant.
* **Practical Remedy (To-Do)**:
  - Beyond mere blocking and delaying, provide an **Online Schema Change Prescription** when high risk is detected.
  - Example: *"This ALTER TABLE query triggers a full table rewrite. We recommend utilizing pt-osc. Given the current replication lag of 1.2s, the recommended execution script is: `pt-online-schema-change --max-lag=2s ...`"*

### ④ The Privilege & Audit Problem (Lack of Minimum Permissions and Logging)
* **Flaw**:
  - The agent holds database connection credentials ([migraguard.yaml](file:///home/gyeongho/Github/MigraGuard/migraguard.yaml)) to harvest PostgreSQL metrics.
  - In fintech, having a third-party tool connect to a production database in real time is a significant security threat. If the agent's privilege limits are not officially documented and it lacks audit logs, enterprise adoption will be rejected.
* **Practical Remedy (To-Do)**:
  - Construct and publish a **Least Privilege Guide** ensuring the agent's DB account has zero access to user table data (DML/DCL) and only read-only access to metadata views (`pg_catalog` and `pg_stat_statements`).
  - Implement a built-in Audit Log specification where the agent records every single query it runs against the target database.

---

## 3. 🎯 Graduation Defense & Thesis Presentation Q&A Scenarios

Use these guidelines when committee members ask if this prototype is actually adoptable in a real-world DevSecOps pipeline.

> **[Q] Since this relies on mock simulation data, it doesn't seem ready for real production environments?**
> - **[A]** "That is not the case. The system's architecture cleanly isolates **Live Mode for actual deployment pipelines** and **Sandbox Mode for research evaluation**. In Live Mode, the background agent resides on the real PostgreSQL host, safely collecting real-time production traffic metrics (TPS, P99) into the local database. Sandbox Mode is built purely for academic evaluation to construct a consistent, reproducible, and mathematically controlled experimental environment."
>
> **[Q] Why should we adopt MigraGuard if we can just use Online DDL tools like pt-osc anyway?**
> - **[A]** "While online schema change tools bypass locks, they can still degrade performance depending on the current replication lag and cannot prevent human errors where a developer accidentally bypasses the tool and runs raw DDL directly in production. MigraGuard serves as the **final DevSecOps Safety Net** that forcibly intercepts raw DDL before deployment, checks live traffic safety, and provides the correct online tool prescriptions when necessary. Thus, they are complementary, not mutually exclusive."
