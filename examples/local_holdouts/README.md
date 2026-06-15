# Local Holdouts Bug Bash -- Go SDK

## What is this?

A hands-on exploration tool for testing local holdouts in the SDK. This is
**not an automated test suite** -- it's a starting point for manual QA.

`main.go` contains working SDK code with commented-out blocks you can
uncomment, modify, and extend. The goal is to try unusual things and find
bugs that automated tests miss.

## How local holdouts work

Holdouts exclude users from experiments so you can measure the overall impact
of running experiments. They come in two types:

| Type | Datafile field | Behavior |
|------|---------------|----------|
| **Global** | `includedRules: null` | Applies to ALL rules on ALL flags |
| **Local** | `includedRules: ["rule_id"]` | Applies only to the specified rules |
| **Empty local** | `includedRules: []` | Targets nothing, effectively disabled (NOT global) |

**Evaluation priority** (highest to lowest):
1. Global holdouts -- flag-level, before any rule evaluation
2. Forced decisions (`SetForcedDecision`) -- per-rule, overrides everything below
3. Local holdouts -- per-rule, after forced decisions
4. Normal experiment/rollout bucketing

**When a user is held out**, the decision looks like this:
- `decision.RuleKey` = the holdout key (e.g. `"my_holdout"`)
- `decision.VariationKey` = `"ho_off_key"`
- `decision.Enabled` = `false`
- Impression event metadata: `rule_type: "holdout"`, `campaign_id: ""`

Use these fields to tell whether a decision is a holdout or a normal
experiment result when debugging.

## Quick start

### 1. Static sanity check (no setup needed)

Runs 7 automated checks using the bundled datafile to verify the SDK basics
work. Do this first.

```bash
go run examples/local_holdouts/main.go -mode=static
```

### 2. Set up your project

Create or reuse a project in your Optimizely environment:

1. **Create a custom audience:**
   - Name: `Custom Attr Audience`
   - Condition: custom attribute `customattr` equals `"yes"`

2. **Create flags with rules:**

   | Flag Key | Rule Key (A/B Test) | Variations | Traffic | Audience |
   |----------|---------------------|------------|---------|----------|
   | `flag_a` | `rule_a` | on, off | 100% | Everyone |
   | `flag_b` | `rule_b` | on, off | 100% | Everyone |

3. **Create holdouts:**

   | Holdout Name | Type | Targeted Rules | Traffic | Audience |
   |--------------|------|----------------|---------|----------|
   | `local_holdout` | Local | `rule_a` only | 50% | Everyone |
   | `global_holdout` | Global | All rules | 10% | Everyone |

4. Activate all rules and holdouts.

### 3. Update and run

1. Set `SDK_KEY` in `main.go` to your SDK key
2. Run: `go run examples/local_holdouts/main.go`
3. The code will print the current project state (flags, rules, holdouts
   with their traffic %) and run a basic decide call

### 4. Explore

Open `main.go` and start uncommenting code blocks. Each block is a
self-contained example you can modify:

- **Basic decide** -- see which users get held out
- **Attributes** -- test audience-targeted holdouts
- **Forced decisions** -- override holdouts at flag or rule level
- **DecideAll / DecideForKeys** -- see holdouts across multiple flags
- **EnabledFlagsOnly** -- verify held-out flags get excluded
- **Listeners** -- inspect holdout metadata in decision notifications
- **Distribution check** -- verify holdout traffic percentages
- **Wait for refresh** -- keep SDK alive while you make UI changes

Combine blocks, change user IDs, add your own logic. The SDK client polls
every 30 seconds, so changes you make in the Optimizely UI will be picked
up automatically.

## Scenario ideas

These are things to try during the bug bash. They are NOT prescriptive
scripts -- use them as inspiration for exploration. The code blocks in
`main.go` give you the building blocks.

