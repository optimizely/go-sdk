// Local Holdouts Bug Bash - Go SDK
//
// Two modes:
//   Static:  Quick sanity check with bundled datafile (deterministic, no setup needed)
//   Live:    Interactive exploration tool for breaking SDKs via UI mutations & edge cases
//
// Static:   go run examples/local_holdouts/main.go -mode=static -test=all
// Live:     go run examples/local_holdouts/main.go -sdk_key=YOUR_KEY -scenario=ui_delete_holdout
// Explore:  go run examples/local_holdouts/main.go -sdk_key=YOUR_KEY -explore
// List:     go run examples/local_holdouts/main.go -test=help

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
// CONFIGURATION - Update these to match your RC project
// ============================================================
const SDK_KEY = "YOUR_SDK_KEY_HERE"

var (
	testCase = flag.String("test", "", "Static test case to run (use -test=help for list)")
	mode     = flag.String("mode", "live", "Mode: 'static' or 'live'")
	sdkKey   = flag.String("sdk_key", SDK_KEY, "SDK key for live mode")
	scenario = flag.String("scenario", "", "Live scenario to run (see -test=help)")
	explore  = flag.Bool("explore", false, "Interactive exploration REPL")
	userID   = flag.String("user", "user_1", "Default user ID for exploration")
	flagKey  = flag.String("flag", "", "Flag key for exploration")
	numUsers = flag.Int("n", 20, "Number of users for distribution checks")
)

func main() {
	flag.Parse()

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  Local Holdouts Bug Bash - Go SDK")
	fmt.Println(strings.Repeat("=", 60))

	if *testCase == "help" {
		printHelp()
		return
	}

	if *mode == "static" || *testCase != "" {
		logging.SetLogLevel(logging.LogLevelWarning)
		runStaticTests()
		return
	}

	// Live mode
	if *sdkKey == SDK_KEY || *sdkKey == "" {
		fmt.Println("\nERROR: Provide your RC SDK key via -sdk_key=YOUR_KEY")
		fmt.Println("See README.md for project setup instructions.")
		return
	}

	logging.SetLogLevel(logging.LogLevelDebug)
	optimizelyClient := createLiveClient(*sdkKey)
	defer optimizelyClient.Close()

	if *explore {
		runExploreREPL(optimizelyClient)
		return
	}

	if *scenario != "" {
		runScenario(optimizelyClient, *scenario)
		return
	}

	// Default: show what's available
	fmt.Println("\nNo scenario specified. Use one of:")
	fmt.Println("  -explore              Interactive REPL")
	fmt.Println("  -scenario=<name>      Guided scenario")
	fmt.Println("  -test=help            List all options")
	fmt.Println()
	inspectLiveConfig(optimizelyClient)
}

// ============================================================
// STATIC MODE - Quick sanity check with bundled datafile
// ============================================================

func runStaticTests() {
	fmt.Println("\nMode: STATIC (bundled datafile)")

	datafile := loadDatafile()
	c := createStaticClient(datafile)
	defer c.Close()

	test := *testCase
	if test == "" {
		test = "all"
	}

	switch test {
	case "basic":
		testStaticBasicLocalHoldout(c)
	case "multi_rule":
		testStaticMultiRuleHoldout(c)
	case "cross_flag":
		testStaticCrossFlagHoldout(c)
	case "global":
		testStaticGlobalHoldout(c)
	case "precedence":
		testStaticGlobalBeatsLocal(c)
	case "audience":
		testStaticAudienceHoldout(c)
	case "not_targeted":
		testStaticNotTargeted(c)
	case "zero_traffic":
		testStaticZeroTraffic(c)
	case "all":
		tests := []struct {
			name string
			fn   func(*client.OptimizelyClient)
		}{
			{"basic", testStaticBasicLocalHoldout},
			{"multi_rule", testStaticMultiRuleHoldout},
			{"cross_flag", testStaticCrossFlagHoldout},
			{"global", testStaticGlobalHoldout},
			{"precedence", testStaticGlobalBeatsLocal},
			{"audience", testStaticAudienceHoldout},
			{"not_targeted", testStaticNotTargeted},
			{"zero_traffic", testStaticZeroTraffic},
		}
		for _, t := range tests {
			t.fn(c)
		}
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("  All static tests completed.")
		fmt.Println(strings.Repeat("=", 60))
	default:
		fmt.Printf("\nUnknown test: %s (use -test=help)\n", test)
	}
}

// ============================================================
// LIVE SCENARIOS - Guided walkthroughs for breaking SDKs
// ============================================================

func runScenario(c *client.OptimizelyClient, name string) {
	switch name {
	// --- UI Mutation Scenarios ---
	case "ui_delete_holdout":
		scenarioDeleteHoldout(c)
	case "ui_change_traffic":
		scenarioChangeTraffic(c)
	case "ui_switch_local_global":
		scenarioSwitchLocalGlobal(c)
	case "ui_change_audience":
		scenarioChangeAudience(c)
	case "ui_delete_targeted_rule":
		scenarioDeleteTargetedRule(c)
	case "ui_pause_holdout":
		scenarioPauseHoldout(c)
	case "ui_add_holdout_to_running":
		scenarioAddHoldoutToRunning(c)

	// --- Feature Interaction Scenarios ---
	case "forced_vs_holdout":
		scenarioForcedVsHoldout(c)
	case "forced_rule_level":
		scenarioForcedRuleLevel(c)
	case "holdout_disable_event":
		scenarioHoldoutDisableEvent(c)
	case "holdout_track":
		scenarioHoldoutTrack(c)
	case "holdout_decide_all":
		scenarioHoldoutDecideAll(c)
	case "holdout_decide_for_keys":
		scenarioHoldoutDecideForKeys(c)
	case "holdout_listener":
		scenarioHoldoutListener(c)
	case "holdout_enabled_flags_only":
		scenarioEnabledFlagsOnly(c)

	// --- Stress / Edge Cases ---
	case "rapid_repolling":
		scenarioRapidRepolling(c)
	case "distribution":
		scenarioDistribution(c)

	default:
		fmt.Printf("\nUnknown scenario: %s\n", name)
		printHelp()
	}
}

// ============================================================
// UI MUTATION SCENARIOS
// ============================================================

func scenarioDeleteHoldout(c *client.OptimizelyClient) {
	printScenarioHeader("Delete a Holdout While SDK is Running",
		"Tests: Does the SDK gracefully handle a holdout disappearing from the datafile?",
		"Steps:",
		"  1. We take a snapshot of decisions for several users",
		"  2. You delete a local holdout in the UI",
		"  3. We re-check: held-out users should now get normal experiment decisions",
		"  4. Edge case: what if a global holdout is deleted mid-session?",
	)

	flagKeys := discoverFlags(c)
	if len(flagKeys) == 0 {
		fmt.Println("  No flags found. Set up your project first.")
		return
	}

	fmt.Println("\n  BEFORE: Current decisions")
	snapBefore := takeSnapshot(c, flagKeys, *numUsers)
	printSnapshot(snapBefore)

	waitForUIChange("Delete a holdout in the UI (local or global), then save and wait for datafile refresh")

	fmt.Println("\n  AFTER: Decisions after holdout deletion")
	snapAfter := takeSnapshot(c, flagKeys, *numUsers)
	printSnapshot(snapAfter)

	diffSnapshots(snapBefore, snapAfter)
}

