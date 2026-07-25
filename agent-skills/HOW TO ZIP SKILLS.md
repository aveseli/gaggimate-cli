# How to Zip Skills for Claude Desktop

## Command

From the project root, zip all skills:

```bash
cd /Users/julianleopold/code/gaggimate-mcp/agent-skills && \
  for skill in diagnose feedback gaggimate-profiles knowledge-lookup new-coffee; do \
    zip -r ../${skill}-skill.zip ${skill}/; \
  done
```

Or zip a single skill:

```bash
cd /Users/julianleopold/code/gaggimate-mcp/agent-skills && zip -r ../gaggimate-profiles-skill.zip gaggimate-profiles/
```

## Skill Structure

Each skill is now a single-file SKILL.md that loads references on-demand via MCP resources:

```
gaggimate-profiles/
└── SKILL.md        # References gaggimate://knowledge/profiles/*.md
```

## Output Location

Zip files are created in the project root:
- `diagnose-skill.zip`
- `feedback-skill.zip`
- `gaggimate-profiles-skill.zip`
- `knowledge-lookup-skill.zip`
- `new-coffee-skill.zip`
