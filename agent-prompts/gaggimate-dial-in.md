---
description: Start dialing in a coffee with iterative shot analysis
argument-hint: "[coffee-name]"
---
Let's dial in ${1:-my current coffee} on the Gaggimate.

Steps:
1. List recent shots: `gaggimate shots list --limit 10`
2. Check if there's existing feedback: `gaggimate notes get <id>` for recent shots
3. Analyze the most recent shot: `gaggimate shots analyze <id>`
4. Look for patterns across recent shots (improving/declining/inconsistent?)
5. Ask about taste: sour, bitter, balanced, body, sweetness
6. Provide a prioritized adjustment plan:
   - Primary: most likely fix (usually grind)
   - Secondary: if primary doesn't work
   - Profile: if extraction mechanics need changing
7. After the next shot, repeat the cycle
