SELECT
    to_char(clock_timestamp(), 'YYYY-MM-DD"T"HH24:MI:SS.MS') AS measured_at,
    count(*) FILTER (
        WHERE backend_type = 'client backend'
          AND datname = current_database()
          AND pid <> pg_backend_pid()
    ) AS connections,
    count(*) FILTER (
        WHERE backend_type = 'client backend'
          AND datname = current_database()
          AND state = 'active'
          AND pid <> pg_backend_pid()
    ) AS active_connections,
    count(*) FILTER (
        WHERE backend_type = 'client backend'
          AND datname = current_database()
          AND wait_event_type = 'Lock'
          AND pid <> pg_backend_pid()
    ) AS lock_waiters,
    coalesce((SELECT xact_commit FROM pg_stat_database WHERE datname = current_database()), 0) AS xact_commit,
    coalesce((SELECT xact_rollback FROM pg_stat_database WHERE datname = current_database()), 0) AS xact_rollback,
    coalesce((SELECT deadlocks FROM pg_stat_database WHERE datname = current_database()), 0) AS deadlocks
FROM pg_stat_activity;
