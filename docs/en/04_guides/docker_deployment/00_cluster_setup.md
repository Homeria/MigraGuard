# 🐳 Docker: 00. Cluster Setup

Deploy the full MigraGuard ecosystem (DB, Agent, LoadGen) using Docker Compose.

---

## 1. Initial Deployment
Builds and starts the target database, background collector, and load generator.
```bash
docker compose up -d --build
```

## 2. Service Overview
- **`db`**: PostgreSQL 15 instance.
- **`agent`**: Background collector.
- **`load-generator`**: Real-time traffic simulator.
- **`analyze`**: CLI tool for risk assessment.

## 3. Persistent Storage
- SQLite data is stored in the `migraguard-data` volume.
- PostgreSQL data is stored in the `postgres-data` volume.

---

## 💡 Developer Tip: Rebuilding on Code Changes
When you modify Go source code (e.g., adding a new command), you **must rebuild the image** for changes to take effect. Since the `analyze` service is now part of the default profile, simply run:

```bash
# Rebuild and start everything including Analyze
docker compose up -d --build
```
