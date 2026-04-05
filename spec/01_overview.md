# 🛡️ MigraGuard Project Overview (v3.2 Capstone Finalized)

## 1. Project Definition (What is MigraGuard?)
**MigraGuard** is a DevSecOps agent and CLI tool designed to pre-emptively block database schema changes (DDL) that could cause lock contention or service outages. It is the core outcome of the Capstone Design project, integrating a background **Agent** that continuously learns traffic patterns with an **Analyze CLI** that evaluates migration risks.

## 2. Core Value
> "A traffic-aware analysis model deactivates DDL time bombs."
- **Dual-Process Architecture:** Separates data collection (Agent) from analysis (CLI) for system stability.
- **Advanced 5-Step Risk Model:** Combines static SQL analysis with dynamic peak traffic weighting for precise diagnosis.
- **Instant Feedback:** Leverages SQLite time-series data for immediate risk scoring without additional wait times.
- **Deployment Gatekeeper:** Integrates with CI/CD pipelines (e.g., GitHub Actions) to provide risk-based automated deployment control.

## 3. Key Features (v3.2)
1. **Background Agent (Metrics Collector):**
   - Docker-based service that captures `pg_stat_statements` metrics from the production database.
   - Accumulates time-series data in a shared SQLite database with automated retention policies.
2. **Analyze CLI (Migration Analyzer):**
   - Parses SQL AST to identify target tables and DDL types (e.g., Full Rewrite detection).
   - Utilizes historical data collected by the Agent to trigger the **5-step analysis logic**.
3. **5-Step Risk Evaluation Engine:**
   - **Step 1: Schema Analysis ($T_{ddl}$)** - Static DDL impact assessment.
   - **Step 2: Lock Contention ($T_{block}$)** - Potential lock duration estimation.
   - **Step 3: Peak Concurrency ($C_{peak}$)** - 24h peak traffic and load analysis.
   - **Step 4: Recovery Assessment ($T_{rec}$)** - Cost of failure based on table size.
   - **Step 5: Final Risk Score ($\lambda_{final}$)** - Weighted risk score calculation.
4. **Unified Reporting:** Supports multi-format output including ANSI color consoles and Markdown for GitHub.

## 4. Technology Stack
- **Language:** Go (v1.22+)
- **Database:** SQLite (Metrics Storage), PostgreSQL (Target Database)
- **Infrastructure:** Docker & Docker Compose (Volume Sharing)
- **Architecture:** SOA-based service separation with interface-driven Dependency Injection.

## 5. Feature Roadmap & Milestones

### ✅ v3.2 Finalized (Current State)
- **SOA Refactoring**: Separation of domain logic (Analyze/Agent) into service layers.
- **Docker Persistence**: Implementation of shared volume structure for Agent-Analyze communication.
- **Risk Engine Verification**: 5-step risk model ($T_{ddl}, C_{peak}$, etc.) logic finalized and tested.

### 🚀 Future Roadmap (Phase 8~10)
1. **[Phase 8] CI/CD & GitHub Integration**:
   - `feat/github-actions`: Automated analysis posting on GitHub PRs and deployment gating.
2. **[Phase 9] API Mode & Observability**:
   - `feat/api-mode`: Transitioning the Agent into an HTTP API server for remote dashboard support.
3. **[Phase 10] ML-based Predictive Risk Engine**:
   - `feat/advanced-ml`: Time-series prediction for preemptive risk calculation.

---
*Last Updated: 2026-04-05 (v3.2 Final)*
