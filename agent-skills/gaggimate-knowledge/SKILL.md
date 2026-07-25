---
name: knowledge-lookup
metadata:
  version: 53269c1 (2026-02-23)
description: >
  Answer espresso knowledge questions from the authoritative knowledge files.
  Use when the user asks about: temperature, pressure, ratios, grind settings, freshness,
  extraction theory, puck prep, channeling, baskets, decaf, blends, milk steaming, drink specs,
  profiles, shot styles, or any espresso concept. Routes to the correct knowledge file
  and answers from its content rather than from memory or training data.
---

# Knowledge Lookup Skill

You are answering an espresso knowledge question by consulting the authoritative knowledge files in your context. Do NOT answer from memory or training data — find the answer in the knowledge files first, then respond from their content.

> Adapted from [gaggimate-barista](https://github.com/charleshall888/gaggimate-barista) by Charlie Hall.

## Why This Skill Exists

Knowledge files are available on-demand via the `read_knowledge` MCP tool. This skill ensures you **load and cite the correct file** rather than answering from general training data. The knowledge files contain Gaggimate-specific advice, tested parameters, and opinionated guidance that may differ from generic espresso advice.

## Workflow

### 1. CLASSIFY — Map Question to Knowledge File

Use the routing table below to identify which file(s) to load via `read_knowledge(action="read", filename="...")`. Match on keywords in the user's question.

| Keywords | Primary File | Secondary File |
|----------|-------------|----------------|
| temperature, temp, roast level, how hot | `ESPRESSO_BREWING_BASICS` | — |
| pressure, bar, processing method | `PRESSURE_GUIDE` | — |
| ratio, yield, dose, output, how much | `ESPRESSO_BREWING_BASICS` | — |
| grind, finer, coarser, grind setting | `ESPRESSO_BREWING_BASICS` | `EXTRACTION_SCIENCE` |
| sour, bitter, taste, flavor, tasting | `ESPRESSO_TASTING_GUIDE` | `ESPRESSO_BREWING_BASICS` |
| channeling, puck prep, WDT, distribution | `EXTRACTION_SCIENCE` | — |
| freshness, rest, degas, storage, freeze | `BEAN_FRESHNESS_AND_STORAGE` | — |
| profile, bloom, turbo, lever, allonge | `PROFILE_LIBRARY` | `PRESSURE_GUIDE` |
| decaf, blend, decaffeinated | `SPECIAL_CATEGORIES` | — |
| milk, steam, drink, cortado, cappuccino, latte, flat white | `MILK_AND_DRINKS` | — |
| basket, dose rule, headroom, puck depth | `BASKETS` | — |
| adjust, dial in, what to change, improvement | `ESPRESSO_BREWING_BASICS` | — |
| extraction, TDS, particle, solubility | `EXTRACTION_SCIENCE` | — |

If the question spans two topics (e.g., "what pressure for a natural at light roast?"), load and consult both the primary and secondary files.

If no keywords match, default to `ESPRESSO_BREWING_BASICS` — it covers the broadest range of topics.

#### User Data Tools

If the question is about the user's personal data or history (rather than espresso theory), use the data management tools instead of — or in addition to — knowledge files.

| Keywords | Tool Call | Notes |
|----------|-----------|-------|
| grind map, grind setting history, what worked, what grind | `manage_grind_map(action="read")` | Successful grind settings across coffees |
| my setup, my equipment, my machine, my grinder, my basket | `manage_user_setup(action="read")` | Equipment, preferences, puck prep routine |
| history with, last time I brewed, previous shots, this coffee, how did [coffee] go | `manage_coffee(action="read", coffee_name="...")` | Per-coffee brewing journal and insights |
| patterns, what works for, experience with, similar coffees, learnings | `manage_brewing_insights(action="read")` | Cross-coffee patterns and accumulated learnings |

To find the right coffee file, first list available coffees via `manage_coffee(action="list")`, then load the specific one via `manage_coffee(action="read", coffee_name="...")`.

### 2. ANSWER from File Content

Load the identified knowledge file(s) via `read_knowledge(action="read", filename="...")`, then answer the user's question using their content. **Cite specific tables, thresholds, or sections** from the file. Do not paraphrase from memory — use the actual data.

**Good:** "The pressure matrix in PRESSURE_GUIDE recommends 7-8 bar for light roast naturals."
**Bad:** "Light roast naturals generally do well at lower pressure." (no file reference, no specific value)

### 3. CROSS-REFERENCE (if needed)

If the answer touches multiple topics, check the cross-references at the bottom of each knowledge file. These point to related files that may have additional relevant information.

**Stop rule:** Consult at most **2 knowledge files** per question. If the primary file answers it, stop there.

---

## Quick Reference

**User asks:** "What temperature for a light roast?"
**Action:** Consult ESPRESSO_BREWING_BASICS.md → cite Temperature Guidelines table → answer "94-96°C"

**User asks:** "What pressure for an anaerobic natural?"
**Action:** Consult PRESSURE_GUIDE.md → cite Comprehensive Pressure Matrix → answer with specific bar range

**User asks:** "How do I store beans long-term?"
**Action:** Consult BEAN_FRESHNESS_AND_STORAGE.md → answer from storage section

**User asks:** "What basket should I use for 18g?"
**Action:** Consult BASKETS.md → cite dose rules → answer with basket recommendation

**User asks:** "My shot tastes sour AND bitter"
**Action:** Consult ESPRESSO_TASTING_GUIDE.md + ESPRESSO_BREWING_BASICS.md → cite channeling rule → recommend puck prep fix