### UI mutation scenarios
Make changes in the Optimizely UI while the SDK is running (use the "wait
for datafile refresh" code block to keep the SDK alive).

| # | What to try | What could break |
|---|------------|-----------------|
| 1 | **Delete a running holdout** -- note held-out users, delete holdout, re-check | SDK crashes, stale holdout still applied, users don't recover |
| 2 | **Change holdout traffic** -- try 50%->0%, 50%->100%, 10%->90% | Wrong distribution, users not re-bucketed |
| 3 | **Switch local to global** -- local holdout on rule_a, switch to global | flag_b not affected when it should be |
| 4 | **Switch global to local** -- global holdout, switch to local targeting one rule | Other flags still incorrectly held out |
| 5 | **Add audience to holdout** -- holdout with Everyone, add audience condition | Users without attribute still held out |
| 6 | **Remove audience from holdout** -- audience holdout, remove audience | Holdout doesn't expand to all users |
| 7 | **Delete the rule a holdout targets** -- local holdout targets rule_a, delete rule_a | Panic, nil pointer, or holdout leaks to other rules |
| 8 | **Pause a holdout** -- running holdout at 50%, pause it | Holdout still applied after pause |
| 9 | **Add holdout to running experiment** -- experiment with no holdouts, add one | New holdout not picked up |

### Feature interaction edge cases

| # | What to try | Expected |
|---|------------|----------|
| 10 | **Forced decision on held-out user** -- find held-out user, SetForcedDecision | Forced variation wins, holdout bypassed. Remove forced decision: holdout returns. |
| 11 | **Rule-level forced decision** -- force at rule level, not flag level | Overrides holdout for that rule only |
| 12 | **DecideAll with global holdout** -- user hits global holdout | ALL flags show holdout decision |
| 13 | **DecideForKeys vs DecideAll** -- compare results for same user | Identical decisions for requested flags |
| 14 | **EnabledFlagsOnly + holdout** -- held-out flags have enabled=false | Held-out flags excluded from results |
| 15 | **Decision listener metadata** -- register listener, trigger holdout | Listener fires with holdout ID and rule_type="holdout" |
| 16 | **Track after holdout** -- user held out, then TrackEvent | Conversion event fires successfully |
| 17 | **DisableDecisionEvent + holdout** -- decide with option | Same decision, no impression dispatched |

### Stress and boundary scenarios

| # | What to try | What could break |
|---|------------|-----------------|
| 18 | **10+ holdouts on same rule** -- different traffic each | Wrong evaluation order, incorrect cumulative rate |
| 19 | **Large project** (100+ flags, ask Jae Kim) -- add holdouts | Performance degradation, timeouts |
| 20 | **0% traffic holdout** -- should never apply | Holdout incorrectly applied |
| 21 | **Empty includedRules []** -- local holdout targeting nothing | Incorrectly treated as global |
| 22 | **Rapid UI changes** -- 5 changes in 60 seconds | Stale decisions, race conditions |
| 23 | **Concurrent decide calls** -- multiple goroutines | Panics, data races |
| 24 | **Same user, different attributes** -- decide without, then with attributes | Audience holdout activates only with matching attributes |
| 25 | **Long/special user IDs** -- 1000+ chars, unicode, emojis | Bucketing breaks, panics |
| 26 | **Global holdout consistency** -- user held out on flag_a, check flag_b, flag_c | Same user should be held out on all flags |
| 27 | **Holdout on flag with no rules** -- create holdout targeting non-existent rule | Should return default off without crash |

## Transcoding to other SDKs

This Go implementation is the reference. To port to your SDK:

1. Copy `main.go` structure into your language
2. Adapt SDK API calls:
   - `CreateUserContext` with attributes
   - `Decide` / `DecideAll` / `DecideForKeys` with options
   - `SetForcedDecision` / `RemoveForcedDecision`
   - Decision and Track listeners
   - `GetOptimizelyConfig` for project inspection
3. Keep the commented-out blocks pattern -- testers uncomment and modify
4. Keep the scenario ideas list in your README

## Log bugs here

https://episerver99-my.sharepoint.com/:x:/g/personal/matjaz_pirnovar_optimizely_com/IQCkcX_sg-ZeS7uNrKPc6wf8AVKqgPsQJrSjciNNJy035KM?e=doxfH4&nav=MTVfezAwMDAwMDAwLTAwMDEtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMH0
