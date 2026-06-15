/*
============================================================
Local Holdouts Bug Bash -- Go SDK
============================================================

OVERVIEW:
Local holdouts target specific experiment/delivery rules rather than applying
globally to all rules across all flags. This bug bash validates that the SDK
correctly evaluates local holdouts, handles UI changes (datafile updates),
and doesn't break under edge cases.

HOW LOCAL HOLDOUTS WORK:

  Evaluation priority (highest to lowest):
    1. Global holdouts  -- flag-level, before any rule evaluation
    2. Forced decisions  -- per-rule, SetForcedDecision overrides everything below
    3. Local holdouts    -- per-rule, after forced decisions
    4. Normal experiment/rollout bucketing

  Holdout types (determined by includedRules field in datafile):
    - Global:  includedRules = null     --> applies to ALL rules on ALL flags
    - Local:   includedRules = ["5001"] --> applies only to rule 5001
    - Empty:   includedRules = []       --> local holdout targeting nothing (effectively disabled, NOT global)

  When a user is held out:
    - decision.VariationKey = "ho_off_key"
    - decision.Enabled      = false
    - decision.RuleKey      = the holdout key (e.g. "my_local_holdout")
    - Impression event:  rule_type = "holdout", campaign_id = ""

OPTIMIZELY PROJECT SETUP:

  1. Create or reuse a project in your Optimizely environment.
  2. Create a custom audience:
     - Name: "Custom Attr Audience"
     - Condition: custom attribute "customattr" equals "yes"
  3. Create flags with A/B test rules:

     Flag Key   | Rule Key (A/B Test) | Variations | Traffic | Audience
     -----------|---------------------|------------|---------|----------
     flag_a     | rule_a              | on, off    | 100%    | Everyone
     flag_b     | rule_b              | on, off    | 100%    | Everyone

  4. Create holdouts:

     Holdout Name     | Type   | Targeted Rules   | Traffic | Audience
     -----------------|--------|------------------|---------|-------------------
     local_holdout    | Local  | rule_a only      | 50%     | Everyone
     global_holdout   | Global | All rules        | 10%     | Everyone

  5. Activate all rules and holdouts.
  6. Copy your SDK Key from Settings -> Environments.
  7. Update the SDK_KEY constant below.

RUNNING:
  go run main.go                      # runs the basic exploration code
  go run main.go -mode=static         # runs static sanity checks with bundled datafile

============================================================
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/client"
	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/event"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/optimizely/go-sdk/v2/pkg/notification"
)

// ============================================================
// CONFIGURATION -- Update these for your environment
// ============================================================
const (
	SDK_KEY = "YOUR_SDK_KEY_HERE"

	// Flags and rules -- must match your project setup
	FLAG_A = "flag_a"
	FLAG_B = "flag_b"

	// Audience attribute
	ATTR_KEY   = "customattr"
	ATTR_MATCH = "yes"
)

var (
	mode = flag.String("mode", "live", "Mode: 'static' (bundled datafile) or 'live' (actual project)")
)

func main() {
	flag.Parse()

	if *mode == "static" {
		logging.SetLogLevel(logging.LogLevelWarning)
		runStaticSanityCheck()
		return
	}

	// ============================================================
	// LIVE MODE -- Modify the code below to explore holdouts
	// ============================================================
	//
	// This is your sandbox. The SDK client is created with polling enabled
	// (refreshes datafile every 30 seconds). Make changes in the Optimizely
	// UI, wait ~30s, then re-run to see the updated behavior.
	//
	// Uncomment sections below to try different things. Modify freely.

	if SDK_KEY == "YOUR_SDK_KEY_HERE" {
		fmt.Println("ERROR: Set your SDK key in the SDK_KEY constant at the top of main.go")
		return
	}

	logging.SetLogLevel(logging.LogLevelDebug)
	optimizelyClient := createLiveClient(SDK_KEY)
	defer optimizelyClient.Close()

	// Show current project state (flags, rules, holdouts)
	inspectProject(optimizelyClient)

	// ----------------------------------------------------------
	// BASIC: Decide on a flag and see what happens
	// ----------------------------------------------------------
	fmt.Println("\n--- Basic decide ---")
	user := optimizelyClient.CreateUserContext("user_123", nil)
	d := user.Decide(FLAG_A, []decide.OptimizelyDecideOptions{decide.IncludeReasons})
	printDecision("user_123 on "+FLAG_A, d)

	// ----------------------------------------------------------
	// TRY DIFFERENT USERS: Some will be held out, some won't
	// ----------------------------------------------------------
	fmt.Println("\n--- Try multiple users ---")
	for i := 1; i <= 20; i++ {
		uid := fmt.Sprintf("user_%d", i)
		uc := optimizelyClient.CreateUserContext(uid, nil)
		d := uc.Decide(FLAG_A, nil)
		tag := "normal"
		if isHoldout(d) {
			tag = "HOLDOUT:" + d.RuleKey
		}
		fmt.Printf("  %-15s %-30s var=%-10s enabled=%v\n", uid, tag, d.VariationKey, d.Enabled)
	}

	// ----------------------------------------------------------
	// DECIDE WITH ATTRIBUTES: Test audience-targeted holdouts
	// ----------------------------------------------------------
	// fmt.Println("\n--- With audience attributes ---")
	// userWithAttr := optimizelyClient.CreateUserContext("user_123", map[string]interface{}{
	// 	ATTR_KEY: ATTR_MATCH,
	// })
	// d2 := userWithAttr.Decide(FLAG_A, []decide.OptimizelyDecideOptions{decide.IncludeReasons})
	// printDecision("user_123 with customattr=yes", d2)
	//
	// // Same user WITHOUT the attribute -- should NOT hit audience holdout
	// userNoAttr := optimizelyClient.CreateUserContext("user_123", nil)
	// d3 := userNoAttr.Decide(FLAG_A, nil)
	// printDecision("user_123 without attribute", d3)

	// ----------------------------------------------------------
	// FORCED DECISIONS: Override holdout at flag level
	// ----------------------------------------------------------
	// fmt.Println("\n--- Forced decision (flag level) ---")
	// uc := optimizelyClient.CreateUserContext("user_123", nil)
	//
	// // Normal decision first
	// before := uc.Decide(FLAG_A, nil)
	// printDecision("Before forced decision", before)
	//
	// // Force variation to "on" -- should bypass holdout
	// ctx := decision.OptimizelyDecisionContext{FlagKey: FLAG_A}
	// fd := decision.OptimizelyForcedDecision{VariationKey: "on"}
	// uc.SetForcedDecision(ctx, fd)
	//
	// forced := uc.Decide(FLAG_A, nil)
	// printDecision("With forced decision", forced)
	//
	// // Remove forced decision -- holdout should return
	// uc.RemoveForcedDecision(ctx)
	// after := uc.Decide(FLAG_A, nil)
	// printDecision("After removing forced decision", after)

	// ----------------------------------------------------------
	// FORCED DECISIONS: Override at rule level (more specific)
	// ----------------------------------------------------------
	// fmt.Println("\n--- Forced decision (rule level) ---")
	// uc := optimizelyClient.CreateUserContext("user_123", nil)
	// ruleCtx := decision.OptimizelyDecisionContext{FlagKey: FLAG_A, RuleKey: "rule_a"}
	// ruleFD := decision.OptimizelyForcedDecision{VariationKey: "on"}
	// uc.SetForcedDecision(ruleCtx, ruleFD)
	// d := uc.Decide(FLAG_A, nil)
	// printDecision("Rule-level forced decision", d)
	// uc.RemoveAllForcedDecisions()

	// ----------------------------------------------------------
	// DECIDE ALL: See holdouts across all flags at once
	// ----------------------------------------------------------
	// fmt.Println("\n--- DecideAll ---")
	// uc := optimizelyClient.CreateUserContext("user_123", nil)
	// all := uc.DecideAll(nil)
	// for fk, d := range all {
	// 	tag := ""
	// 	if isHoldout(d) {
	// 		tag = " [HOLDOUT]"
	// 	}
	// 	fmt.Printf("  %-15s rule=%-20s var=%-10s enabled=%v%s\n",
	// 		fk, d.RuleKey, d.VariationKey, d.Enabled, tag)
	// }

	// ----------------------------------------------------------
	// DECIDE FOR KEYS: Check subset of flags
	// ----------------------------------------------------------
	// fmt.Println("\n--- DecideForKeys ---")
	// uc := optimizelyClient.CreateUserContext("user_123", nil)
	// subset := uc.DecideForKeys([]string{FLAG_A, FLAG_B}, nil)
	// for fk, d := range subset {
	// 	fmt.Printf("  %s: rule=%s var=%s enabled=%v\n", fk, d.RuleKey, d.VariationKey, d.Enabled)
	// }

	// ----------------------------------------------------------
	// ENABLED FLAGS ONLY: Held-out flags should be excluded
	// ----------------------------------------------------------
	// fmt.Println("\n--- EnabledFlagsOnly (holdout = enabled=false, should be excluded) ---")
	// uc := optimizelyClient.CreateUserContext("user_123", nil)
	// allDecisions := uc.DecideAll(nil)
	// enabledOnly := uc.DecideAll([]decide.OptimizelyDecideOptions{decide.EnabledFlagsOnly})
	// fmt.Printf("  All flags: %d | EnabledFlagsOnly: %d\n", len(allDecisions), len(enabledOnly))
	// for fk, d := range allDecisions {
	// 	_, inEnabled := enabledOnly[fk]
	// 	fmt.Printf("  %-15s enabled=%v  in_enabled_only=%v\n", fk, d.Enabled, inEnabled)
	// }

	// ----------------------------------------------------------
	// DISABLE DECISION EVENT: Holdout decision without impression
	// ----------------------------------------------------------
	// fmt.Println("\n--- DisableDecisionEvent ---")
	// uc := optimizelyClient.CreateUserContext("user_123", nil)
	// d1 := uc.Decide(FLAG_A, nil) // fires impression
	// d2 := uc.Decide(FLAG_A, []decide.OptimizelyDecideOptions{decide.DisableDecisionEvent}) // no impression
	// fmt.Printf("  With event:    rule=%s var=%s\n", d1.RuleKey, d1.VariationKey)
	// fmt.Printf("  Without event: rule=%s var=%s (should be same decision, no impression sent)\n", d2.RuleKey, d2.VariationKey)

	// ----------------------------------------------------------
	// TRACK EVENT: Send conversion for a held-out user
	// ----------------------------------------------------------
	// fmt.Println("\n--- Track after holdout ---")
	// uc := optimizelyClient.CreateUserContext("user_123", nil)
	// d := uc.Decide(FLAG_A, nil)
	// printDecision("Decision before track", d)
	// err := uc.TrackEvent("my_event", map[string]interface{}{"revenue": 100})
	// fmt.Printf("  TrackEvent result: %v\n", err)

	// ----------------------------------------------------------
	// DECISION LISTENER: See holdout metadata in notifications
	// ----------------------------------------------------------
	// fmt.Println("\n--- Decision listener ---")
	// nc := optimizelyClient.GetNotificationCenter()
	// nc.AddHandler(notification.Decision, func(payload interface{}) {
	// 	if dn, ok := payload.(notification.DecisionNotification); ok {
	// 		info, _ := json.Marshal(dn.DecisionInfo)
	// 		fmt.Printf("  LISTENER: user=%s type=%s info=%s\n", dn.UserContext.ID, dn.Type, string(info))
	// 	}
	// })
	// uc := optimizelyClient.CreateUserContext("user_123", nil)
	// uc.Decide(FLAG_A, nil)

	// ----------------------------------------------------------
	// TRACK LISTENER: Verify conversion events for held-out users
	// ----------------------------------------------------------
	// fmt.Println("\n--- Track listener ---")
	// optimizelyClient.OnTrack(func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
	// 	fmt.Printf("  TRACK: event=%s user=%s\n", eventKey, userContext.ID)
	// })
	// uc := optimizelyClient.CreateUserContext("user_123", nil)
	// uc.Decide(FLAG_A, nil)
	// uc.TrackEvent("my_event", nil)

	// ----------------------------------------------------------
	// DISTRIBUTION CHECK: Verify holdout traffic percentage
	// ----------------------------------------------------------
	// fmt.Println("\n--- Distribution (1000 users) ---")
	// logging.SetLogLevel(logging.LogLevelWarning)
	// counts := map[string]int{}
	// for i := 1; i <= 1000; i++ {
	// 	uid := fmt.Sprintf("dist_%d", i)
	// 	uc := optimizelyClient.CreateUserContext(uid, nil)
	// 	d := uc.Decide(FLAG_A, nil)
	// 	key := d.RuleKey + "/" + d.VariationKey
	// 	if isHoldout(d) {
	// 		key = "HOLDOUT:" + d.RuleKey
	// 	}
	// 	counts[key]++
	// }
	// for key, count := range counts {
	// 	fmt.Printf("  %-35s %4d (%.1f%%)\n", key, count, float64(count)/10)
	// }
	// logging.SetLogLevel(logging.LogLevelDebug)

	// ----------------------------------------------------------
	// WAIT FOR DATAFILE REFRESH: Keep SDK alive to test UI changes
	// ----------------------------------------------------------
	// Use this when testing UI mutations. Make a change in the UI,
	// then wait for the SDK to pick it up via polling.
	//
	// fmt.Println("\n--- Waiting for datafile refresh (Ctrl+C to stop) ---")
	// fmt.Println("Make a change in the UI, then watch for decision changes.")
	// uid := "watch_user_42"
	// lastRule := ""
	// for {
	// 	uc := optimizelyClient.CreateUserContext(uid, nil)
	// 	d := uc.Decide(FLAG_A, nil)
	// 	if d.RuleKey != lastRule {
	// 		fmt.Printf("  [%s] CHANGED: rule=%s var=%s enabled=%v\n",
	// 			time.Now().Format("15:04:05"), d.RuleKey, d.VariationKey, d.Enabled)
	// 		lastRule = d.RuleKey
	// 	}
	// 	time.Sleep(10 * time.Second)
	// }

	// Silence unused import warnings -- remove these as you uncomment code above
	_ = decision.OptimizelyDecisionContext{}
	_ = notification.Decision
	_ = entities.UserContext{}
	_ = event.ConversionEvent{}
	_ = json.Marshal
	_ = time.Second
}

// ============================================================
// SCENARIO IDEAS -- Things to try during the bug bash
// ============================================================
//
// These are NOT automated tests. They are ideas for manual exploration.
// Use the code blocks above as building blocks, combine them, modify them.
//
// ---- UI MUTATION SCENARIOS ----
// (Make changes in the Optimizely UI while the SDK is running)
//
// 1. DELETE A RUNNING HOLDOUT
//    - Run the distribution check, note which users are held out
//    - Delete the holdout in the UI, wait for datafile refresh
//    - Re-run: previously held-out users should now get normal decisions
//    - What if you delete a GLOBAL holdout? Do all flags recover?
//
// 2. CHANGE HOLDOUT TRAFFIC
//    - Start with 50%, run distribution check
//    - Change to 0% in UI --> everyone should get normal decisions
//    - Change to 100% --> everyone should be held out
//    - Change to 1% --> only ~1% held out
//    - Does the SDK re-bucket correctly each time?
//
// 3. SWITCH LOCAL <-> GLOBAL
//    - Start with a local holdout targeting rule_a only
//    - Verify flag_b is NOT affected
//    - Switch it to global in the UI
//    - After refresh: flag_b should NOW be affected too
//    - Switch back to local: flag_b should stop being affected
//
// 4. ADD/REMOVE AUDIENCE ON HOLDOUT
//    - Holdout with audience: only users with customattr=yes get held out
//    - Remove the audience: ALL users should now get held out
//    - Add it back: only matching users get held out again
//    - Try changing the attribute value in the audience condition
//
// 5. DELETE THE RULE A HOLDOUT TARGETS
//    - Local holdout targets rule_a
//    - Delete rule_a from the flag in the UI
//    - Does the SDK crash? Panic? Or gracefully ignore the holdout?
//
// 6. PAUSE A HOLDOUT
//    - Running holdout with 50% traffic
//    - Pause it in the UI
//    - After refresh: NO users should be held out
//    - Re-activate: users should be held out again
//
// 7. ADD A HOLDOUT TO A RUNNING EXPERIMENT
//    - Experiment running with no holdouts, users getting normal decisions
//    - Create a new local holdout targeting that experiment
//    - After refresh: some users should now be held out
//
// ---- FEATURE INTERACTION EDGE CASES ----
//
// 8. FORCED DECISION BEATS HOLDOUT
//    - Find a user who IS held out (run distribution check)
//    - SetForcedDecision for that user --> should get forced variation, NOT holdout
//    - RemoveForcedDecision --> holdout should return
//    - Try both flag-level and rule-level forced decisions
//
// 9. DECIDE ALL WITH GLOBAL HOLDOUT
//    - User hits global holdout
//    - DecideAll should show holdout on EVERY flag
//    - DecideForKeys for a subset should match DecideAll for those keys
//
// 10. ENABLED FLAGS ONLY + HOLDOUT
//     - Holdout sets enabled=false
//     - DecideAll with EnabledFlagsOnly should EXCLUDE held-out flags
//     - Verify the excluded flags are exactly the held-out ones
//
// 11. DECISION LISTENER METADATA
//     - Register a decision listener
//     - Make a decide call that hits a holdout
//     - Check: does the listener fire? What's in DecisionInfo?
//     - Expected: experiment_id = holdout_id, rule_type = "holdout"
//
// 12. TRACK AFTER HOLDOUT
//     - User is held out, then TrackEvent is called
//     - Does the conversion event fire? (it should)
//     - Check the event payload for correct metadata
//
// 13. DISABLE DECISION EVENT + HOLDOUT
//     - Decide with DisableDecisionEvent option
//     - Decision result should be the same (holdout still applies)
//     - But no impression event should be dispatched
//
// ---- STRESS & BOUNDARY SCENARIOS ----
//
// 14. MANY HOLDOUTS ON SAME RULE
//     - Create 10+ local holdouts all targeting the same rule
//     - Each with different traffic (5%, 10%, 15%, ...)
//     - Are they evaluated in datafile order? First match wins?
//     - Run distribution check -- does total holdout rate make sense?
//
// 15. LARGE PROJECT WITH HOLDOUTS
//     - Use a project with 100+ flags (ask Jae Kim for a large datafile)
//     - Add local holdouts targeting a few rules
//     - Does decision performance degrade? Any timeouts?
//
// 16. HOLDOUT WITH 0% TRAFFIC
//     - Create holdout with traffic set to 0%
//     - Should NEVER hold out any user
//     - Verify across 1000+ users
//
// 17. EMPTY INCLUDED RULES (includedRules: [])
//     - This is a local holdout that targets NOTHING
//     - Should NOT be treated as global
//     - Should have NO effect on any flag
//
// 18. RAPID DATAFILE CHANGES
//     - Use the "wait for datafile refresh" code block
//     - Make 5 changes in rapid succession in the UI
//     - Does the SDK eventually settle on the correct state?
//     - Any race conditions or stale decisions?
//
// 19. CONCURRENT DECIDE CALLS
//     - Spin up multiple goroutines calling Decide simultaneously
//     - Any panics, data races, or inconsistent results?
//
// 20. HOLDOUT + SAME USER DIFFERENT ATTRIBUTES
//     - Same user ID, first decide with no attributes
//     - Then decide WITH attributes that match a holdout audience
//     - Does the holdout correctly activate only with matching attributes?
//
// 21. VERY LONG USER IDS / SPECIAL CHARACTERS
//     - User ID with 1000+ characters
//     - User ID with unicode, emojis, spaces, null bytes
//     - Does bucketing still work? Any panics?
//
// 22. MULTIPLE FLAGS, ONE GLOBAL HOLDOUT
//     - 5+ flags, one global holdout at 10%
//     - For a given user, if they're held out on flag_a, are they also
//       held out on flag_b, flag_c, etc? (they should be -- same bucketing)
//
// 23. HOLDOUT ON A FLAG WITH NO RULES
//     - Create a flag with no experiment rules
//     - Create a holdout (global or local targeting a non-existent rule)
//     - What does Decide return? Should be default off without crash
//
// ============================================================

// ============================================================
// HELPERS -- Used by the code above, no need to modify
// ============================================================

func isHoldout(d client.OptimizelyDecision) bool {
	return d.VariationKey == "ho_off_key" && !d.Enabled
}

func printDecision(label string, d client.OptimizelyDecision) {
	fmt.Printf("\n  %s:\n", label)
	fmt.Printf("    flag_key:      %s\n", d.FlagKey)
	fmt.Printf("    rule_key:      %s\n", d.RuleKey)
	fmt.Printf("    variation_key: %s\n", d.VariationKey)
	fmt.Printf("    enabled:       %v\n", d.Enabled)
	if d.Variables != nil {
		vars := d.Variables.ToMap()
		if len(vars) > 0 {
			jsonBytes, _ := json.Marshal(vars)
			fmt.Printf("    variables:     %s\n", string(jsonBytes))
		}
	}
	if len(d.Reasons) > 0 {
		fmt.Println("    reasons:")
		for _, r := range d.Reasons {
			fmt.Printf("      - %s\n", r)
		}
	}
}

func inspectProject(c *client.OptimizelyClient) {
	fmt.Println("\n============================================================")
	fmt.Println("  Current Project State")
	fmt.Println("============================================================")

	optConfig := c.GetOptimizelyConfig()
	if optConfig == nil {
		fmt.Println("  Could not get OptimizelyConfig")
		return
	}

	fmt.Printf("\n  Revision: %s\n", optConfig.Revision)
	fmt.Printf("  Flags (%d):\n", len(optConfig.FeaturesMap))
	for fk, feat := range optConfig.FeaturesMap {
		fmt.Printf("    %s:\n", fk)
		for _, rule := range feat.ExperimentRules {
			vars := make([]string, 0)
			for vk := range rule.VariationsMap {
				vars = append(vars, vk)
			}
			fmt.Printf("      experiment: %s (id=%s) variations=%v\n", rule.Key, rule.ID, vars)
		}
		for _, rule := range feat.DeliveryRules {
			fmt.Printf("      delivery:   %s (id=%s)\n", rule.Key, rule.ID)
		}
	}

	// Show holdouts from project config
	projectConfig, err := c.ConfigManager.GetConfig()
	if err != nil {
		return
	}

	fmt.Println("\n  Holdouts:")
	globalHoldouts := projectConfig.GetGlobalHoldouts()
	for _, ho := range globalHoldouts {
		trafficPct := 0.0
		if len(ho.TrafficAllocation) > 0 {
			trafficPct = float64(ho.TrafficAllocation[0].EndOfRange) / 100.0
		}
		audience := "Everyone"
		if len(ho.AudienceIds) > 0 {
			audience = fmt.Sprintf("%v", ho.AudienceIds)
		}
		fmt.Printf("    [GLOBAL] %-25s traffic=%.0f%%  audience=%s\n", ho.Key, trafficPct, audience)
	}

	// Find local holdouts by checking rules
	seen := map[string]bool{}
	for _, feat := range optConfig.FeaturesMap {
		for _, rule := range feat.ExperimentRules {
			for _, ho := range projectConfig.GetHoldoutsForRule(rule.ID) {
				if seen[ho.ID] {
					continue
				}
				seen[ho.ID] = true
				trafficPct := 0.0
				if len(ho.TrafficAllocation) > 0 {
					trafficPct = float64(ho.TrafficAllocation[0].EndOfRange) / 100.0
				}
				audience := "Everyone"
				if len(ho.AudienceIds) > 0 {
					audience = fmt.Sprintf("%v", ho.AudienceIds)
				}
				rules := "none"
				if ho.IncludedRules != nil {
					rules = fmt.Sprintf("%v", *ho.IncludedRules)
				}
				fmt.Printf("    [LOCAL]  %-25s traffic=%.0f%%  audience=%s  rules=%s\n",
					ho.Key, trafficPct, audience, rules)
			}
		}
	}

	if len(globalHoldouts) == 0 && len(seen) == 0 {
		fmt.Println("    (no holdouts found)")
	}
	fmt.Println()
}

// ============================================================
// CLIENT CREATION
// ============================================================

func createLiveClient(key string) *client.OptimizelyClient {
	configManager := config.NewPollingProjectConfigManager(key,
		// Remove or change this line if not using the staging CDN
		config.WithDatafileURLTemplate("https://optimizely-staging.s3.amazonaws.com/datafiles/%s.json"),
		config.WithPollingInterval(30*time.Second),
	)
	factory := &client.OptimizelyFactory{SDKKey: key}
	c, err := factory.Client(
		client.WithConfigManager(configManager),
	)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Waiting for datafile...")
	time.Sleep(3 * time.Second)
	fmt.Println("Client ready.")
	return c
}

func createStaticClient(datafile []byte) *client.OptimizelyClient {
	configManager := config.NewStaticProjectConfigManagerWithOptions("",
		config.WithInitialDatafile(datafile),
	)
	factory := &client.OptimizelyFactory{}
	c, err := factory.Client(
		client.WithConfigManager(configManager),
	)
	if err != nil {
		fmt.Printf("Error creating static client: %v\n", err)
		os.Exit(1)
	}
	return c
}

func loadDatafile() []byte {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	path := filepath.Join(dir, "local_holdouts.json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading datafile: %v\n", err)
		os.Exit(1)
	}
	return data
}

// ============================================================
// STATIC SANITY CHECK
// ============================================================
// Quick automated check using the bundled datafile.
// Run first to verify the SDK basics work before live exploration.
//
// Datafile: local_holdouts.json (from fullstack-sdk-compatibility-suite)
// Flags: flag_a (3 rules), flag_b (2 rules), flag_c (1 rule)
// Holdouts: 7 holdouts of various types (local, global, audience, zero-traffic)

func runStaticSanityCheck() {
	fmt.Println("============================================================")
	fmt.Println("  Static Sanity Check (bundled datafile)")
	fmt.Println("============================================================\n")

	datafile := loadDatafile()
	c := createStaticClient(datafile)
	defer c.Close()

	passed := 0
	failed := 0

	// Test 1: Basic local holdout applies
	fmt.Println("--- Test 1: Basic local holdout ---")
	fmt.Println("ho_local_single_rule (30%, targets rule 5001 on flag_a)")
	holdoutCount := 0
	for i := 1; i <= 30; i++ {
		uid := fmt.Sprintf("static_user_%d", i)
		uc := c.CreateUserContext(uid, nil)
		d := uc.Decide("flag_a", nil)
		if isHoldout(d) {
			holdoutCount++
		}
	}
	if holdoutCount > 0 && holdoutCount < 30 {
		fmt.Printf("  PASS: %d/30 held out (expected mix at 30%%)\n", holdoutCount)
		passed++
	} else {
		fmt.Printf("  FAIL: %d/30 held out (expected mix)\n", holdoutCount)
		failed++
	}

	// Test 2: Global holdout affects all flags
	fmt.Println("\n--- Test 2: Global holdout affects all flags ---")
	fmt.Println("ho_global_all_rules (15%, includedRules=null)")
	flags := []string{"flag_a", "flag_b", "flag_c"}
	globalHits := map[string]int{}
	for i := 1; i <= 50; i++ {
		uid := fmt.Sprintf("global_user_%d", i)
		uc := c.CreateUserContext(uid, nil)
		for _, f := range flags {
			d := uc.Decide(f, nil)
			if d.RuleKey == "ho_global_all_rules" {
				globalHits[f]++
			}
		}
	}
	allHit := true
	for _, f := range flags {
		if globalHits[f] == 0 {
			allHit = false
		}
		fmt.Printf("  %s: %d/50 global holdout hits\n", f, globalHits[f])
	}
	if allHit {
		fmt.Println("  PASS: Global holdout applied to all flags")
		passed++
	} else {
		fmt.Println("  FAIL: Global holdout missing on some flags")
		failed++
	}

	// Test 3: Local holdout does NOT affect non-targeted rules
	fmt.Println("\n--- Test 3: Local holdout scoping ---")
	fmt.Println("ho_local_single_rule targets 5001 (flag_a) -- should NOT affect flag_b")
	wrongHits := 0
	for i := 1; i <= 30; i++ {
		uid := fmt.Sprintf("nontarget_%d", i)
		uc := c.CreateUserContext(uid, nil)
		d := uc.Decide("flag_b", nil)
		if d.RuleKey == "ho_local_single_rule" {
			wrongHits++
		}
	}
	if wrongHits == 0 {
		fmt.Println("  PASS: flag_b unaffected by local holdout")
		passed++
	} else {
		fmt.Printf("  FAIL: %d flag_b decisions hit ho_local_single_rule\n", wrongHits)
		failed++
	}

	// Test 4: Zero-traffic holdout never applied
	fmt.Println("\n--- Test 4: Zero-traffic holdout ---")
	fmt.Println("ho_local_zero_traffic (0%, targets 5005) should NEVER apply")
	zeroHits := 0
	for i := 1; i <= 50; i++ {
		uid := fmt.Sprintf("zero_%d", i)
		uc := c.CreateUserContext(uid, nil)
		d := uc.Decide("flag_b", nil)
		if d.RuleKey == "ho_local_zero_traffic" {
			zeroHits++
		}
	}
	if zeroHits == 0 {
		fmt.Println("  PASS: Zero-traffic holdout never applied")
		passed++
	} else {
		fmt.Printf("  FAIL: %d hits from zero-traffic holdout\n", zeroHits)
		failed++
	}

	// Test 5: Global holdout beats local holdout
	fmt.Println("\n--- Test 5: Global beats local (precedence) ---")
	fmt.Println("Rule 5001 targeted by global (15%) AND local (30%)")
	globalCount := 0
	for i := 1; i <= 100; i++ {
		uid := fmt.Sprintf("prec_user_%d", i)
		uc := c.CreateUserContext(uid, nil)
		d := uc.Decide("flag_a", nil)
		if d.RuleKey == "ho_global_all_rules" {
			globalCount++
		}
	}
	if globalCount > 0 {
		fmt.Printf("  PASS: %d/100 hit global holdout (evaluated first)\n", globalCount)
		passed++
	} else {
		fmt.Println("  FAIL: No global holdout hits")
		failed++
	}

	// Test 6: Cross-flag local holdout
	fmt.Println("\n--- Test 6: Cross-flag holdout ---")
	fmt.Println("ho_local_cross_flag (20%, targets 5001 on flag_a + 5004 on flag_b)")
	crossHits := 0
	for i := 1; i <= 50; i++ {
		uid := fmt.Sprintf("cross_user_%d", i)
		uc := c.CreateUserContext(uid, nil)
		dA := uc.Decide("flag_a", nil)
		dB := uc.Decide("flag_b", nil)
		if dA.RuleKey == "ho_local_cross_flag" && dB.RuleKey == "ho_local_cross_flag" {
			crossHits++
		}
	}
	if crossHits > 0 {
		fmt.Printf("  PASS: %d/50 users held out on BOTH flags\n", crossHits)
		passed++
	} else {
		fmt.Println("  FAIL: No cross-flag holdout hits (may be expected at 20%% with competing holdouts)")
		failed++
	}

	// Test 7: Audience holdout respects conditions
	fmt.Println("\n--- Test 7: Audience holdout ---")
	fmt.Println("ho_local_with_audience (40%, targets 5002, audience=customattr:yes)")
	noAttrHits := 0
	for i := 1; i <= 30; i++ {
		uid := fmt.Sprintf("aud_no_%d", i)
		uc := c.CreateUserContext(uid, nil)
		d := uc.Decide("flag_a", nil)
		if d.RuleKey == "ho_local_with_audience" {
			noAttrHits++
		}
	}
	if noAttrHits == 0 {
		fmt.Println("  PASS: Users without attribute NOT held out by audience holdout")
		passed++
	} else {
		fmt.Printf("  FAIL: %d users hit audience holdout without having the attribute\n", noAttrHits)
		failed++
	}

	fmt.Printf("\n============================================================\n")
	fmt.Printf("  Results: %d passed, %d failed\n", passed, failed)
	fmt.Printf("============================================================\n")
}
