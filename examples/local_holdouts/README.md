# Local Holdouts Bug Bash - Go SDK

## Overview

This bug bash validates that the Go SDK correctly evaluates **local holdouts** --
holdouts that target specific experiment/delivery rules rather than applying
globally to all rules across all flags.

**Key concepts:**
- **Global holdout** (`includedRules: null` in datafile): Applies to ALL rules across ALL flags. Evaluated first at the flag level.
- **Local holdout** (`includedRules: ["rule_id_1", ...]`): Applies only to the specified rules. Evaluated per-rule, after forced decisions, before audience/traffic checks.

**Evaluation priority (highest to lowest):**
1. Global holdouts (flag-level, before any rule evaluation)
2. Forced decisions (`SetForcedDecision`) -- per-rule, overrides everything below
3. Local holdouts -- per-rule, after forced decisions
4. Normal experiment/rollout bucketing

**When a user is held out:**
- `decision.RuleKey` = the holdout key (e.g. `"ho_local_single_rule"`)
- `decision.VariationKey` = `"ho_off_key"`
- `decision.Enabled` = `false`
- Event metadata: `rule_type: "holdout"`, `campaign_id: ""`

## Prerequisites

1. Go 1.21+ installed
2. Access to Optimizely RC (Prep) environment
3. Go SDK with local holdouts support (PR #451, already merged to master)
4. TDD: https://confluence.sso.episerver.net/display/EXPENG/TDD%3A+Local+Holdouts

## Running Tests

There are two modes: **live** (with RC project) and **static** (with bundled datafile).

### Static datafile mode (quick validation, no project setup needed)

Uses the bundled `local_holdouts.json` datafile with pre-calculated bucketing IDs:

```bash
# From go-sdk root
go run examples/local_holdouts/main.go -mode=static -test=basic
go run examples/local_holdouts/main.go -mode=static -test=all
```

### Live mode (full bug bash with UI interaction)

Uses your RC (Prep) project via SDK key:

```bash
# Update SDK_KEY in main.go first (or pass via flag)
go run examples/local_holdouts/main.go -sdk_key=YOUR_KEY -test=basic
go run examples/local_holdouts/main.go -sdk_key=YOUR_KEY -test=distribution
```

### Available test cases

```
static mode tests (deterministic, pre-calculated bucketing):
  basic              Local holdout targeting single rule
  multi_rule         Local holdout targeting multiple rules on same flag
  cross_flag         Local holdout targeting rules across different flags
  global             Global holdout applies to all rules
  precedence         Global holdout beats local holdout
  audience           Local holdout with audience conditions
  not_targeted       Local holdout only affects targeted rules
  zero_traffic       Zero-traffic holdout never applied

live mode tests (requires RC project):
  live_basic         Basic local holdout with live project
  live_global        Global holdout with live project
  live_forced        SetForcedDecision overrides holdout
  live_decide_all    DecideAll with holdouts
  live_distribution  Statistical distribution over 1000 users
  live_ui_refresh    Interactive: change UI, verify SDK picks up changes

  all                Run all tests for current mode
```

## Project Setup (Live Mode Only)

### Step 1: Create Project

Create a new project in the RC (Prep) environment.

### Step 2: Create Custom Audience

Go to **Audiences** -> **Create New Audience**:
- Name: `Custom Attr Audience`
- Condition: custom attribute `customattr` equals `"yes"`

### Step 3: Create Flags and Rules

Create the following flags. For each, add the A/B test rule shown below,
set the variation to "On", set traffic to 100%, audience to Everyone,
and activate (Run rule + Run ruleset).

| Flag Key    | Rule Key (A/B Test) | Variations |
|-------------|---------------------|------------|
| `flag_a`    | `rule_a`            | on, off    |
| `flag_b`    | `rule_b`            | on, off    |

### Step 4: Create Holdouts

Create holdouts with the following configuration:

| Holdout Name           | Type   | Targeted Rules | Traffic | Audience            |
|------------------------|--------|----------------|---------|---------------------|
| `local_holdout`        | Local  | `rule_a` only  | 50%     | Everyone            |
| `global_holdout`       | Global | All rules      | 10%     | Everyone            |
| `audience_holdout`     | Local  | `rule_b` only  | 50%     | Custom Attr Audience|

Activate all holdouts.

### Step 5: Update Code

1. Copy your SDK Key from **Settings** -> **Environments** -> **Development**
2. Pass it via `-sdk_key=YOUR_KEY` or update the `SDK_KEY` constant in `main.go`

## Holdout Datafile Structure Reference

When a holdout is created in the UI, it appears in the datafile JSON under `"holdouts"`:

```json
{
  "holdouts": [
    {
      "id": "ho_local_single_rule_id",
      "key": "ho_local_single_rule",
      "status": "Running",
      "audienceIds": [],
      "audienceConditions": [],
      "trafficAllocation": [
        { "endOfRange": 3000, "entityId": "$opt_dummy_variation_id" }
      ],
      "variations": [
        { "featureEnabled": false, "id": "$opt_dummy_variation_id", "key": "ho_off_key" }
      ],
      "includedRules": ["5001"]
    },
    {
      "id": "ho_global_all_rules_id",
      "key": "ho_global_all_rules",
      "status": "Running",
      "trafficAllocation": [
        { "endOfRange": 1500, "entityId": "$opt_dummy_variation_id" }
      ],
      "variations": [
        { "featureEnabled": false, "id": "$opt_dummy_variation_id", "key": "ho_off_key" }
      ],
      "includedRules": null
    }
  ]
}
```

Key differences:
- `includedRules: null` = **global** holdout (all rules)
- `includedRules: ["rule_id"]` = **local** holdout (specific rules only)
- Holdout variations always have `featureEnabled: false` and key `"ho_off_key"`

## Debugging

### Enable debug logging
Debug logging is enabled by default in the test code. Look for these log patterns:

- Holdout evaluation: `"User bucketed into holdout"` / `"User not in holdout"`
- Audience evaluation: `"User meets audience conditions for holdout"`
- Bucketing: `"Assigned bucket"` with holdout ID

### Verify via decision reasons
Use `-test=basic` which runs with `INCLUDE_REASONS` option. The `reasons` field in the
decision output shows the full evaluation path.

### Check dispatched events
Holdout impression events should have:
- `rule_type: "holdout"`
- `campaign_id: ""` (empty)
- `rule_key: <holdout_key>`

Normal experiment events have:
- `rule_type: "feature-test"`
- `campaign_id: <layer_id>`

## Validation Checklist

- [ ] Local holdout applies only to targeted rules
- [ ] Local holdout does NOT affect non-targeted rules on same flag
- [ ] Local holdout does NOT affect rules on other flags (unless explicitly targeted)
- [ ] Cross-flag local holdout works across different flags
- [ ] Global holdout applies to ALL rules on ALL flags
- [ ] Global holdout is evaluated BEFORE local holdouts
- [ ] Audience conditions on holdouts are respected
- [ ] Zero-traffic holdout is never applied
- [ ] Forced decisions override holdout decisions
- [ ] DecideAll and DecideForKeys return correct holdout results
- [ ] Impression events contain holdout metadata (rule_type=holdout)
- [ ] Decision listener provides holdout info (experiment_id=holdout_id)
- [ ] UI changes to holdout settings reflect in SDK after datafile refresh

## Log Bugs Here
<!-- Update with your team's bug tracking link -->
https://optimizely-ext.atlassian.net/browse/FSSDK

## Notes

1. **Backward compatibility**: Old datafiles without `includedRules` field treat all holdouts as global (same as `includedRules: null`).

2. **Empty includedRules**: `includedRules: []` (empty array) is a local holdout that targets NO rules -- effectively disabled. This is NOT the same as `null` (global).

3. **Multiple holdouts on same rule**: When multiple local holdouts target the same rule, they are evaluated in datafile order. First match wins.

4. **Holdout variation**: Holdout variations always have `featureEnabled: false`. The variation key is `"ho_off_key"`. Variables return their default values.
