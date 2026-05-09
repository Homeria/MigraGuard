# 🧪 MigraGuard Simulation Guide

This guide describes the procedures for validating MigraGuard's risk detection capabilities using the real-world load generator.

---

## 1. Environment Initialization
```bash
docker compose down -v
docker compose up -d --build
```

## 2. Load Simulation
- **Steady State**: Automatically starts with `docker-compose` (15 workers).
- **Peak State (Flash Sale)**:
  ```bash
  docker compose run -d --name load-gen-peak load-generator --conns 50 --profile flash-sale
  ```

## 3. Risk Analysis Validation
- **Table Rewrite (Danger)**: Analyze `007_danger_rewrite_type.sql`.
- **Metadata Update (Safe)**: Analyze `004_safe_add_column.sql`.

## 4. Verification Checklist
- Do `Avg(1h)` and `Peak(24h)` metrics reflect the actual load?
- Does the risk level for the same DDL change dynamically based on traffic?
- Is the recommended golden window reasonable?
