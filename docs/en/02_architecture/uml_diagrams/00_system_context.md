# Level 0: System Context Diagram

This diagram provides a high-level overview of how **MigraGuard** interacts with external actors and system components during the database migration lifecycle.

```mermaid
graph TD
    subgraph Users
        DEV[Developer]
        CICD[CI/CD Pipeline]
    end

    subgraph MigraGuard Ecosystem
        CLI[MigraGuard CLI]
        SDK[Core SDK - pkg/migraguard]
        PY[Python Visualizer - Matplotlib]
    end

    subgraph Data Layers
        PG[(Target PostgreSQL)]
        SQLITE[(Sandbox/Metrics SQLite)]
    end

    DEV -->|Runs Commands| CLI
    CICD -->|Gatekeeper/Analyze| CLI
    
    CLI --> SDK
    SDK -->|Fetch Real-time Metrics| PG
    SDK -->|Persist/Retrieve History| SQLITE
    SDK -->|Generate Sandbox| SQLITE
    
    CLI -->|Export CSV| PY
    PY -->|Generate Heatmap PNG| CICD
    PY -->|Generate Heatmap PNG| DEV

    style CLI fill:#f9f,stroke:#333,stroke-width:4px
    style SDK fill:#bbf,stroke:#333,stroke-width:2px
    style PG fill:#3498db,stroke:#333,stroke-width:2px
    style SQLITE fill:#95a5a6,stroke:#333,stroke-width:2px
```

### Key Interactions
1.  **Developer/Pipeline**: Initiates the process via CLI (e.g., `analyze`, `simulate`).
2.  **CLI & SDK**: The CLI acts as a wrapper for the Core SDK, which contains the mathematical risk engine.
3.  **Data Persistence**: Real-time metrics are harvested from PostgreSQL, while historical trends and simulation data are stored in SQLite.
4.  **Visualization**: Python scripts are used as a downstream consumer of the analysis data to produce high-fidelity reports.
