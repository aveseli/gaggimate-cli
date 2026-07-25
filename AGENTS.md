# Agent Guide: gaggimate-cli

This guide explains how an AI agent should use `gaggimate-cli` to analyze espresso shots, manage profiles, and help users dial in their coffee.

## Setup

### Install Skills and Prompts

```bash
gaggimate-cli install --harness pi          # Global install
gaggimate-cli install --harness pi --local  # Project-local
gaggimate-cli install --harness claude      # Claude Desktop
```

This installs:
- **6 skills** (gaggimate-core, gaggimate-diagnose, gaggimate-feedback, gaggimate-profiles, gaggimate-knowledge, gaggimate-new-coffee)
- **4 prompt templates** (Pi only: /gaggimate-analyze-shot, /gaggimate-dial-in, /gaggimate-new-coffee, /gaggimate-shot-feedback)

### Prerequisites

- The CLI is installed and accessible (check `gaggimate version`)
- The Gaggimate device is on the same network as the CLI
- Set `GAGGIMATE_HOST` env var if not using `gaggimate.local`

## CI/CD: Building Releases

Pushing a version tag triggers GitHub Actions to cross-compile binaries and publish a GitHub Release.

```bash
git tag v1.0.0
git push origin v1.0.0
```

**Build matrix:**

| Platform | Binary |
|----------|--------|
| Linux x86_64 | `gaggimate-linux-amd64` |
| Linux ARM64 | `gaggimate-linux-arm64` |
| macOS Intel | `gaggimate-darwin-amd64` |
| macOS Apple Silicon | `gaggimate-darwin-arm64` |
| Windows x86_64 | `gaggimate-windows-amd64.exe` |

**Workflow:** `.github/workflows/release.yml`

Binaries are built with `-trimpath -ldflags="-s -w"` and `CGO_ENABLED=0` for fully static, stripped builds. The release includes auto-generated release notes.

## Quick Reference

```bash
gaggimate shots list                       # Recent shots
gaggimate shots analyze <ID>               # Analyze with diagnostics
gaggimate shots analyze <ID> --detail per_phase   # Deeper analysis
gaggimate profiles list                    # List profiles
gaggimate profiles select <ID>             # Select active profile
gaggimate notes get <ID>                   # Get shot notes
gaggimate notes set <ID> --rating 4 --notes "..."  # Save notes
```

## Skills

Skills are loaded on-demand when espresso topics are detected:

| Skill | Purpose |
|-------|---------|
| gaggimate-core | Barista persona + workflow overview (entry point) |
| gaggimate-diagnose | Shot telemetry analysis with taste correlation |
| gaggimate-feedback | Shot feedback loop: gather, analyze, record, recommend |
| gaggimate-profiles | Profile creation with phase/pump/transition guidance |
| gaggimate-knowledge | Espresso knowledge Q&A (pressure, extraction, tasting) |
| gaggimate-new-coffee | Research beans, recommend parameters, upload profile |

## Prompt Templates (Pi only)

Type `/name` in the editor to invoke:

| Template | Description |
|----------|-------------|
| /gaggimate-analyze-shot | Analyze a shot with full diagnostics |
| /gaggimate-dial-in | Start iterative dialing workflow |
| /gaggimate-new-coffee | Research and set up a new coffee |
| /gaggimate-shot-feedback | Record tasting feedback |

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