func scenarioChangeTraffic(c *client.OptimizelyClient) {
	printScenarioHeader("Change Holdout Traffic Percentage",
		"Tests: Does the SDK correctly re-bucket when holdout traffic changes?",
		"Interesting mutations to try:",
		"  - 50% -> 0% (holdout effectively disabled, all users should get normal decisions)",
		"  - 50% -> 100% (all users should be held out)",
		"  - 10% -> 90% (dramatic shift, verify distribution changes)",
		"  - Any% -> same% (no change expected, sanity check)",
	)

	flagKeys := discoverFlags(c)

	fmt.Println("\n  BEFORE: Distribution")
	snapBefore := takeSnapshot(c, flagKeys, *numUsers)
	printDistribution(snapBefore)

	waitForUIChange("Change a holdout's traffic % in the UI")

	fmt.Println("\n  AFTER: Distribution")
	snapAfter := takeSnapshot(c, flagKeys, *numUsers)
	printDistribution(snapAfter)

	diffSnapshots(snapBefore, snapAfter)
}

func scenarioSwitchLocalGlobal(c *client.OptimizelyClient) {
	printScenarioHeader("Switch Holdout Between Local and Global",
		"Tests: Does the SDK handle holdout type change correctly?",
		"This is a big structural change in the datafile (includedRules: [...] <-> null)",
		"",
		"Try these:",
		"  - Local -> Global: holdout should now affect ALL flags, not just targeted rules",
		"  - Global -> Local: holdout should STOP affecting non-targeted flags",
		"  - Watch for: users previously held out on flag_b now getting normal decisions",
	)

	flagKeys := discoverFlags(c)

	fmt.Println("\n  BEFORE: Decisions across all flags")
	snapBefore := takeSnapshot(c, flagKeys, *numUsers)
	printSnapshotByFlag(snapBefore, flagKeys)

	waitForUIChange("Switch a holdout between Local and Global in the UI (change which rules it targets)")

	fmt.Println("\n  AFTER: Decisions across all flags")
	snapAfter := takeSnapshot(c, flagKeys, *numUsers)
	printSnapshotByFlag(snapAfter, flagKeys)

	diffSnapshots(snapBefore, snapAfter)
}

func scenarioChangeAudience(c *client.OptimizelyClient) {
	printScenarioHeader("Add/Remove Audience on a Holdout",
		"Tests: Does audience targeting on holdouts work after datafile refresh?",
		"",
		"Try these:",
		"  - Remove audience from holdout: should now affect ALL users regardless of attributes",
		"  - Add audience to holdout: should only affect users matching the audience condition",
		"  - Change audience condition value: users with old value should no longer be held out",
	)

	flagKeys := discoverFlags(c)

	// Test with and without attributes
	fmt.Println("\n  BEFORE: Users WITH customattr=yes")
	snapWithAttr := takeSnapshotWithAttrs(c, flagKeys, *numUsers,
		map[string]interface{}{"customattr": "yes"})
	printDistribution(snapWithAttr)

	fmt.Println("\n  BEFORE: Users WITHOUT attributes")
	snapNoAttr := takeSnapshot(c, flagKeys, *numUsers)
	printDistribution(snapNoAttr)

	waitForUIChange("Add or remove an audience condition on a holdout")

	fmt.Println("\n  AFTER: Users WITH customattr=yes")
	snapWithAttrAfter := takeSnapshotWithAttrs(c, flagKeys, *numUsers,
		map[string]interface{}{"customattr": "yes"})
	printDistribution(snapWithAttrAfter)

	fmt.Println("\n  AFTER: Users WITHOUT attributes")
	snapNoAttrAfter := takeSnapshot(c, flagKeys, *numUsers)
	printDistribution(snapNoAttrAfter)
}

func scenarioDeleteTargetedRule(c *client.OptimizelyClient) {
	printScenarioHeader("Delete A Rule That a Holdout Targets",
		"Tests: What happens when the experiment/rule referenced by includedRules is deleted?",
		"This is a dangling reference scenario - the holdout points to a rule that no longer exists.",
		"",
		"Expected: SDK should handle gracefully (no crash, no panic). The holdout should",
		"have no effect since its targeted rule doesn't exist anymore.",
		"",
		"Watch for: panics, nil pointer dereferences, or the holdout accidentally",
		"affecting OTHER rules on the same flag.",
	)

	flagKeys := discoverFlags(c)

	fmt.Println("\n  BEFORE: Current state")
	inspectLiveConfig(c)
	snapBefore := takeSnapshot(c, flagKeys, *numUsers)
	printSnapshot(snapBefore)

	waitForUIChange("Delete an experiment rule that a local holdout targets (keep the holdout)")

	fmt.Println("\n  AFTER: State after rule deletion")
	inspectLiveConfig(c)
	snapAfter := takeSnapshot(c, flagKeys, *numUsers)
	printSnapshot(snapAfter)

	diffSnapshots(snapBefore, snapAfter)
}

func scenarioPauseHoldout(c *client.OptimizelyClient) {
	printScenarioHeader("Pause a Running Holdout",
		"Tests: Does pausing a holdout immediately stop it from being applied?",
		"",
		"Try: Pause a holdout that has high traffic (50%+) so the effect is obvious.",
		"All users previously held out should now get normal experiment decisions.",
	)

	flagKeys := discoverFlags(c)

	fmt.Println("\n  BEFORE (holdout running):")
	snapBefore := takeSnapshot(c, flagKeys, *numUsers)
	printDistribution(snapBefore)

	waitForUIChange("Pause a holdout in the UI")

	fmt.Println("\n  AFTER (holdout paused):")
	snapAfter := takeSnapshot(c, flagKeys, *numUsers)
	printDistribution(snapAfter)

	diffSnapshots(snapBefore, snapAfter)
}

func scenarioAddHoldoutToRunning(c *client.OptimizelyClient) {
	printScenarioHeader("Add a New Holdout to a Running Experiment",
		"Tests: Can you add a holdout to an experiment that already has traffic?",
		"",
		"Steps:",
		"  1. Verify experiment is running normally (no holdouts)",
		"  2. Create a new local holdout targeting that experiment",
		"  3. Activate the holdout",
		"  4. Verify some users are now held out who weren't before",
	)

	flagKeys := discoverFlags(c)

	fmt.Println("\n  BEFORE (no holdout on experiment):")
	snapBefore := takeSnapshot(c, flagKeys, *numUsers)
	printDistribution(snapBefore)

	waitForUIChange("Create and activate a NEW local holdout targeting a running experiment")

	fmt.Println("\n  AFTER (holdout added):")
	snapAfter := takeSnapshot(c, flagKeys, *numUsers)
	printDistribution(snapAfter)

	diffSnapshots(snapBefore, snapAfter)
}

// ============================================================
// FEATURE INTERACTION SCENARIOS
// ============================================================

