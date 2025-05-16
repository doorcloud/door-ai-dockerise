#!/usr/bin/env bash
set -eo pipefail

EXAMPLES_DIR=${1:-spring-boot-examples}
LOG_DIR=./test_results
mkdir -p "$LOG_DIR"

total=0
ok=0
skip=0
fail=0

# run offline so the test is deterministic and fast
export OPENAI_MOCK=${OPENAI_MOCK:-1}

is_war() {
    grep -q "<packaging>war" "$1/pom.xml" 2>/dev/null
}

for demo in "$EXAMPLES_DIR"/spring-boot-*; do
    [ -d "$demo" ] || continue
    name=$(basename "$demo")
    
    if is_war "$demo"; then
        printf "⏭  %-35s (WAR – not supported yet)\n" "$name"
        skip=$((skip+1))
        continue
    fi
    
    total=$((total+1))
    printf "🔄 %-35s … " "$name"
    
    if go run ./cmd/dockergen "$demo" \
        >"$LOG_DIR/$name.log" 2>&1; then
        printf "✅\n"
        ok=$((ok+1))
    else
        printf "❌ (see %s)\n" "$LOG_DIR/$name.log"
        fail=$((fail+1))
    fi
done

printf "\nSummary: %d ok, %d skipped (WAR), %d failed, %d total\n" \
    "$ok" "$skip" "$fail" "$((ok+skip+fail))"

exit $fail 