# gaggimate-cli

CLI tool for interacting with a Gaggimate espresso machine. Built in Go for agents and humans.

## Installation

```bash
cd gaggimate-cli
go build -o gaggimate .
# Or install globally:
go install .
```

## Quick Start

```bash
# Make sure you're on the same network as your Gaggimate
export GAGGIMATE_HOST=gaggimate.local  # or your device IP

# List your shots
gaggimate shots list

# Analyze a shot
gaggimate shots analyze 196

# List profiles
gaggimate profiles list

# Check connectivity
gaggimate diagnose
```

## Commands

### `shots` — Shot History & Analysis

```bash
gaggimate shots list [--limit N]
```
Lists recent shots (newest first) with ID, profile, duration, weight, rating.

```bash
gaggimate shots analyze <ID> [--detail summary|per_phase|per_phase_detailed]
```
Analyzes a shot with physics-informed diagnostics. The `--detail` flag controls depth:

| Level | What you get | Use for |
|-------|-------------|---------|
| `summary` (default) | Key indicators — resistance, channeling risk, temperature stability, profile compliance | Quick triage |
| `per_phase` | Full diagnostics + per-phase breakdown (preinfusion/brew/decline metrics) | Identifying which phase has an issue |
| `per_phase_detailed` | Everything + ~5 averaged samples per phase showing curve shape | Seeing pressure/flow trends |

```bash
gaggimate shots get <ID>
```
Returns raw parsed shot data as JSON (all samples, phases, header info).

### `profiles` — Profile Management

```bash
gaggimate profiles list
gaggimate profiles get <PROFILE_ID>
gaggimate profiles select <PROFILE_ID>
gaggimate profiles delete <PROFILE_ID>  # AI profiles only
```

### `notes` — Shot Notes

```bash
gaggimate notes get <SHOT_ID>
gaggimate notes set <SHOT_ID> --rating 4 --notes "sweet, balanced" \
    --balance balanced --grind 12 --dose-in 18 --dose-out 36
```

### `diagnose` — Connectivity Check

```bash
gaggimate diagnose
```
Tests HTTP API connectivity and shot fetching.

## Configuration

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `GAGGIMATE_HOST` | `gaggimate.local` | Device hostname or IP |
| `GAGGIMATE_PROTOCOL` | `ws` | Protocol (`ws` or `wss`) |
| `GAGGIMATE_TIMEOUT` | `15` | Request timeout (seconds) |

## Shot Diagnostics Reference

The `analyze` command computes these metrics:

### Puck Resistance (P/F²)
The master diagnostic metric. Captures grind fineness, puck prep quality, and channeling.

| Annotation | Meaning |
|-----------|---------|
| `VERY_LOW` | Grind too coarse, under-extraction |
| `MODERATE` | Normal range |
| `HIGH` / `VERY_HIGH` | Grind too fine, possible over-extraction |

### Channeling Risk
Scored 0-8 from 4 independent indicators: flow jitter, target tracking, pressure cliff, late flow runaway.

| Level | Interpretation |
|-------|---------------|
| `LOW` | Clean extraction |
| `MODERATE` | Minor instability — verify with taste |
| `HIGH` | Multiple indicators aligned — channeling likely |
| `VERY_HIGH` | Severe puck degradation |

### Profile Compliance
How well the machine followed the programmed target.

| Metric | Meaning |
|--------|---------|
| `pressure_rmse_bar` | 0 = perfect adherence, >1.5 = POOR |
| `max_pressure_overshoot_bar` | >1.0 bar is highly unusual — grind too fine |

### Temperature Stability
Standard deviation of brew temperature. `VERY_STABLE` (<0.3°C) is ideal.

## Agent Usage Guide

For AI agents analyzing shots:

1. **Start with summary** — `gaggimate shots analyze <ID>`
2. **If channeling is HIGH** — use `--detail per_phase` to find which phase
3. **If you need curve shape** — use `--detail per_phase_detailed` for ~5 samples per phase
4. **Cross-reference with notes** — check `gaggimate notes get <ID>` for taste feedback

### Common Diagnostic Patterns

| Symptom | Diagnostic Signal | Recommendation |
|---------|------------------|----------------|
| Sour taste | Low resistance, high flow, short time | Grind finer |
| Bitter taste | High resistance, long time, high temp | Grind coarser or lower temp |
| Channeling | HIGH/VERY_HIGH channeling risk | Improve puck prep (WDT, distribution) |
| Inconsistent | VOLATILE resistance stability | Check grinder, puck prep |
| Pressure overshoot >1.5 bar | SEVERE_OVERSHOOT annotation | Grind coarser immediately |

## Output Format

All commands output JSON (for machine parsing) or human-readable text (for `diagnose`, error messages).

## License

MIT