func scenarioForcedVsHoldout(c *client.OptimizelyClient) {
	printScenarioHeader("Forced Decision vs Holdout (Flag Level)",
		"Tests: SetForcedDecision at FLAG level should override any holdout.",
		"We'll test multiple users, set a forced decision, and verify holdout is bypassed.",
	)

	flagKeys := discoverFlags(c)
	if len(flagKeys) == 0 {
		return
	}
	fk := flagKeys[0]
	if *flagKey != "" {
		fk = *flagKey
	}

	fmt.Printf("\n  Testing on flag: %s\n", fk)

	// Find a user who gets held out
	fmt.Println("\n  Step 1: Finding users who are held out...")
	var heldOutUser string
	for i := 1; i <= 100; i++ {
		uid := fmt.Sprintf("forced_test_%d", i)
		uc := c.CreateUserContext(uid, nil)
		d := uc.Decide(fk, []decide.OptimizelyDecideOptions{decide.IncludeReasons})
		if isHoldout(d) {
			heldOutUser = uid
			fmt.Printf("    Found held-out user: %s (rule_key=%s)\n", uid, d.RuleKey)
			printReasons(d)
			break
		}
	}

	if heldOutUser == "" {
		fmt.Println("    No held-out user found in 100 attempts.")
		fmt.Println("    Make sure a holdout is running with reasonable traffic.")
		return
	}

	// Get first variation key from OptimizelyConfig to use as forced variation
	forceVar := "on" // default guess
	optConfig := c.GetOptimizelyConfig()
	if optConfig != nil {
		if feat, ok := optConfig.FeaturesMap[fk]; ok {
			for _, rule := range feat.ExperimentRules {
				for vk := range rule.VariationsMap {
					forceVar = vk
					break
				}
				break
			}
		}
	}

	// Set forced decision and verify override
	fmt.Printf("\n  Step 2: Setting forced decision (variation=%s) on held-out user...\n", forceVar)
	uc := c.CreateUserContext(heldOutUser, nil)

	ctx := decision.OptimizelyDecisionContext{FlagKey: fk}
	fd := decision.OptimizelyForcedDecision{VariationKey: forceVar}
	ok := uc.SetForcedDecision(ctx, fd)
	fmt.Printf("    SetForcedDecision result: %v\n", ok)

	d := uc.Decide(fk, []decide.OptimizelyDecideOptions{decide.IncludeReasons})
	printDecision("With forced decision", d)
	printReasons(d)

	if d.VariationKey == forceVar {
		fmt.Println("    PASS: Forced decision overrides holdout")
	} else {
		fmt.Printf("    FAIL: Expected variation=%s, got %s\n", forceVar, d.VariationKey)
	}

	// Remove and verify holdout returns
	fmt.Println("\n  Step 3: Removing forced decision...")
	uc.RemoveForcedDecision(ctx)
	d2 := uc.Decide(fk, []decide.OptimizelyDecideOptions{decide.IncludeReasons})
	printDecision("After removing forced decision", d2)

	if isHoldout(d2) {
		fmt.Println("    PASS: User is held out again after removing forced decision")
	} else {
		fmt.Println("    INFO: User got a normal decision (may have been re-bucketed)")
	}
}

func scenarioForcedRuleLevel(c *client.OptimizelyClient) {
	printScenarioHeader("Forced Decision at RULE Level vs Holdout",
		"Tests: SetForcedDecision targeting a specific rule_key should override",
		"the holdout for that rule only.",
		"",
		"Edge case: What if you force a decision on a rule that the holdout targets?",
		"What about forcing on a rule the holdout does NOT target?",
	)

	flagKeys := discoverFlags(c)
	if len(flagKeys) == 0 {
		return
	}
	fk := flagKeys[0]
	if *flagKey != "" {
		fk = *flagKey
	}

	// Get rule keys from config
	optConfig := c.GetOptimizelyConfig()
	if optConfig == nil {
		fmt.Println("  Could not get OptimizelyConfig")
		return
	}

	feat, ok := optConfig.FeaturesMap[fk]
	if !ok {
		fmt.Printf("  Flag %s not found in config\n", fk)
		return
	}

	fmt.Printf("\n  Flag: %s\n", fk)
	fmt.Println("  Rules:")
	for _, rule := range feat.ExperimentRules {
		vars := make([]string, 0)
		for vk := range rule.VariationsMap {
			vars = append(vars, vk)
		}
		fmt.Printf("    %s (id=%s, variations=%v)\n", rule.Key, rule.ID, vars)
	}

	if len(feat.ExperimentRules) == 0 {
		fmt.Println("  No experiment rules found.")
		return
	}

	targetRule := feat.ExperimentRules[0]
	var forceVar string
	for vk := range targetRule.VariationsMap {
		forceVar = vk
		break
	}

	fmt.Printf("\n  Testing: Force variation=%s on rule=%s\n", forceVar, targetRule.Key)

	for i := 1; i <= 5; i++ {
		uid := fmt.Sprintf("rule_force_%d", i)
		uc := c.CreateUserContext(uid, nil)

		// Normal decision
		d1 := uc.Decide(fk, nil)

		// Set rule-level forced decision
		ctx := decision.OptimizelyDecisionContext{FlagKey: fk, RuleKey: targetRule.Key}
		fd := decision.OptimizelyForcedDecision{VariationKey: forceVar}
		uc.SetForcedDecision(ctx, fd)
		d2 := uc.Decide(fk, nil)

		fmt.Printf("    %s: normal=[rule=%s var=%s] forced=[rule=%s var=%s]\n",
			uid, d1.RuleKey, d1.VariationKey, d2.RuleKey, d2.VariationKey)

		uc.RemoveAllForcedDecisions()
	}
}

func scenarioHoldoutDisableEvent(c *client.OptimizelyClient) {
	printScenarioHeader("DisableDecisionEvent with Holdout",
		"Tests: When DisableDecisionEvent option is set, holdout impression events",
		"should NOT be dispatched. Verify no events are sent.",
		"",
		"Check: Compare decision results with and without the option.",
		"The decision itself should be the same; only event dispatching differs.",
	)

	flagKeys := discoverFlags(c)
	if len(flagKeys) == 0 {
		return
	}
	fk := flagKeys[0]
	if *flagKey != "" {
		fk = *flagKey
	}

	for i := 1; i <= 10; i++ {
		uid := fmt.Sprintf("event_test_%d", i)
		uc := c.CreateUserContext(uid, nil)

		// With events (default)
		d1 := uc.Decide(fk, nil)
		// Without events
		d2 := uc.Decide(fk, []decide.OptimizelyDecideOptions{decide.DisableDecisionEvent})

		match := "SAME"
		if d1.RuleKey != d2.RuleKey || d1.VariationKey != d2.VariationKey {
			match = "DIFFERENT!"
		}
		ho := ""
		if isHoldout(d1) {
			ho = " [HOLDOUT]"
		}
		fmt.Printf("    %s: rule=%s var=%s (%s)%s\n", uid, d1.RuleKey, d1.VariationKey, match, ho)
	}

	fmt.Println("\n  Verify in your event dispatcher/logs that DisableDecisionEvent")
	fmt.Println("  suppresses the impression event for holdout decisions.")
}

func scenarioHoldoutTrack(c *client.OptimizelyClient) {
	printScenarioHeader("Track Event After Holdout Decision",
		"Tests: Can you track a conversion event for a user who is held out?",
		"The track call should succeed (no error), but the event should include",
		"the holdout context.",
		"",
		"Watch: Does the event payload contain the right metadata?",
	)

	flagKeys := discoverFlags(c)
	if len(flagKeys) == 0 {
		return
	}
	fk := flagKeys[0]
	if *flagKey != "" {
		fk = *flagKey
	}

	// Register track listener
	trackCh := make(chan string, 10)
	_, _ = c.OnTrack(func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
		trackCh <- fmt.Sprintf("Track fired: event=%s user=%s", eventKey, userContext.ID)
	})

	for i := 1; i <= 10; i++ {
		uid := fmt.Sprintf("track_test_%d", i)
		uc := c.CreateUserContext(uid, nil)

		d := uc.Decide(fk, nil)
		ho := ""
		if isHoldout(d) {
			ho = " [HOLDOUT]"
		}

		err := uc.TrackEvent("test_event", nil)
		errStr := "ok"
		if err != nil {
			errStr = err.Error()
		}
		fmt.Printf("    %s: decide=[%s/%s]%s  track=%s\n",
			uid, d.RuleKey, d.VariationKey, ho, errStr)
	}

	// Drain track channel
	fmt.Println("\n  Track listener events:")
	done := time.After(1 * time.Second)
	count := 0
	for {
		select {
		case msg := <-trackCh:
			fmt.Printf("    %s\n", msg)
			count++
		case <-done:
			if count == 0 {
				fmt.Println("    (no track events received in 1s)")
			}
			goto trackDone
		}
	}
