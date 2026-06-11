# Local Holdouts Bug Bash - Go SDK

## Overview

This bug bash validates that SDKs correctly evaluate **local holdouts** under
real-world conditions: UI mutations, feature interactions, and edge cases that
automated tests don't cover.

**This is NOT a test suite.** It's an exploration tool for breaking SDKs.

**Key concepts:**
- **Global holdout** (`includedRules: null`): Applies to ALL rules across ALL flags
- **Local holdout** (`includedRules: ["rule_id"]`): Applies only to specified rules
- **Empty includedRules** (`includedRules: []`): Local holdout targeting nothing (NOT global)

**Evaluation priority:**
1. Global holdouts (flag-level, before any rule evaluation)
2. Forced decisions (`SetForcedDecision`) -- per-rule, overrides everything below
3. Local holdouts -- per-rule, after forced decisions
4. Normal experiment/rollout bucketing

## Prerequisites

1. Go 1.21+ installed
2. Access to an Optimizely test environment (likely prod)
3. Go SDK with local holdouts support (PR #451)
4. TDD: https://confluence.sso.episerver.net/display/EXPENG/TDD%3A+Local+Holdouts

## Quick Start

### Static sanity check (no setup, 2 minutes)

Uses bundled datafile with deterministic bucketing. Run this first to verify
the SDK basics work:

```bash
go run examples/local_holdouts/main.go -mode=static -test=all
```

### Live exploration (actual project)

```bash
# Interactive REPL - ad-hoc exploration
go run examples/local_holdouts/main.go -sdk_key=YOUR_KEY -explore

# Guided scenario
go run examples/local_holdouts/main.go -sdk_key=YOUR_KEY -scenario=ui_delete_holdout

# Distribution check
go run examples/local_holdouts/main.go -sdk_key=YOUR_KEY -scenario=distribution -n=500

# List everything
go run examples/local_holdouts/main.go -test=help
```

## Available Scenarios

### UI Mutation Scenarios

These are the most important tests. Make changes in the Optimizely UI while
the SDK is running and verify it handles them correctly.

| Scenario | What to try | What could break |
|----------|------------|-----------------|
| `ui_delete_holdout` | Delete a running holdout | SDK crashes, stale holdout still applied |
| `ui_change_traffic` | Change holdout traffic (50%->0%, 10%->100%) | Wrong bucketing distribution |
| `ui_switch_local_global` | Switch between local and global | Holdout scope doesn't change |
| `ui_change_audience` | Add/remove audience condition | Audience filter not respected |
| `ui_delete_targeted_rule` | Delete the rule a holdout targets | Dangling reference panic |
| `ui_pause_holdout` | Pause a running holdout | Holdout still applied after pause |
| `ui_add_holdout_to_running` | Add holdout to existing experiment | New holdout not picked up |

### Feature Interaction Scenarios

Test how holdouts interact with other SDK features.

| Scenario | What it tests | Expected |
|----------|--------------|----------|
| `forced_vs_holdout` | SetForcedDecision (flag level) overrides holdout | Forced variation wins, holdout is bypassed. After removing forced decision, holdout returns. |
| `forced_rule_level` | SetForcedDecision (rule level) vs holdout | Rule-level forced decision overrides holdout for that rule only. Other rules still evaluate holdouts normally. |
| `holdout_disable_event` | DisableDecisionEvent option with holdout | Decision result is the same with or without the option. No impression event dispatched when disabled. |
| `holdout_track` | Track conversion event after holdout decision | Track call succeeds (no error). Conversion event is sent even for held-out users. |
| `holdout_decide_all` | DecideAll returns holdouts for correct flags | Global holdout: all flags show holdout. Local holdout: only targeted flags show holdout, others are normal. |
| `holdout_decide_for_keys` | DecideForKeys vs DecideAll consistency | For the same user, decisions for requested flags are identical whether from DecideForKeys or DecideAll. |
| `holdout_listener` | Decision notification listener metadata | Listener fires for holdout decisions. Decision info contains holdout ID and rule_type="holdout". |
| `holdout_enabled_flags_only` | EnabledFlagsOnly excludes held-out flags | Held-out flags (enabled=false) are excluded from results. Non-held-out flags are still returned. |

### Stress / Edge Cases

| Scenario | What it tests | Expected |
|----------|--------------|----------|
| `rapid_repolling` | Continuous polling during rapid UI changes | SDK picks up each datafile change. No stale decisions after polling interval. No crashes during rapid updates. |
| `distribution` | Statistical distribution across many users | Holdout rate matches configured traffic %. Distribution is roughly uniform within tolerance. |
| `large_datafile` | Large datafile (100+ flags) with local holdouts (ask Jae Kim) | SDK handles large datafiles without performance degradation. Holdout targeting is still correct across many flags/rules. |
| `many_holdouts` | 10+ local holdouts targeting the same rule | First matching holdout wins (datafile order). Cumulative holdout rate is reasonable. Correct holdout key in decisions. |

### Interactive REPL (`-explore`)

The REPL is already built into `main.go`. When you run:

```bash
go run examples/local_holdouts/main.go -sdk_key=YOUR_KEY -explore
```

It creates the SDK client with polling configured (30s interval), then drops
you into the REPL loop. The client stays alive in the background polling for
datafile changes while you type commands.

You don't need to write any SDK setup code for the REPL -- it's all in this
`main.go` file already. Just transpile the whole thing into your SDK's
language (Python, Java, etc.), and you get both the guided scenarios AND the
REPL for free.

Available commands:

```
decide <flag>              Decide on a flag (with reasons)
decide_all                 Decide all flags
decide_keys <f1> <f2>      Decide for specific flags
dist <flag> [n]            Distribution for n users
user <id>                  Switch user
attr <key> <val>           Set user attribute
force <flag> <var>         Set forced decision (flag level)
force <flag> <rule> <var>  Set forced decision (rule level)
unforce <flag>             Remove forced decision
config                     Show flags/rules from project config
snapshot [n]               Snapshot decisions for n users
```

## Project Setup (Live Mode)

### Step 1: Create Project

Create a new project in your Optimizely environment.

### Step 2: Create Custom Audience

Go to **Audiences** -> **Create New Audience**:
- Name: `Custom Attr Audience`
- Condition: custom attribute `customattr` equals `"yes"`

### Step 3: Create Flags and Rules

Create at least 2 flags with A/B test rules. Set traffic to 100%, audience
to Everyone, activate rules.

| Flag Key | Rule Key (A/B Test) | Variations |
|----------|---------------------|------------|
| `flag_a` | `rule_a` | on, off |
| `flag_b` | `rule_b` | on, off |

### Step 4: Create Holdouts

Start with a simple local holdout, then create more as you explore:

| Holdout Name | Type | Targeted Rules | Traffic | Audience |
|--------------|------|----------------|---------|----------|
| `local_holdout` | Local | `rule_a` only | 50% | Everyone |
| `global_holdout` | Global | All rules | 10% | Everyone |

Activate all holdouts.

### Step 5: Run

```bash
go run examples/local_holdouts/main.go -sdk_key=YOUR_KEY -explore
```

## Transcoding to Other SDKs

This Go implementation is the **reference**. To transcode:

1. Port the static tests first (sanity check with bundled datafile)
2. Port the interactive REPL (most valuable for exploration)
3. Port UI mutation scenarios (guided walkthroughs)
4. Adapt API calls to your SDK's equivalents:
   - `Decide` / `DecideAll` / `DecideForKeys`
   - `SetForcedDecision` / `RemoveForcedDecision`
   - `CreateUserContext` with attributes
   - Decision listener registration
   - `GetOptimizelyConfig` for flag/rule discovery

## Holdout Decision Markers

Use these fields to tell whether a decision is a holdout or a normal
experiment result. When debugging or writing assertions, check for these
values to confirm the SDK is correctly identifying holdout decisions.

When a user is held out:
- `decision.RuleKey` = the holdout key (e.g. `"ho_local_single_rule"`)
- `decision.VariationKey` = `"ho_off_key"`
- `decision.Enabled` = `false`
- Event metadata: `rule_type: "holdout"`, `campaign_id: ""`

## Log Bugs Here

https://optimizely-ext.atlassian.net/browse/FSSDK
