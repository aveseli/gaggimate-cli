---
name: new-coffee
metadata:
  version: 53269c1 (2026-02-23)
description: >
  Research a new coffee bean and propose starting extraction parameters.
  Use when: (1) user shares a new bag of coffee (photo, name, or description),
  (2) user says "new beans", "dialing in a new coffee",
  (3) user asks "where should I start with this coffee".
  Accepts bag photos (extracts info via vision) or text descriptions.
  Researches origin, process, roast level via web search, then recommends
  temperature, grind, ratio, and profile.
---

# New Coffee Research Skill

Systematically research a coffee and propose starting extraction parameters.

*Adapted from [gaggimate-barista](https://github.com/charleshall888/gaggimate-barista) by Charlie Hall.*

## Workflow

### 1. GATHER Coffee Info

**If photo provided:**
- Extract from label: roaster, coffee name, origin, roast date, tasting notes
- Note any visible processing info (washed, natural, etc.)

**If text provided:**
- Parse roaster and coffee name
- Ask for roast date if not mentioned

### 2. REVIEW Brewing Insights

Before researching externally, check what we already know:
- Read brewing insights: `manage_brewing_insights(action="read")`
- Read grind map: `manage_grind_map(action="read")`
- Look for coffees with matching attributes (origin, process, roast level)
- If a similar coffee was brewed before, leverage those learnings as a starting point

### 3. RESEARCH via Web Search

Search for the specific coffee to find:
- Processing method (washed, natural, honey, anaerobic)
- Origin details (country, region, altitude if available)
- Variety (Bourbon, Gesha, Caturra, etc.)
- Roast level (light, medium, dark) — infer from tasting notes if not stated
- Roaster's tasting notes

**See:** `read_knowledge(action="read", filename="research/RESEARCH_CHECKLIST")` for detailed research patterns, origin profiles, and variety extraction guidance.

### 4. SYNTHESIZE Recommendations

Load the relevant knowledge files via MCP tools and build recommendations:
- **Temperature:** From `read_knowledge(action="read", filename="ESPRESSO_BREWING_BASICS")` roast guidelines
- **Pressure:** From `read_knowledge(action="read", filename="PRESSURE_GUIDE")` roast × processing matrix
- **Ratio:** From processing method patterns (washed: 1:2, natural: 1:2-2.5, etc.)
- **Profile:** From `read_knowledge(action="read", filename="PROFILE_LIBRARY")` by roast/process, adjusted for correct pressure
- **Dose:** Based on user's basket size. **Dose = basket size** (e.g., 18g basket → 18g dose). Don't underdose.

### 5. CONFIRM with User

Before finalizing, ask:
> "This [process] [origin] typically shines with [approach]. Would you like to start there, or prefer a more conservative/adventurous approach?"

Options to offer:
- **Conservative:** Classic profile, standard ratio
- **Recommended:** Profile matched to bean characteristics
- **Adventurous:** Bloom profile or turbo shot if appropriate

### 6. UPLOAD Profile (if requested)

Use MCP tool to upload:
```
manage_profile(action="create", profile_name="[Coffee Name] [AI]", temperature=X, phases=[...])
```

Always add `[AI]` suffix to profile names.

### 7. CREATE Coffee Tracking File

After researching and setting up a new coffee, create a coffee tracking file via MCP:
```
manage_coffee(
  action="create",
  coffee_name="[coffee-name]",
  roaster="[roaster]",
  origin="[country, region]",
  process="[washed/natural/honey/anaerobic]",
  roast_level="[light/medium/dark]",
  variety="[if known]",
  roast_date="YYYY-MM-DD",
  bag_size="[e.g. 250g]",
  roaster_notes="[tasting notes from bag/roaster]",
  freshness_note="[e.g. 10 days off roast, in peak window]",
  approach="[Profile name] at [temp]. [Pressure logic from PRESSURE_GUIDE]. Starting at grind [X], [dose]g in, targeting 1:[ratio]. [Why this approach suits this bean — connect origin, process, and roast level to the profile choice. If similar to a previous coffee, reference that experience.]"
)
```

The `approach` field is a narrative paragraph — not a table of numbers. It should explain
*why* you chose this profile/params for this specific bean, connecting the research to
the recommendation. Reference brewing insights from similar coffees if applicable.

This creates a persistent tracking file accessible in all future sessions via `manage_coffee(action="read", coffee_name="<name>")`.

---

## Output Format

```
## Coffee Research: [Name]

### Bean Profile
- **Roaster:** [roaster]
- **Origin:** [country, region]
- **Process:** [washed/natural/honey/anaerobic]
- **Roast Level:** [light/medium/dark]
- **Variety:** [if known]
- **Tasting Notes:** [from roaster]
- **Days Off Roast:** [X days, or "unknown"]

### Recommended Starting Parameters
| Parameter | Value | Reasoning |
|-----------|-------|-----------|
| Temperature | X°C | [roast level rationale] |
| Grind | Start at [general guidance] | [reasoning] |
| Ratio | 1:X | [process rationale] |
| Profile | [name] | [why this profile] |
| Dose | Xg in → Xg out | [basket size rationale] |

### What to Watch For
- [Specific guidance for first shot based on bean characteristics]
- [What taste outcomes to expect]
- [When to adjust and in which direction]
```

---

## Quick Reference

**User says:** "I got a new bag of [coffee]"
**Action:** Extract info → research → recommend → confirm → upload profile

**User shares photo:**
**Action:** Vision extract → research → recommend → confirm → upload profile