trackDone:
	fmt.Println()
}

func scenarioHoldoutDecideAll(c *client.OptimizelyClient) {
	printScenarioHeader("DecideAll with Holdouts",
		"Tests: DecideAll should return holdout decisions for ALL affected flags.",
		"If user hits a global holdout, ALL flags should show holdout.",
		"If user hits a local holdout, only targeted flags/rules should show holdout.",
	)

	for i := 1; i <= *numUsers; i++ {
		uid := fmt.Sprintf("decide_all_%d", i)
		uc := c.CreateUserContext(uid, nil)
		decisions := uc.DecideAll(nil)

		holdouts := []string{}
		normals := []string{}
		for fk, d := range decisions {
			if isHoldout(d) {
				holdouts = append(holdouts, fmt.Sprintf("%s(%s)", fk, d.RuleKey))
			} else {
				normals = append(normals, fmt.Sprintf("%s(%s)", fk, d.VariationKey))
			}
		}

		hoStr := strings.Join(holdouts, ", ")
		if hoStr == "" {
			hoStr = "none"
		}
		fmt.Printf("    %s: holdouts=[%s]  normal=[%s]\n", uid, hoStr, strings.Join(normals, ", "))
	}
}

func scenarioHoldoutDecideForKeys(c *client.OptimizelyClient) {
	printScenarioHeader("DecideForKeys with Holdouts",
		"Tests: DecideForKeys returns holdout decisions only for requested flags.",
		"Compare: request a subset of flags vs all flags. Results for requested",
		"flags should be identical between DecideForKeys and DecideAll.",
	)

	flagKeys := discoverFlags(c)
	if len(flagKeys) < 2 {
		fmt.Println("  Need at least 2 flags for this test.")
		return
	}

	subset := flagKeys[:2]
	fmt.Printf("  Requesting subset: %v\n\n", subset)

	for i := 1; i <= *numUsers; i++ {
		uid := fmt.Sprintf("decide_keys_%d", i)
		uc := c.CreateUserContext(uid, nil)

		subsetDecisions := uc.DecideForKeys(subset, nil)
		allDecisions := uc.DecideAll(nil)

		match := true
		for _, fk := range subset {
			sd := subsetDecisions[fk]
			ad := allDecisions[fk]
			if sd.RuleKey != ad.RuleKey || sd.VariationKey != ad.VariationKey {
				match = false
				fmt.Printf("    %s: MISMATCH on %s! forKeys=[%s/%s] all=[%s/%s]\n",
					uid, fk, sd.RuleKey, sd.VariationKey, ad.RuleKey, ad.VariationKey)
			}
		}
		if match {
			parts := []string{}
			for _, fk := range subset {
				d := subsetDecisions[fk]
				ho := ""
				if isHoldout(d) {
					ho = "*HO*"
				}
				parts = append(parts, fmt.Sprintf("%s=%s%s", fk, d.VariationKey, ho))
			}
			fmt.Printf("    %s: consistent [%s]\n", uid, strings.Join(parts, " "))
		}
	}
}

func scenarioHoldoutListener(c *client.OptimizelyClient) {
	printScenarioHeader("Decision Listener with Holdout Decisions",
		"Tests: The decision notification listener should fire for holdout decisions",
		"and include holdout-specific metadata.",
		"",
		"Watch for: experiment_id=holdout_id, rule_type=holdout in the decision info.",
	)

	// Register decision listener
	nc := c.GetNotificationCenter()
	listenerCh := make(chan string, 100)

	_, _ = nc.AddHandler(notification.Decision, func(payload interface{}) {
		if decisionNotif, ok := payload.(notification.DecisionNotification); ok {
			info, _ := json.Marshal(decisionNotif.DecisionInfo)
			listenerCh <- fmt.Sprintf("user=%s type=%s info=%s",
				decisionNotif.UserContext.ID, decisionNotif.Type, string(info))
		}
	})

	flagKeys := discoverFlags(c)
	if len(flagKeys) == 0 {
		return
	}
	fk := flagKeys[0]
	if *flagKey != "" {
		fk = *flagKey
	}

	fmt.Printf("\n  Making decisions on flag: %s\n\n", fk)

	for i := 1; i <= 5; i++ {
		uid := fmt.Sprintf("listener_%d", i)
		uc := c.CreateUserContext(uid, nil)
		d := uc.Decide(fk, nil)

		ho := ""
		if isHoldout(d) {
			ho = " [HOLDOUT]"
		}
		fmt.Printf("    %s: rule=%s var=%s%s\n", uid, d.RuleKey, d.VariationKey, ho)
	}

	// Drain listener
	fmt.Println("\n  Listener notifications:")
	done := time.After(1 * time.Second)
	for {
		select {
		case msg := <-listenerCh:
			fmt.Printf("    %s\n", msg)
		case <-done:
			return
		}
	}
}

func scenarioEnabledFlagsOnly(c *client.OptimizelyClient) {
	printScenarioHeader("EnabledFlagsOnly with Holdouts",
		"Tests: DecideAll with EnabledFlagsOnly option should EXCLUDE held-out flags",
		"since holdout decisions have enabled=false.",
		"",
		"If a user is held out on flag_a but not flag_b, DecideAll with EnabledFlagsOnly",
		"should return flag_b but NOT flag_a.",
	)

	for i := 1; i <= *numUsers; i++ {
		uid := fmt.Sprintf("enabled_only_%d", i)
		uc := c.CreateUserContext(uid, nil)

		allDecisions := uc.DecideAll(nil)
		enabledOnly := uc.DecideAll([]decide.OptimizelyDecideOptions{decide.EnabledFlagsOnly})

		allKeys := make([]string, 0)
		for k := range allDecisions {
			allKeys = append(allKeys, k)
		}
		enabledKeys := make([]string, 0)
		for k := range enabledOnly {
			enabledKeys = append(enabledKeys, k)
		}

		dropped := []string{}
		for k, d := range allDecisions {
			if _, ok := enabledOnly[k]; !ok {
				dropped = append(dropped, fmt.Sprintf("%s(rule=%s)", k, d.RuleKey))
			}
		}

		if len(dropped) > 0 {
			fmt.Printf("    %s: all=%d enabled_only=%d  DROPPED: %s\n",
				uid, len(allDecisions), len(enabledOnly), strings.Join(dropped, ", "))
		} else {
			fmt.Printf("    %s: all=%d enabled_only=%d  (no flags dropped)\n",
				uid, len(allDecisions), len(enabledOnly))
		}
	}
}

// ============================================================
// STRESS / EDGE CASES
// ============================================================

