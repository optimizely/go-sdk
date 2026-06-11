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
2. Access to Optimizely RC (Prep) environment
3. Go SDK with local holdouts support (PR #451)
4. TDD: https://confluence.sso.episerver.net/display/EXPENG/TDD%3A+Local+Holdouts

## Quick Start

### Static sanity check (no setup, 2 minutes)

Uses bundled datafile with deterministic bucketing. Run this first to verify
the SDK basics work:

```bash
go run examples/local_holdouts/main.go -mode=static -test=all
```

### Live exploration (requires RC project)

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

| Scenario | What it tests |
|----------|--------------|
| `forced_vs_holdout` | SetForcedDecision (flag level) overrides holdout |
| `forced_rule_level` | SetForcedDecision (rule level) vs holdout |
| `holdout_disable_event` | DisableDecisionEvent option with holdout |
| `holdout_track` | Track conversion event after holdout decision |
| `holdout_decide_all` | DecideAll returns holdouts for correct flags |
| `holdout_decide_for_keys` | DecideForKeys vs DecideAll consistency |
| `holdout_listener` | Decision notification listener metadata |
| `holdout_enabled_flags_only` | EnabledFlagsOnly excludes held-out flags |

### Stress / Edge Cases

| Scenario | What it tests |
|----------|--------------|
| `rapid_repolling` | Continuous polling during rapid UI changes |
| `distribution` | Statistical distribution across many users |

### Interactive REPL (`-explore`)

Free-form exploration mode with commands:

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

Create a new project in the RC (Prep) environment.

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

When a user is held out:
- `decision.RuleKey` = the holdout key (e.g. `"ho_local_single_rule"`)
- `decision.VariationKey` = `"ho_off_key"`
- `decision.Enabled` = `false`
- Event metadata: `rule_type: "holdout"`, `campaign_id: ""`

## Log Bugs Here

https://optimizely-ext.atlassian.net/browse/FSSDK
