---
name: feedback
metadata:
  version: b6a1883 (2026-02-25)
description: >
  Gather shot feedback, analyze extraction, recommend adjustments, and record results.
  Use when user says: "/feedback", "I just pulled a shot", "how was that", "it tasted [sour/bitter/flat/good]",
  provides a star rating, shares taste observations, or asks "what should I adjust" after a shot.
  Owns the full shot feedback loop: gathering, analysis, tasting notes, and drink format.
---

# Shot Feedback & Dialing Skill

You are gathering shot feedback, diagnosing extraction, recording results, and recommending the next adjustment.

> Adapted from [gaggimate-barista](https://github.com/charleshall888/gaggimate-barista) by Charlie Hall.

## Context

Knowledge files are available on-demand via the `read_knowledge` MCP tool. Load only what's needed:
- `read_knowledge(action="read", filename="ESPRESSO_BREWING_BASICS")` — adjustment strategies, diagnostic decision tree, variable hierarchy
- `read_knowledge(action="read", filename="ESPRESSO_TASTING_GUIDE")` — sour vs bitter diagnosis, tasting methodology
- `read_knowledge(action="read", filename="PRESSURE_GUIDE")` — when feedback suggests pressure/profile style change
- `read_knowledge(action="read", filename="MILK_AND_DRINKS")` — when shot is dialed in and user wants drink recommendations

---

## Workflow

### 1. GATHER Context

- Read `manage_user_setup(action="read")` to load the user's equipment, basket size, and preferences.
- Read `manage_coffee(action="list")` to check for existing coffee tracking files, then load the relevant one via `manage_coffee(action="read", coffee_name="...")`.
- If no coffee context exists: ask the user what coffee they're brewing before proceeding.
- **Stale check:** If roast date is 30+ days old, gently ask if user is still on this bag.

### 2. COLLECT Feedback

Gather from the user (ask for what's missing):

| Field | Required | Notes |
|-------|----------|-------|
| **Rating** (1-5 stars) | Yes | Overall satisfaction |
| **Balance** (sour/balanced/bitter) | Yes | Primary extraction indicator |
| **Observations** | Yes (1+ specific note) | Body, sweetness, finish, flavor, mouthfeel |
| **Grind setting** | Ask if not offered | Important for tracking |
| **Dose in** | Ask if not offered | Should match basket size |
| **Dose out** | Ask if not offered | Output weight — needed for ratio |
| **Shot ID** | Optional | From `list_recent_shots` if user doesn't provide |

**Minimum viable feedback:** Rating + balance + one specific observation.

**Dose in & dose out:** The brew ratio (dose in : dose out) is important for diagnosis and tracking. Make sure both are captured:
- For dose out: if a BT scale is connected, check `final_weight_g` in `analyze_shot` output first. If telemetry is unavailable or looks unreliable, ask the user.
- For dose in: ask if not volunteered — it should match the basket size from `manage_user_setup`.
- If the user doesn't mention either, ask casually: "What were your dose in and output weight?"
- Log both in shot notes via `dose_in` and `dose_out` parameters.

**Scale data handling:** Prefer telemetry for dose out when a shot ID is available (see `read_knowledge(action="read", filename="diagnostics/TELEMETRY_PATTERNS")` for scale artifact detection). If telemetry is unavailable or looks unreliable, just ask the user.

### 3. ANALYZE & RECOMMEND

Load these knowledge files for diagnosis:
- `read_knowledge(action="read", filename="ESPRESSO_BREWING_BASICS")` — variable hierarchy (grind → ratio → temp → pressure → puck prep) and diagnostic decision tree
- `read_knowledge(action="read", filename="ESPRESSO_TASTING_GUIDE")` — taste diagnosis and the Scott Rao channeling rule (sour + bitter = channeling → fix puck prep, NOT grind)
- `read_knowledge(action="read", filename="diagnostics/DIAGNOSTIC_TREES")` — full diagnostic decision trees for complex cases

Apply the diagnostic rules to the user's feedback. Always explain *why* you're suggesting a change. One primary recommendation, one backup.

### 4. RECORD

Do all of these automatically after feedback is collected:

#### 4a. Shot Notes → Device (via MCP)

If a shot ID is available, sync feedback to the device:
```
manage_shot_notes(shot_id, action="update", rating=X, balance_taste="...", notes="...", grind_setting="...", dose_in=X, dose_out=X)
```

#### 4b. Coffee Tracking (via MCP)

Log a brewing journal entry to the coffee's tracking file — analysis, not raw numbers:
```
manage_coffee(
  action="log_entry",
  coffee_name="[coffee-name]",
  entry_date="YYYY-MM-DD",
  entry_headline="Grind [X], [Profile Name] — [X/5]",
  entry_body="[Agent's analysis: what worked, what didn't, taste description, what to try next. Reference shot ID if available.]"
)
```

The journal entry should capture *thinking*, not numbers. Include:
- What the shot tasted like and why (connect to extraction theory)
- What worked well and what didn't
- Specific recommendation for next shot
- Shot ID reference if available (for `analyze_shot` later)

If no coffee tracking file exists yet, create one first:
```
manage_coffee(
  action="create",
  coffee_name="[coffee-name]",
  roaster="...",
  origin="...",
  process="...",
  roast_level="...",
  roaster_notes="[tasting notes from bag]",
  approach="[Profile name] at [temp]. [Pressure logic reasoning]. Starting at grind [X], [dose]g in, targeting 1:[ratio]. [Why this approach suits this bean.]"
)
```

#### 4c. Grind Map (if 4+ stars)

**If 4+ stars AND grind setting provided**, add to the grind map:
```
manage_grind_map(
  action="add_entry",
  coffee="[name]",
  roast="[light/medium/dark]",
  process="[washed/natural/honey/anaerobic]",
  origin="[country/region]",
  days_off_roast="[X days]",
  grind="[setting]",
  profile="[name]",
  ratio="1:X",
  temp="X°C",
  rating="X/5",
  date="YYYY-MM-DD"
)
```

#### 4d. Brewing Insights (when patterns emerge)

When a meaningful pattern emerges (not after every shot), update the cross-coffee insights:
1. Read current insights: `manage_brewing_insights(action="read")` (or init via `manage_brewing_insights(action="init")` if it doesn't exist)
2. Update with the new learning:
```
manage_brewing_insights(
  action="write",
  content="[updated full markdown with the new insight added to the appropriate section]"
)
```

**When to update brewing insights:**
- Coffee is dialed in (4+ stars, balanced) — record what worked and why
- A clear pattern emerges (e.g. "declining pressure works for all medium-roast honeys we've tried")
- A surprising finding worth remembering across coffees

**When NOT to update:** After every single shot. Only when there's a genuine cross-coffee learning.

### 5. SUGGEST Next Steps

Based on the analysis:

**If still dialing in (rating < 4 or not balanced):**
- State the specific change for the next shot
- Explain what to watch for ("Time to first drip should increase" / "Look for more body")

**If dialed in (rating 4+ AND balanced):**
- Celebrate briefly
- Recommend a drink format based on shot character:

| Shot Character | Recommended Format |
|----------------|-------------------|
| Bright, fruity, delicate | Cortado or piccolo |
| Sweet, balanced, medium body | Cappuccino or flat white |
| Intense, heavy body | Latte |
| Clarity-focused, tea-like | Cortado or piccolo |

**Core principle:** Extract for the bean's best expression first, then match the drink format. Never adjust grind/ratio/pressure/temp to "make the shot work in milk."

If user wants full milk science, steaming technique, or drink recipes → reference MILK_AND_DRINKS.md knowledge file.

---

## Integration with Other Skills

- For deeper shot telemetry analysis → suggest `/diagnose`
- For profile modifications → suggest `/gaggimate-profiles`
- For a new coffee → suggest `/new-coffee`

---

## Quick Reference

**User says:** "3 stars, sour, grind 12, 22g in"
**Action:** Gather context → analyze → record journal entry (analysis, not numbers) → diagnose (sour = extract more) → recommend grind/yield change

**User says:** "5 stars, balanced, amazing sweetness"
**Action:** Gather context → celebrate → record journal entry → add to grind map → update brewing insights if pattern emerged → recommend drink format

**User says:** "it was sour AND bitter"
**Action:** Gather context → diagnose **channeling** → recommend puck prep fix, NOT grind change → record journal entry
