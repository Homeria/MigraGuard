#!/bin/bash
for f in experiments/scenarios/*.yaml; do
    echo "🔄 Processing: $(basename $f)"
    "$(dirname "$0")/gen-db-from-scenario.sh" "$f" "$1"
done
echo "✅ All sandboxes seeded."
