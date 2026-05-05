@echo off
for %%f in (experiments\scenarios\*.yaml) do (
    echo 🔄 Processing: %%~nxf
    call %~dp0gen-db-from-scenario.bat %%f %1
)
echo ✅ All sandboxes seeded.