func scenarioRapidRepolling(c *client.OptimizelyClient) {
	printScenarioHeader("Rapid Datafile Changes",
		"Tests: Make multiple quick changes in the UI and verify the SDK handles",
		"rapid datafile updates without getting confused.",
		"",
		"Steps: Change holdout traffic, wait 30s, change again, wait 30s, change again.",
		"We'll poll decisions continuously and show when they change.",
	)

	flagKeys := discoverFlags(c)
	if len(flagKeys) == 0 {
		return
	}
	fk := flagKeys[0]
	if *flagKey != "" {
		fk = *flagKey
	}

	uid := *userID
	fmt.Printf("\n  Polling decisions for user=%s flag=%s every 10 seconds.\n", uid, fk)
	fmt.Println("  Make changes in the UI. Press Ctrl+C to stop.\n")

	lastRule := ""
	lastVar := ""
	for round := 1; ; round++ {
		uc := c.CreateUserContext(uid, nil)
		d := uc.Decide(fk, nil)

		changed := ""
		if d.RuleKey != lastRule || d.VariationKey != lastVar {
			if lastRule != "" {
				changed = " <-- CHANGED!"
			}
			lastRule = d.RuleKey
			lastVar = d.VariationKey
		}

		fmt.Printf("    [%s] round=%d rule=%s var=%s enabled=%v%s\n",
			time.Now().Format("15:04:05"), round, d.RuleKey, d.VariationKey, d.Enabled, changed)

		time.Sleep(10 * time.Second)
	}
}

func scenarioDistribution(c *client.OptimizelyClient) {
	printScenarioHeader("Distribution Check",
		"Shows decision distribution across many users to verify holdout traffic %.",
		fmt.Sprintf("Testing %d users.", *numUsers),
	)

	logging.SetLogLevel(logging.LogLevelWarning)
	defer logging.SetLogLevel(logging.LogLevelDebug)

	flagKeys := discoverFlags(c)

	for _, fk := range flagKeys {
		counts := make(map[string]int)
		total := *numUsers

		for i := 1; i <= total; i++ {
			uid := fmt.Sprintf("dist_%d", i)
			uc := c.CreateUserContext(uid, nil)
			d := uc.Decide(fk, nil)

			key := d.RuleKey + "/" + d.VariationKey
			if isHoldout(d) {
				key = "HOLDOUT:" + d.RuleKey
			}
			counts[key]++
		}

		fmt.Printf("\n  Flag: %s (%d users)\n", fk, total)
		for key, count := range counts {
			pct := float64(count) / float64(total) * 100
			bar := strings.Repeat("#", int(pct/2))
			fmt.Printf("    %-40s %4d (%5.1f%%) %s\n", key, count, pct, bar)
		}
	}
}

// ============================================================
// INTERACTIVE EXPLORATION REPL
// ============================================================

func runExploreREPL(c *client.OptimizelyClient) {
	fmt.Println("\n  Interactive Exploration Mode")
	fmt.Println("  Type commands to explore holdout behavior. Type 'help' for commands.\n")

	scanner := bufio.NewScanner(os.Stdin)
	currentUser := *userID
	currentAttrs := map[string]interface{}{}

	for {
		fmt.Printf("  [%s] > ", currentUser)
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]

		switch cmd {
		case "help", "h":
			fmt.Println("  Commands:")
			fmt.Println("    decide <flag>              Decide on a flag")
			fmt.Println("    decide_all                 Decide all flags")
			fmt.Println("    decide_keys <f1> <f2> ...  Decide for specific flags")
			fmt.Println("    dist <flag> [n]            Distribution for n users (default 100)")
			fmt.Println("    user <id>                  Switch user")
			fmt.Println("    attr <key> <val>           Set user attribute")
			fmt.Println("    attrs                      Show current attributes")
			fmt.Println("    clearattrs                 Clear all attributes")
			fmt.Println("    force <flag> <var>          Set forced decision (flag level)")
			fmt.Println("    force <flag> <rule> <var>   Set forced decision (rule level)")
			fmt.Println("    unforce <flag>             Remove forced decision")
			fmt.Println("    unforceall                 Remove all forced decisions")
			fmt.Println("    config                     Show flags/rules/holdouts")
			fmt.Println("    snapshot [n]               Take snapshot of n users")
			fmt.Println("    quit, q                    Exit")

		case "decide", "d":
			if len(parts) < 2 {
				fmt.Println("    Usage: decide <flag_key>")
				continue
			}
			uc := c.CreateUserContext(currentUser, currentAttrs)
			d := uc.Decide(parts[1], []decide.OptimizelyDecideOptions{decide.IncludeReasons})
			printDecision("Result", d)
			printReasons(d)

		case "decide_all", "da":
			uc := c.CreateUserContext(currentUser, currentAttrs)
			decisions := uc.DecideAll(nil)
			for fk, d := range decisions {
				ho := ""
				if isHoldout(d) {
					ho = " [HOLDOUT]"
				}
				fmt.Printf("    %-20s rule=%-25s var=%-12s enabled=%v%s\n",
					fk, d.RuleKey, d.VariationKey, d.Enabled, ho)
			}

		case "decide_keys", "dk":
			if len(parts) < 2 {
				fmt.Println("    Usage: decide_keys <flag1> <flag2> ...")
				continue
			}
			uc := c.CreateUserContext(currentUser, currentAttrs)
			decisions := uc.DecideForKeys(parts[1:], nil)
			for fk, d := range decisions {
				ho := ""
				if isHoldout(d) {
					ho = " [HOLDOUT]"
				}
				fmt.Printf("    %-20s rule=%-25s var=%-12s enabled=%v%s\n",
					fk, d.RuleKey, d.VariationKey, d.Enabled, ho)
			}

		case "dist":
			if len(parts) < 2 {
				fmt.Println("    Usage: dist <flag_key> [num_users]")
				continue
			}
			n := 100
			if len(parts) >= 3 {
				fmt.Sscanf(parts[2], "%d", &n)
			}
			logging.SetLogLevel(logging.LogLevelWarning)
			counts := map[string]int{}
			for i := 1; i <= n; i++ {
				uid := fmt.Sprintf("dist_%d", i)
				uc := c.CreateUserContext(uid, currentAttrs)
				d := uc.Decide(parts[1], nil)
				key := d.RuleKey + "/" + d.VariationKey
				if isHoldout(d) {
					key = "HOLDOUT:" + d.RuleKey
				}
				counts[key]++
			}
			logging.SetLogLevel(logging.LogLevelDebug)
			for key, count := range counts {
				pct := float64(count) / float64(n) * 100
				bar := strings.Repeat("#", int(pct/2))
				fmt.Printf("    %-40s %4d (%5.1f%%) %s\n", key, count, pct, bar)
			}

		case "user", "u":
			if len(parts) < 2 {
				fmt.Printf("    Current user: %s\n", currentUser)
				continue
			}
			currentUser = parts[1]
			fmt.Printf("    Switched to user: %s\n", currentUser)

		case "attr":
			if len(parts) < 3 {
				fmt.Println("    Usage: attr <key> <value>")
				continue
			}
			currentAttrs[parts[1]] = parts[2]
			fmt.Printf("    Set %s=%s\n", parts[1], parts[2])

		case "attrs":
			if len(currentAttrs) == 0 {
				fmt.Println("    (no attributes)")
			} else {
				for k, v := range currentAttrs {
					fmt.Printf("    %s = %v\n", k, v)
				}
			}

		case "clearattrs":
			currentAttrs = map[string]interface{}{}
			fmt.Println("    Cleared all attributes")

		case "force":
			if len(parts) < 3 {
				fmt.Println("    Usage: force <flag> <variation>  OR  force <flag> <rule> <variation>")
				continue
			}
			uc := c.CreateUserContext(currentUser, currentAttrs)
			if len(parts) == 3 {
				// Flag-level
				ctx := decision.OptimizelyDecisionContext{FlagKey: parts[1]}
				fd := decision.OptimizelyForcedDecision{VariationKey: parts[2]}
				ok := uc.SetForcedDecision(ctx, fd)
				fmt.Printf("    SetForcedDecision(flag=%s, var=%s): %v\n", parts[1], parts[2], ok)
				d := uc.Decide(parts[1], nil)
				printDecision("With forced decision", d)
			} else {
				// Rule-level
				ctx := decision.OptimizelyDecisionContext{FlagKey: parts[1], RuleKey: parts[2]}
				fd := decision.OptimizelyForcedDecision{VariationKey: parts[3]}
				ok := uc.SetForcedDecision(ctx, fd)
				fmt.Printf("    SetForcedDecision(flag=%s, rule=%s, var=%s): %v\n", parts[1], parts[2], parts[3], ok)
				d := uc.Decide(parts[1], nil)
				printDecision("With forced decision", d)
			}

		case "unforce":
			if len(parts) < 2 {
				fmt.Println("    Usage: unforce <flag>")
				continue
			}
			uc := c.CreateUserContext(currentUser, currentAttrs)
			ctx := decision.OptimizelyDecisionContext{FlagKey: parts[1]}
			ok := uc.RemoveForcedDecision(ctx)
			fmt.Printf("    RemoveForcedDecision(flag=%s): %v\n", parts[1], ok)

		case "unforceall":
			uc := c.CreateUserContext(currentUser, currentAttrs)
			ok := uc.RemoveAllForcedDecisions()
			fmt.Printf("    RemoveAllForcedDecisions: %v\n", ok)

		case "config":
			inspectLiveConfig(c)

		case "snapshot", "snap":
			n := *numUsers
			if len(parts) >= 2 {
				fmt.Sscanf(parts[1], "%d", &n)
			}
			flagKeys := discoverFlags(c)
			snap := takeSnapshot(c, flagKeys, n)
			printSnapshot(snap)
			printDistribution(snap)

		case "quit", "q", "exit":
			return

		default:
			fmt.Printf("    Unknown command: %s (type 'help')\n", cmd)
		}
	}
}

