---
description: Analyze an espresso shot with full diagnostics
argument-hint: "<shot-id> [sour|bitter|balanced|...]"
---
Analyze shot ${1:-latest} from my Gaggimate.

${@:+The shot tasted: $ARGUMENTS}

Steps:
1. Run `gaggimate shots analyze ${1:-latest}` to get telemetry
2. If the user provided taste feedback, correlate it with the telemetry
3. Check: resistance level, channeling risk, temperature stability, profile compliance
4. Identify the shot style (classic, bloom, turbo, lever decline, etc.)
5. Provide specific recommendations with reasoning
6. Ask if they want to record feedback with `gaggimate notes set`
