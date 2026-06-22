param(
    [string]$Database = "benchmark_run",
    [int]$ExtraOrders = 0
)

$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$initSql = Join-Path $repoRoot "build\postgres\init-db.sql"

function Assert-NativeCommand([string]$Step) {
    if ($LASTEXITCODE -ne 0) {
        throw "$Step failed with exit code $LASTEXITCODE."
    }
}

Push-Location $repoRoot
try {
    docker compose up -d db | Out-Host
    Assert-NativeCommand "Starting PostgreSQL"
    "ALTER ROLE `"user`" WITH PASSWORD 'pass';" | docker compose exec -T db psql -U user -d postgres -v ON_ERROR_STOP=1 | Out-Host
    Assert-NativeCommand "Synchronizing benchmark database credentials"
    docker compose exec -T db psql -U user -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS $Database WITH (FORCE);" | Out-Host
    Assert-NativeCommand "Dropping benchmark database"
    docker compose exec -T db psql -U user -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE $Database;" | Out-Host
    Assert-NativeCommand "Creating benchmark database"
    Get-Content $initSql -Raw | docker compose exec -T db psql -U user -d $Database -v ON_ERROR_STOP=1 | Out-Host
    Assert-NativeCommand "Loading benchmark fixtures"
    if ($ExtraOrders -gt 0) {
        $extraOrdersSql = @"
INSERT INTO orders (order_no, user_id, total_amount, status, order_details)
SELECT
    'EXTRA-' || i || '-' || floor(random()*1000000),
    floor(random() * 1000) + 1,
    (random() * 500 + 10)::decimal(19,4),
    'PAID',
    '{"source": "benchmark", "campaign": "extra_load"}'::jsonb
FROM generate_series(1, $ExtraOrders) s(i);
"@
        $extraOrdersSql | docker compose exec -T db psql -U user -d $Database -v ON_ERROR_STOP=1 | Out-Host
        Assert-NativeCommand "Loading extra benchmark orders"
    }
    docker compose exec -T db psql -U user -d $Database -v ON_ERROR_STOP=1 -c "ANALYZE;" | Out-Host
    Assert-NativeCommand "Analyzing benchmark fixtures"
    docker compose exec -T db psql -U user -d $Database -v ON_ERROR_STOP=1 -c "SELECT pg_stat_statements_reset();" | Out-Host
    Assert-NativeCommand "Resetting PostgreSQL statement statistics"
}
finally {
    Pop-Location
}

Write-Host "Benchmark database '$Database' is ready."