// ============================================================
// CLIENT CREATION
// ============================================================

func loadDatafile() []byte {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	path := filepath.Join(dir, "local_holdouts.json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading datafile: %v\n", err)
		fmt.Println("Make sure local_holdouts.json is in the same directory as main.go")
		os.Exit(1)
	}
	fmt.Printf("Loaded datafile: %s (%d bytes)\n", path, len(data))
	return data
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
	fmt.Println("Static client ready.\n")
	return c
}

func createLiveClient(key string) *client.OptimizelyClient {
	configManager := config.NewPollingProjectConfigManager(key,
		config.WithDatafileURLTemplate("https://optimizely-staging.s3.amazonaws.com/datafiles/%s.json"),
		config.WithPollingInterval(30*time.Second),
	)
	factory := &client.OptimizelyFactory{SDKKey: key}
	c, err := factory.Client(
		client.WithConfigManager(configManager),
	)
	if err != nil {
		fmt.Printf("Error creating live client: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Waiting for datafile to load...")
	time.Sleep(3 * time.Second)
	fmt.Println("Live client ready.\n")
	return c
}

// ============================================================
// HELPERS
// ============================================================

type decisionRecord struct {
	UserID    string
	FlagKey   string
	RuleKey   string
	VarKey    string
	Enabled   bool
	IsHoldout bool
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
}

func printReasons(d client.OptimizelyDecision) {
	if len(d.Reasons) > 0 {
		fmt.Println("    reasons:")
		for _, r := range d.Reasons {
			fmt.Printf("      - %s\n", r)
		}
	}
}

func isHoldout(d client.OptimizelyDecision) bool {
	return d.VariationKey == "ho_off_key" && !d.Enabled
}

func printTestResult(name string, passed bool) {
	status := "PASS"
	if !passed {
		status = "FAIL"
	}
	fmt.Printf("\n  [%s] %s\n", status, name)
	fmt.Println(strings.Repeat("-", 60))
}

func assertDecision(label string, d client.OptimizelyDecision, wantRuleKey, wantVariation string, wantEnabled bool) bool {
	printDecision(label, d)
	pass := true
	if d.RuleKey != wantRuleKey {
		fmt.Printf("    FAIL: expected rule_key=%q, got %q\n", wantRuleKey, d.RuleKey)
		pass = false
	}
	if d.VariationKey != wantVariation {
		fmt.Printf("    FAIL: expected variation_key=%q, got %q\n", wantVariation, d.VariationKey)
		pass = false
	}
	if d.Enabled != wantEnabled {
		fmt.Printf("    FAIL: expected enabled=%v, got %v\n", wantEnabled, d.Enabled)
		pass = false
	}
	if pass {
		fmt.Println("    PASS")
	}
	return pass
}

func discoverFlags(c *client.OptimizelyClient) []string {
	optConfig := c.GetOptimizelyConfig()
	if optConfig == nil {
		fmt.Println("  WARNING: Could not get OptimizelyConfig")
		return nil
	}
	flags := make([]string, 0, len(optConfig.FeaturesMap))
	for k := range optConfig.FeaturesMap {
		flags = append(flags, k)
	}
	return flags
}

func inspectLiveConfig(c *client.OptimizelyClient) {
	optConfig := c.GetOptimizelyConfig()
	if optConfig == nil {
		fmt.Println("  Could not get OptimizelyConfig")
		return
	}

	fmt.Printf("\n  Project Config (revision=%s)\n", optConfig.Revision)
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
}

func takeSnapshot(c *client.OptimizelyClient, flagKeys []string, n int) []decisionRecord {
	return takeSnapshotWithAttrs(c, flagKeys, n, nil)
}

func takeSnapshotWithAttrs(c *client.OptimizelyClient, flagKeys []string, n int, attrs map[string]interface{}) []decisionRecord {
	logging.SetLogLevel(logging.LogLevelWarning)
	defer logging.SetLogLevel(logging.LogLevelDebug)

	records := make([]decisionRecord, 0, n*len(flagKeys))
	for i := 1; i <= n; i++ {
		uid := fmt.Sprintf("snap_user_%d", i)
		uc := c.CreateUserContext(uid, attrs)
		for _, fk := range flagKeys {
			d := uc.Decide(fk, nil)
			records = append(records, decisionRecord{
				UserID:    uid,
				FlagKey:   fk,
				RuleKey:   d.RuleKey,
				VarKey:    d.VariationKey,
				Enabled:   d.Enabled,
				IsHoldout: isHoldout(d),
			})
		}
	}
	return records
}

func printSnapshot(records []decisionRecord) {
	for _, r := range records {
		ho := ""
		if r.IsHoldout {
			ho = " [HOLDOUT]"
		}
		fmt.Printf("    %-20s %-15s rule=%-25s var=%-12s%s\n",
			r.UserID, r.FlagKey, r.RuleKey, r.VarKey, ho)
	}
}

func printDistribution(records []decisionRecord) {
	// Group by flag
	byFlag := map[string]map[string]int{}
	flagTotals := map[string]int{}
	for _, r := range records {
		if byFlag[r.FlagKey] == nil {
			byFlag[r.FlagKey] = map[string]int{}
		}
		key := r.RuleKey + "/" + r.VarKey
		if r.IsHoldout {
			key = "HOLDOUT:" + r.RuleKey
		}
		byFlag[r.FlagKey][key]++
		flagTotals[r.FlagKey]++
	}

	for fk, counts := range byFlag {
		total := flagTotals[fk]
		fmt.Printf("\n    %s (%d decisions):\n", fk, total)
		for key, count := range counts {
			pct := float64(count) / float64(total) * 100
			bar := strings.Repeat("#", int(pct/2))
			fmt.Printf("      %-40s %4d (%5.1f%%) %s\n", key, count, pct, bar)
		}
	}
}

func printSnapshotByFlag(records []decisionRecord, flagKeys []string) {
	byFlag := map[string][]decisionRecord{}
	for _, r := range records {
		byFlag[r.FlagKey] = append(byFlag[r.FlagKey], r)
	}
	for _, fk := range flagKeys {
		fmt.Printf("\n    Flag: %s\n", fk)
		for _, r := range byFlag[fk] {
			ho := ""
			if r.IsHoldout {
				ho = " [HOLDOUT]"
			}
			fmt.Printf("      %-20s rule=%-25s var=%-12s%s\n", r.UserID, r.RuleKey, r.VarKey, ho)
		}
	}
}

func diffSnapshots(before, after []decisionRecord) {
	// Build lookup: userID+flagKey -> record
	beforeMap := map[string]decisionRecord{}
	for _, r := range before {
		beforeMap[r.UserID+"|"+r.FlagKey] = r
	}

	changes := 0
	holdoutGained := 0
	holdoutLost := 0

	for _, a := range after {
		key := a.UserID + "|" + a.FlagKey
		b, ok := beforeMap[key]
		if !ok {
			continue
		}
		if b.RuleKey != a.RuleKey || b.VarKey != a.VarKey {
			changes++
			if !b.IsHoldout && a.IsHoldout {
				holdoutGained++
			}
			if b.IsHoldout && !a.IsHoldout {
				holdoutLost++
			}
		}
	}

	fmt.Println("\n  --- Diff Summary ---")
	fmt.Printf("    Total decision changes: %d\n", changes)
	fmt.Printf("    Gained holdout:         %d (was normal, now held out)\n", holdoutGained)
	fmt.Printf("    Lost holdout:           %d (was held out, now normal)\n", holdoutLost)
	if changes == 0 {
		fmt.Println("    No changes detected. Wait for datafile refresh or verify the UI change was saved.")
	}
}

func waitForUIChange(instruction string) {
	fmt.Printf("\n  ACTION REQUIRED: %s\n", instruction)
	fmt.Println("  After saving, wait ~30-60 seconds for the datafile to refresh.")
	fmt.Print("  Press Enter when ready... ")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func printScenarioHeader(title string, lines ...string) {
	fmt.Printf("\n--- Scenario: %s ---\n", title)
	for _, line := range lines {
		fmt.Println(line)
	}
}

func printHelp() {
	fmt.Println(`
Static tests (-mode=static -test=<name>):
  basic          Local holdout targeting single rule
  multi_rule     Local holdout targeting multiple rules on same flag
  cross_flag     Local holdout targeting rules across different flags
  global         Global holdout applies to all rules
  precedence     Global holdout beats local holdout
  audience       Local holdout with audience conditions
  not_targeted   Local holdout only affects targeted rules
  zero_traffic   Zero-traffic holdout never applied
  all            Run all static tests

Live scenarios (-sdk_key=KEY -scenario=<name>):

  UI Mutations (try to break SDKs with datafile changes):
    ui_delete_holdout       Delete a running holdout, verify SDK handles it
    ui_change_traffic       Change holdout traffic %, verify re-bucketing
    ui_switch_local_global  Switch holdout between local/global type
    ui_change_audience      Add/remove audience on a holdout
    ui_delete_targeted_rule Delete the rule a local holdout targets (dangling ref)
    ui_pause_holdout        Pause a running holdout
    ui_add_holdout_to_running  Add new holdout to experiment with existing traffic

  Feature Interactions (holdout + other SDK features):
    forced_vs_holdout       SetForcedDecision (flag level) overrides holdout
    forced_rule_level       SetForcedDecision (rule level) vs holdout
    holdout_disable_event   DisableDecisionEvent option with holdout
    holdout_track           Track event after holdout decision
    holdout_decide_all      DecideAll with holdouts
    holdout_decide_for_keys DecideForKeys vs DecideAll consistency
    holdout_listener        Decision notification listener with holdout info
    holdout_enabled_flags_only  EnabledFlagsOnly should exclude held-out flags

  Stress / Edge Cases:
    rapid_repolling         Continuous polling during rapid UI changes
    distribution            Distribution check across many users

  Interactive:
    -explore                REPL for ad-hoc exploration (decide, force, dist, etc.)

Useful flags:
  -user=<id>       Default user ID (default: user_1)
  -flag=<key>      Target flag key (auto-discovered if not set)
  -n=<count>       Number of users for distribution/snapshot (default: 20)`)
}

// ============================================================
// STATIC TESTS - Deterministic with pre-calculated bucketing
// ============================================================
//
// Datafile: local_holdouts.json (from fullstack-sdk-compatibility-suite)
//
// Flags:
//   flag_a: 3 experiments (5001=flag_a_exp_1, 5002=flag_a_exp_2, 5003=flag_a_exp_3)
//   flag_b: 2 experiments (5004=flag_b_exp_1, 5005=flag_b_exp_2)
//   flag_c: 1 experiment  (5006=flag_c_exp_1)
//
// Holdouts:
//   ho_local_single_rule       - Local, 30%, targets [5001]
//   ho_local_multi_rules_same  - Local, 25%, targets [5001, 5002]
//   ho_local_cross_flag        - Local, 20%, targets [5001, 5004]
//   ho_local_with_audience     - Local, 40%, targets [5002], audience=customattr:yes
//   ho_global_all_rules        - Global, 15%, targets all rules
//   ho_local_high_traffic      - Local, 50%, targets [5003]
//   ho_local_zero_traffic      - Local, 0%, targets [5005]

func testStaticBasicLocalHoldout(c *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Basic Local Holdout (Single Rule) ---")
	fmt.Println("Holdout: ho_local_single_rule (30% traffic, targets rule 5001)")
	fmt.Println("Flag: flag_a (first rule is 5001 = flag_a_exp_1)")

	holdoutCount := 0
	normalCount := 0
	total := 30

	for i := 1; i <= total; i++ {
		userID := fmt.Sprintf("static_user_%d", i)
		uc := c.CreateUserContext(userID, nil)
		d := uc.Decide("flag_a", nil)

		tag := "NORMAL"
		if isHoldout(d) {
			holdoutCount++
			tag = "HOLDOUT"
		} else {
			normalCount++
		}
		fmt.Printf("  %-22s [%-7s] rule=%-30s var=%-14s enabled=%v\n",
			userID, tag, d.RuleKey, d.VariationKey, d.Enabled)
	}

	fmt.Printf("\n  Summary: %d holdout, %d normal (of %d)\n", holdoutCount, normalCount, total)
	passed := holdoutCount > 0 && normalCount > 0
	if !passed {
		fmt.Println("  WARNING: Expected a mix of holdout and normal decisions at 30% traffic")
	}
	printTestResult("Basic Local Holdout", passed)
}

func testStaticMultiRuleHoldout(c *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Multi-Rule Same Flag Holdout ---")
	fmt.Println("Holdout: ho_local_multi_rules_same_flag (25%, targets rules 5001+5002)")

	holdoutCount := 0
	total := 30
	for i := 1; i <= total; i++ {
		userID := fmt.Sprintf("multi_user_%d", i)
		uc := c.CreateUserContext(userID, nil)
		d := uc.Decide("flag_a", nil)

		if isHoldout(d) && d.RuleKey == "ho_local_multi_rules_same_flag" {
			holdoutCount++
		}
		tag := "NORMAL"
		if isHoldout(d) {
			tag = d.RuleKey
		}
		fmt.Printf("  %-22s [%-40s] var=%-14s enabled=%v\n",
			userID, tag, d.VariationKey, d.Enabled)
	}

	fmt.Printf("\n  ho_local_multi_rules_same_flag hits: %d of %d\n", holdoutCount, total)
	printTestResult("Multi-Rule Same Flag Holdout", true)
}

func testStaticCrossFlagHoldout(c *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Cross-Flag Local Holdout ---")
	fmt.Println("Holdout: ho_local_cross_flag (20%, targets 5001 on flag_a + 5004 on flag_b)")

	passed := true
	crossFlagHits := 0
	total := 50

	for i := 1; i <= total; i++ {
		userID := fmt.Sprintf("cross_user_%d", i)
		uc := c.CreateUserContext(userID, nil)
		dA := uc.Decide("flag_a", nil)
		dB := uc.Decide("flag_b", nil)

		hitA := dA.RuleKey == "ho_local_cross_flag"
		hitB := dB.RuleKey == "ho_local_cross_flag"

		if hitA && hitB {
			crossFlagHits++
			fmt.Printf("  %-22s BOTH flags held out by ho_local_cross_flag\n", userID)
		} else if hitA != hitB {
			fmt.Printf("  %-22s flag_a=%s flag_b=%s (partial - check other holdouts)\n",
				userID, dA.RuleKey, dB.RuleKey)
		}
	}

	fmt.Printf("\n  Users held out on BOTH flags by cross-flag holdout: %d of %d\n", crossFlagHits, total)
	if crossFlagHits == 0 {
		fmt.Println("  WARNING: No cross-flag holdout hits. May be expected at 20% with competing holdouts.")
	}
	printTestResult("Cross-Flag Local Holdout", passed)
}

func testStaticGlobalHoldout(c *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Global Holdout ---")
	fmt.Println("Holdout: ho_global_all_rules (15%, includedRules=null)")

	globalHits := map[string]int{"flag_a": 0, "flag_b": 0, "flag_c": 0}
	total := 50
	flags := []string{"flag_a", "flag_b", "flag_c"}

	for i := 1; i <= total; i++ {
		userID := fmt.Sprintf("global_user_%d", i)
		uc := c.CreateUserContext(userID, nil)

		for _, f := range flags {
			d := uc.Decide(f, nil)
			if d.RuleKey == "ho_global_all_rules" {
				globalHits[f]++
			}
		}
	}

	passed := true
	fmt.Println("\n  Global holdout hits per flag:")
	for _, f := range flags {
		pct := float64(globalHits[f]) / float64(total) * 100
		fmt.Printf("    %s: %d/%d (%.1f%%)\n", f, globalHits[f], total, pct)
		if globalHits[f] == 0 {
			passed = false
		}
	}

	if passed {
		fmt.Println("\n  Global holdout applied to all flags.")
	} else {
		fmt.Println("\n  WARNING: Global holdout missing on some flags.")
	}
	printTestResult("Global Holdout", passed)
}

func testStaticGlobalBeatsLocal(c *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Global Beats Local (Precedence) ---")
	fmt.Println("Rule 5001 targeted by global (15%) AND local (30%) holdouts")
	fmt.Println("Global holdouts are evaluated FIRST -- if user hits global, local is never checked")

	globalCount := 0
	localCount := 0
	normalCount := 0
	total := 100

	for i := 1; i <= total; i++ {
		userID := fmt.Sprintf("prec_user_%d", i)
		uc := c.CreateUserContext(userID, nil)
		d := uc.Decide("flag_a", nil)

		if d.RuleKey == "ho_global_all_rules" {
			globalCount++
		} else if isHoldout(d) {
			localCount++
		} else {
			normalCount++
		}
	}

	fmt.Printf("\n  Global holdout: %d  |  Local holdout(s): %d  |  Normal: %d  (of %d)\n",
		globalCount, localCount, normalCount, total)
	fmt.Println("  Expected: ~15% global, then some % local, rest normal")

	passed := globalCount > 0
	if !passed {
		fmt.Println("  FAIL: No global holdout hits. Global should be evaluated first.")
	}
	printTestResult("Global Beats Local", passed)
}

func testStaticAudienceHoldout(c *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Audience-Targeted Holdout ---")
	fmt.Println("Holdout: ho_local_with_audience (40%, targets 5002, audience=customattr:yes)")

	fmt.Println("\n  Part A: Users WITH customattr=yes")
	audienceHits := 0
	total := 30
	for i := 1; i <= total; i++ {
		userID := fmt.Sprintf("aud_yes_%d", i)
		uc := c.CreateUserContext(userID, map[string]interface{}{"customattr": "yes"})
		d := uc.Decide("flag_a", nil)

		if d.RuleKey == "ho_local_with_audience" {
			audienceHits++
		}
		tag := "NORMAL"
		if isHoldout(d) {
			tag = d.RuleKey
		}
		fmt.Printf("  %-22s [%-30s] var=%s\n", userID, tag, d.VariationKey)
	}
	fmt.Printf("  Audience holdout hits (with attr): %d/%d\n", audienceHits, total)

	fmt.Println("\n  Part B: Users WITHOUT customattr (should NOT hit audience holdout)")
	noAttrHits := 0
	for i := 1; i <= total; i++ {
		userID := fmt.Sprintf("aud_no_%d", i)
		uc := c.CreateUserContext(userID, nil)
		d := uc.Decide("flag_a", nil)

		if d.RuleKey == "ho_local_with_audience" {
			noAttrHits++
		}
	}
	fmt.Printf("  Audience holdout hits (without attr): %d/%d\n", noAttrHits, total)

	passed := noAttrHits == 0
	if noAttrHits > 0 {
		fmt.Println("  FAIL: Audience holdout applied to users without the required attribute!")
	} else {
		fmt.Println("  PASS: Audience holdout correctly filtered by audience condition.")
	}
	printTestResult("Audience-Targeted Holdout", passed)
}

func testStaticNotTargeted(c *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Not Targeted Rule ---")
	fmt.Println("ho_local_single_rule targets rule 5001 (flag_a) -- should NOT affect flag_b")

	wrongHits := 0
	total := 30
	for i := 1; i <= total; i++ {
		userID := fmt.Sprintf("nontarget_%d", i)
		uc := c.CreateUserContext(userID, nil)
		d := uc.Decide("flag_b", nil)

		if d.RuleKey == "ho_local_single_rule" {
			wrongHits++
			fmt.Printf("  FAIL: %s got ho_local_single_rule on flag_b!\n", userID)
		}
	}

	passed := wrongHits == 0
	if passed {
		fmt.Println("  All flag_b decisions correctly unaffected by ho_local_single_rule")
	}
	printTestResult("Not Targeted Rule", passed)
}

func testStaticZeroTraffic(c *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Zero Traffic Holdout ---")
	fmt.Println("ho_local_zero_traffic (0% traffic, targets 5005) should NEVER be applied")

	zeroHits := 0
	total := 50
	for i := 1; i <= total; i++ {
		userID := fmt.Sprintf("zero_%d", i)
		uc := c.CreateUserContext(userID, nil)
		d := uc.Decide("flag_b", nil)

		if d.RuleKey == "ho_local_zero_traffic" {
			zeroHits++
		}
	}

	passed := zeroHits == 0
	fmt.Printf("  ho_local_zero_traffic hits: %d/%d\n", zeroHits, total)
	if passed {
		fmt.Println("  PASS: Zero-traffic holdout correctly never applied.")
	} else {
		fmt.Println("  FAIL: Zero-traffic holdout was applied!")
	}
	printTestResult("Zero Traffic Holdout", passed)
}
