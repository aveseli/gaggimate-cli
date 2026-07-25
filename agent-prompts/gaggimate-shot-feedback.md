---
description: Record feedback for a shot
argument-hint: "<shot-id> <rating> [taste-notes]"
---
Record feedback for shot ${1}.

Rating: ${2:-?}
Notes: ${@:3:-Ask the user for their assessment}

Steps:
1. If the user didn't provide details, ask:
   - How would you rate it (1-5)?
   - Was it sour, bitter, or balanced?
   - Any specific flavors (chocolate, fruit, nutty)?
   - Body (thin, syrupy, watery)?
   - What grind setting did you use?
   - Dose in and dose out?
2. Run: `gaggimate notes set <id> --rating <N> --notes "..." --balance <sour|balanced|bitter>`
3. Add grind/dose if provided: `--grind <setting> --dose-in <g> --dose-out <g>`
4. Confirm what was saved
5. Suggest what to watch for on the next shot
