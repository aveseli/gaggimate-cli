# Agent Guide: gaggimate-cli

This guide explains how an AI agent should use `gaggimate-cli` to analyze espresso shots, manage profiles, and help users dial in their coffee.

## Prerequisites

- The CLI is installed and accessible (check `gaggimate version`)
- The Gaggimate device is on the same network as the CLI
- Set `GAGGIMATE_HOST` env var if not using `gaggimate.local`

## Core Workflow: Analyzing a Shot

### Step 1: Quick Assessment

```bash
gaggimate shots analyze <SHOT_ID>
```

This returns JSON with:
- `summary` — key metrics with band annotations
- `phases` — phase list with stats
- `diagnostics` — summary-level diagnostics
- `notes` — any tasting notes from the device

**What to look for first:**
1. `diagnostics.channeling_risk` — if HIGH or VERY_HIGH, investigate
2. `diagnostics.resistance_avg` and annotations.resistance_level
3. `summary.pressure.avg_bar` — is it in the expected range for the style?
4. `summary.flow.total_volume_ml` — does the yield match expectations?
5. `diagnostics.annotations.pressure_overshoot` — SEVERE means grind too fine

### Step 2: Deeper Investigation (if needed)

If the summary shows issues, go deeper:

```bash
gaggimate shots analyze <SHOT_ID> --detail per_phase
```

This adds per-phase diagnostics. Look at the `phases` array to see which phase has the problem:
- Preinfusion: ramp rate, saturation time
- Brew: resistance, channeling risk per phase
- Decline: taper smoothness

### Step 3: See Curve Shape (if needed)

```bash
gaggimate shots analyze <SHOT_ID> --detail per_phase_detailed
```

This adds ~5 averaged samples per phase showing the pressure/flow trajectory.

## Diagnostic Interpretation

### Grind Too Fine
- `resistance.annotations.level` = HIGH or VERY_HIGH
- `profile_compliance.max_pressure_overshoot_bar` > 1.0
- `flow.avg_flow_ml_s` below expected for style
- **Fix:** Grind coarser

### Grind Too Coarse
- `resistance.annotations.level` = LOW or VERY_LOW
- Pressure never reaches target (`max_pressure_undershoot_bar` high)
- `flow.avg_flow_ml_s` above expected
- `extraction.total_time_s` too short
- **Fix:** Grind finer

### Channeling
- `channeling.channeling_risk` = HIGH or VERY_HIGH
- `channeling.annotations.primary_signal` shows which indicators fired
- `resistance.annotations.erosion` = STEEP_DECLINE
- **Fix:** Improve puck prep (WDT, distribution, leveling)

### Temperature Issues
- `temperature.stability.annotations.stability` = UNSTABLE
- `temperature.overshoot_c` > 2.0
- **Fix:** Check PID, pre-heat machine longer

### Good Shot
- `resistance.annotations.level` = MODERATE
- `resistance.annotations.stability` = STABLE or VERY_STABLE
- `channeling.channeling_risk` = LOW
- `profile_compliance.annotations.pressure_adherence` = EXCELLENT or GOOD

## Profile Management

### List Available Profiles
```bash
gaggimate profiles list
```

### Select a Profile for Next Shot
```bash
gaggimate profiles select <PROFILE_ID>
```

### Get Profile Details
```bash
gaggimate profiles get <PROFILE_ID>
```

## Recording Feedback

After the user tastes a shot, record notes:

```bash
gaggimate notes set <SHOT_ID> \
    --rating 4 \
    --notes "sweet chocolate, balanced body" \
    --balance balanced \
    --grind 12 \
    --dose-in 18 \
    --dose-out 36
```

## Shot History Analysis

To look at trends across multiple shots:

```bash
gaggimate shots list --limit 20
```

This returns JSON array of recent shots. Look for:
- Rating trends (improving or declining?)
- Duration changes (getting longer = finer grind?)
- Volume consistency

## Context Gathering

Before making recommendations, gather context:

1. **Recent shots:** `gaggimate shots list --limit 5`
2. **Target shot analysis:** `gaggimate shots analyze <ID>`
3. **Notes/taste feedback:** `gaggimate notes get <ID>`
4. **Current profile:** `gaggimate profiles get <PROFILE_ID>`

## Common Issues to Watch For

| Observation | Likely Cause | Action |
|-------------|-------------|--------|
| Shots getting progressively longer | Grinder drift or bean aging | Grind slightly coarser |
| Inconsistent shot times | Puck prep issues | Focus on WDT and distribution |
| Channeling risk HIGH on every shot | Grinder producing fines | Check grinder burrs |
| Temperature UNSTABLE | Machine not heated | Pre-heat longer, check PID |
| Pressure never reaches 9 bar | Grind too coarse or pump issue | Grind finer first |

## Output Format Notes

- All shot analysis output is JSON
- Numeric values are rounded to 2 decimal places
- Band annotations provide human-readable labels
- Timestamps are Unix epoch (convert with `date -d @<timestamp>`)
- Shot IDs are zero-padded to 6 digits in the API (e.g., "000196")
