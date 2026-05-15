# 🖥️ Native: 07. Global & Persistent Flags

These flags are managed at the root level and affect the behavior of all subcommands.

---

## 1. Configuration Flag

| Flag | Description | Default |
| :--- | :--- | :--- |
| `--config` | Points to the `migraguard.yaml` file. This file contains default DSNs, intervals, and risk thresholds to avoid long CLI strings. | `./migraguard.yaml` |

---

## 2. Debugging & Logging

| Flag | Shorthand | Description |
| :--- | :--- | :--- |
| `--verbose` | `-v` | **Verbose Mode**. Enables high-detail logging. Use this to troubleshoot connection issues or to see the raw SQL parser output. |

---

## 3. Config File Priority
MigraGuard follows a strict configuration hierarchy:
1.  **CLI Flag** (Highest Priority)
2.  **Environment Variable** (`MIGRAGUARD_...`)
3.  **Config File** (`migraguard.yaml`)
4.  **Internal Default** (Lowest Priority)
