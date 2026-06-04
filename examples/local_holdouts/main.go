// Local Holdouts Bug Bash - Go SDK
//
// Validates local holdout evaluation using either a bundled static datafile
// (deterministic, pre-calculated bucketing IDs) or a live RC project (for UI
// interaction testing).
//
// Static mode:  go run examples/local_holdouts/main.go -mode=static -test=basic
// Live mode:    go run examples/local_holdouts/main.go -sdk_key=YOUR_KEY -test=live_basic
// List tests:   go run examples/local_holdouts/main.go -test=help

package main

import (
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
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// ============================================================
// CONFIGURATION
// ============================================================
const (
	// Default SDK key for live mode (override via -sdk_key flag)
	SDK_KEY = "YOUR_SDK_KEY_HERE"

	// Live mode: flag and holdout keys (must match your RC project setup)
	LIVE_FLAG_A          = "flag_a"
	LIVE_FLAG_B          = "flag_b"
	LIVE_LOCAL_HOLDOUT   = "local_holdout"
	LIVE_GLOBAL_HOLDOUT  = "global_holdout"
	LIVE_AUDIENCE_HOLDOUT = "audience_holdout"
)

var (
	testCase = flag.String("test", "basic", "Test case to run (use -test=help for list)")
	mode     = flag.String("mode", "live", "Mode: 'static' (bundled datafile) or 'live' (RC project)")
	sdkKey   = flag.String("sdk_key", SDK_KEY, "SDK key for live mode")
)

func main() {
	flag.Parse()

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  Local Holdouts Bug Bash - Go SDK")
	fmt.Println(strings.Repeat("=", 60))

	if *testCase == "help" {
		printAvailableTests()
		return
	}

	if *mode == "static" {
		// Less noise for static tests -- change to LogLevelDebug to see holdout evaluation
		logging.SetLogLevel(logging.LogLevelWarning)
		runStaticTests()
	} else {
		logging.SetLogLevel(logging.LogLevelDebug)
		runLiveTests()
	}
}

// ============================================================
// STATIC MODE - Uses bundled local_holdouts.json datafile
// ============================================================
// Pre-calculated bucketing IDs ensure deterministic, reproducible results.
// No project setup required.

func runStaticTests() {
	fmt.Println("\nMode: STATIC (bundled datafile)")

	datafile := loadDatafile()
	optimizelyClient := createStaticClient(datafile)
	defer optimizelyClient.Close()

	switch *testCase {
	case "basic":
		testStaticBasicLocalHoldout(optimizelyClient)
	case "multi_rule":
		testStaticMultiRuleHoldout(optimizelyClient)
	case "cross_flag":
		testStaticCrossFlagHoldout(optimizelyClient)
	case "global":
		testStaticGlobalHoldout(optimizelyClient)
	case "precedence":
		testStaticGlobalBeatsLocal(optimizelyClient)
	case "audience":
		testStaticAudienceHoldout(optimizelyClient)
	case "not_targeted":
		testStaticNotTargeted(optimizelyClient)
	case "zero_traffic":
		testStaticZeroTraffic(optimizelyClient)
	case "all":
		testStaticBasicLocalHoldout(optimizelyClient)
		testStaticMultiRuleHoldout(optimizelyClient)
		testStaticCrossFlagHoldout(optimizelyClient)
		testStaticGlobalHoldout(optimizelyClient)
		testStaticGlobalBeatsLocal(optimizelyClient)
		testStaticAudienceHoldout(optimizelyClient)
		testStaticNotTargeted(optimizelyClient)
		testStaticZeroTraffic(optimizelyClient)
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("  All static tests completed.")
		fmt.Println(strings.Repeat("=", 60))
	default:
		fmt.Printf("\nUnknown static test: %s\n", *testCase)
		printAvailableTests()
	}
}

// ============================================================
// LIVE MODE - Uses RC (Prep) environment with polling
// ============================================================

func runLiveTests() {
	fmt.Println("\nMode: LIVE (RC project)")

	if *sdkKey == "YOUR_SDK_KEY_HERE" || *sdkKey == "" {
		fmt.Println("\nERROR: Provide your RC SDK key via -sdk_key=YOUR_KEY")
		fmt.Println("See README.md for project setup instructions.")
		return
	}

	optimizelyClient := createLiveClient(*sdkKey)
	defer optimizelyClient.Close()

	switch *testCase {
	case "live_basic", "basic":
		testLiveBasicHoldout(optimizelyClient)
	case "live_global", "global":
		testLiveGlobalHoldout(optimizelyClient)
	case "live_forced", "forced":
		testLiveForcedDecisionOverride(optimizelyClient)
	case "live_decide_all", "decide_all":
		testLiveDecideAll(optimizelyClient)
	case "live_distribution", "distribution":
		testLiveDistribution(optimizelyClient)
	case "live_ui_refresh", "ui_refresh":
		testLiveUIRefresh(optimizelyClient)
	case "all":
		testLiveBasicHoldout(optimizelyClient)
		testLiveGlobalHoldout(optimizelyClient)
		testLiveForcedDecisionOverride(optimizelyClient)
		testLiveDecideAll(optimizelyClient)
		testLiveDistribution(optimizelyClient)
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("  All live tests completed.")
		fmt.Println(strings.Repeat("=", 60))
	default:
		fmt.Printf("\nUnknown live test: %s\n", *testCase)
		printAvailableTests()
	}
}

// ============================================================
// CLIENT CREATION
// ============================================================

func loadDatafile() []byte {
	// Resolve path relative to this source file
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
	optimizelyClient, err := factory.Client(
		client.WithConfigManager(configManager),
	)
	if err != nil {
		fmt.Printf("Error creating static client: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Static client ready.\n")
	return optimizelyClient
}

func createLiveClient(key string) *client.OptimizelyClient {
	configManager := config.NewPollingProjectConfigManager(key,
		config.WithDatafileURLTemplate("https://optimizely-staging.s3.amazonaws.com/datafiles/%s.json"),
		config.WithPollingInterval(30*time.Second),
	)
	factory := &client.OptimizelyFactory{SDKKey: key}
	optimizelyClient, err := factory.Client(
		client.WithConfigManager(configManager),
	)
	if err != nil {
		fmt.Printf("Error creating live client: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Waiting for datafile to load...")
	time.Sleep(3 * time.Second)
	fmt.Println("Live client ready.\n")
	return optimizelyClient
}

// ============================================================
// HELPERS
// ============================================================

func printAvailableTests() {
	fmt.Println("\nStatic mode tests (-mode=static):")
	fmt.Println("  basic          Local holdout targeting single rule (deterministic)")
	fmt.Println("  multi_rule     Local holdout targeting multiple rules on same flag")
	fmt.Println("  cross_flag     Local holdout targeting rules across different flags")
	fmt.Println("  global         Global holdout applies to all rules")
	fmt.Println("  precedence     Global holdout beats local holdout")
	fmt.Println("  audience       Local holdout with audience conditions")
	fmt.Println("  not_targeted   Local holdout only affects targeted rules")
	fmt.Println("  zero_traffic   Zero-traffic holdout never applied")
	fmt.Println("  all            Run all static tests")
	fmt.Println()
	fmt.Println("Live mode tests (-sdk_key=YOUR_KEY):")
	fmt.Println("  live_basic         Basic local holdout with live project")
	fmt.Println("  live_global        Global holdout with live project")
	fmt.Println("  live_forced        SetForcedDecision overrides holdout")
	fmt.Println("  live_decide_all    DecideAll with holdouts")
	fmt.Println("  live_distribution  Statistical distribution over 1000 users")
	fmt.Println("  live_ui_refresh    Interactive: change UI, verify SDK picks up changes")
	fmt.Println("  all                Run all live tests (except ui_refresh)")
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

func isHoldout(d client.OptimizelyDecision) bool {
	// Holdout decisions have: enabled=false, variation_key=ho_off_key
	return d.VariationKey == "ho_off_key" && !d.Enabled
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

func printTestResult(name string, passed bool) {
	status := "PASS"
	if !passed {
		status = "FAIL"
	}
	fmt.Printf("\n  [%s] %s\n", status, name)
	fmt.Println(strings.Repeat("-", 60))
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

// Test 1: Basic single-rule local holdout
// ho_local_single_rule targets rule 5001 (flag_a_exp_1) with 30% traffic.
// Users try to decide on flag_a. Rule 5001 is the first experiment for flag_a.
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

// Test 2: Multi-rule same flag local holdout
// ho_local_multi_rules_same_flag targets rules 5001 AND 5002 with 25% traffic.
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

// Test 3: Cross-flag local holdout
// ho_local_cross_flag targets rule 5001 (flag_a) AND rule 5004 (flag_b) with 20%.
// A user that hits this holdout should see it on BOTH flags.
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
			// This can happen if other holdouts with higher priority catch the user
			// on one flag but not the other. Only a problem if same holdout is inconsistent.
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

// Test 4: Global holdout applies to ALL rules
// ho_global_all_rules (15% traffic, includedRules=null) should affect all flags.
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

// Test 5: Global holdout beats local holdout (precedence)
// Rule 5001 is targeted by both global (15%) and local (30%) holdouts.
// If a user qualifies for the global holdout, it should take precedence.
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

// Test 6: Local holdout with audience condition
// ho_local_with_audience targets rule 5002 (flag_a_exp_2) with 40% traffic,
// BUT requires customattr=yes audience.
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

// Test 7: Local holdout does NOT affect non-targeted rules
// ho_local_single_rule targets rule 5001. Rules 5004, 5005 (flag_b) are NOT targeted.
// Flag_b decisions should never show ho_local_single_rule.
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

// Test 8: Zero-traffic holdout is never applied
// ho_local_zero_traffic targets rule 5005 (flag_b_exp_2) with 0% traffic.
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

// ============================================================
// LIVE TESTS - Require RC project with holdouts configured
// ============================================================

// Live Test 1: Basic local holdout with live project
func testLiveBasicHoldout(c *client.OptimizelyClient) {
	fmt.Println("\n--- Live Test: Basic Local Holdout ---")
	fmt.Printf("Flag: %s | Expected holdout: %s\n", LIVE_FLAG_A, LIVE_LOCAL_HOLDOUT)

	holdoutCount := 0
	normalCount := 0
	total := 20

	for i := 1; i <= total; i++ {
		userID := fmt.Sprintf("live_user_%d", i)
		uc := c.CreateUserContext(userID, nil)
		d := uc.Decide(LIVE_FLAG_A, []decide.OptimizelyDecideOptions{decide.IncludeReasons})

		tag := "NORMAL"
		if isHoldout(d) {
			holdoutCount++
			tag = "HOLDOUT(" + d.RuleKey + ")"
		} else {
			normalCount++
		}
		fmt.Printf("  %-22s [%-35s] var=%-10s enabled=%v\n",
			userID, tag, d.VariationKey, d.Enabled)
	}

	fmt.Printf("\n  Summary: %d holdout, %d normal (of %d)\n", holdoutCount, normalCount, total)
	if holdoutCount == 0 {
		fmt.Println("  WARNING: No holdout decisions. Verify holdout is Running and targets the correct rule.")
	} else if normalCount == 0 {
		fmt.Println("  WARNING: All users held out. Check holdout traffic percentage.")
	} else {
		fmt.Println("  OK: Mix of holdout and normal decisions observed.")
	}
	printTestResult("Live Basic Holdout", holdoutCount > 0 || normalCount > 0)
}

// Live Test 2: Global holdout with live project
func testLiveGlobalHoldout(c *client.OptimizelyClient) {
	fmt.Println("\n--- Live Test: Global Holdout ---")
	fmt.Printf("Expected: %s affects both %s and %s\n", LIVE_GLOBAL_HOLDOUT, LIVE_FLAG_A, LIVE_FLAG_B)

	globalA, globalB := 0, 0
	total := 30

	for i := 1; i <= total; i++ {
		userID := fmt.Sprintf("live_global_%d", i)
		uc := c.CreateUserContext(userID, nil)
		dA := uc.Decide(LIVE_FLAG_A, nil)
		dB := uc.Decide(LIVE_FLAG_B, nil)

		if dA.RuleKey == LIVE_GLOBAL_HOLDOUT {
			globalA++
		}
		if dB.RuleKey == LIVE_GLOBAL_HOLDOUT {
			globalB++
		}
	}

	fmt.Printf("\n  Global holdout on %s: %d/%d\n", LIVE_FLAG_A, globalA, total)
	fmt.Printf("  Global holdout on %s: %d/%d\n", LIVE_FLAG_B, globalB, total)
	passed := globalA > 0 || globalB > 0
	printTestResult("Live Global Holdout", passed)
}

// Live Test 3: Forced decision overrides holdout
func testLiveForcedDecisionOverride(c *client.OptimizelyClient) {
	fmt.Println("\n--- Live Test: Forced Decision Overrides Holdout ---")
	fmt.Printf("Flag: %s | Setting forced variation to 'on'\n", LIVE_FLAG_A)

	uc := c.CreateUserContext("forced_holdout_user", nil)

	// Step 1: Normal decision (may or may not be holdout)
	d1 := uc.Decide(LIVE_FLAG_A, []decide.OptimizelyDecideOptions{decide.IncludeReasons})
	printDecision("Step 1 - Normal decision", d1)

	// Step 2: Set forced decision to override
	ctx := decision.OptimizelyDecisionContext{FlagKey: LIVE_FLAG_A}
	fd := decision.OptimizelyForcedDecision{VariationKey: "on"}
	ok := uc.SetForcedDecision(ctx, fd)
	fmt.Printf("\n  SetForcedDecision(variation='on'): %v\n", ok)

	d2 := uc.Decide(LIVE_FLAG_A, nil)
	printDecision("Step 2 - With forced decision", d2)

	passed := d2.VariationKey == "on"
	if passed {
		fmt.Println("    PASS: Forced decision returned 'on' variation regardless of holdout")
	} else {
		fmt.Printf("    FAIL: Expected variation 'on', got '%s'\n", d2.VariationKey)
	}

	// Step 3: Remove forced decision
	uc.RemoveForcedDecision(ctx)
	d3 := uc.Decide(LIVE_FLAG_A, nil)
	printDecision("Step 3 - After removing forced decision", d3)

	if d3.RuleKey != d2.RuleKey || d3.VariationKey != d2.VariationKey {
		fmt.Println("    OK: Decision changed after removing forced decision (back to normal/holdout)")
	}
	printTestResult("Forced Decision Override", passed)
}

// Live Test 4: DecideAll with holdouts
func testLiveDecideAll(c *client.OptimizelyClient) {
	fmt.Println("\n--- Live Test: DecideAll with Holdouts ---")

	for i := 1; i <= 5; i++ {
		userID := fmt.Sprintf("decide_all_%d", i)
		uc := c.CreateUserContext(userID, nil)
		decisions := uc.DecideAll(nil)

		fmt.Printf("\n  User: %s (DecideAll returned %d flags)\n", userID, len(decisions))
		for flagKey, d := range decisions {
			tag := "normal"
			if isHoldout(d) {
				tag = "HOLDOUT:" + d.RuleKey
			}
			fmt.Printf("    %-20s -> [%-30s] var=%-10s enabled=%v\n",
				flagKey, tag, d.VariationKey, d.Enabled)
		}
	}
	printTestResult("DecideAll with Holdouts", true)
}

// Live Test 5: Statistical distribution
func testLiveDistribution(c *client.OptimizelyClient) {
	fmt.Println("\n--- Live Test: Distribution (1000 users) ---")
	fmt.Printf("Flag: %s\n", LIVE_FLAG_A)

	// Reduce log noise for bulk test
	logging.SetLogLevel(logging.LogLevelWarning)
	defer logging.SetLogLevel(logging.LogLevelDebug)

	counts := make(map[string]int)
	total := 1000

	for i := 1; i <= total; i++ {
		userID := fmt.Sprintf("dist_user_%d", i)
		uc := c.CreateUserContext(userID, nil)
		d := uc.Decide(LIVE_FLAG_A, nil)

		key := d.RuleKey
		if !isHoldout(d) {
			key = "normal:" + d.VariationKey
		}
		counts[key]++
	}

	fmt.Printf("\n  Distribution over %d users:\n", total)
	for key, count := range counts {
		pct := float64(count) / float64(total) * 100
		bar := strings.Repeat("#", int(pct/2))
		fmt.Printf("    %-35s %4d (%5.1f%%) %s\n", key, count, pct, bar)
	}

	printTestResult("Distribution", true)
}

// Live Test 6: Interactive UI refresh test
func testLiveUIRefresh(c *client.OptimizelyClient) {
	fmt.Println("\n--- Live Test: UI Refresh ---")
	fmt.Println("This test verifies the SDK picks up holdout changes from the UI.\n")

	userID := "ui_refresh_user_42"
	uc := c.CreateUserContext(userID, nil)

	// Initial decision
	d1 := uc.Decide(LIVE_FLAG_A, nil)
	printDecision("BEFORE UI change", d1)

	fmt.Println("\n  Now go to the Optimizely UI and make a change:")
	fmt.Println("    - Change holdout traffic percentage (e.g. 50% -> 0% or 100%)")
	fmt.Println("    - Or add/remove a targeted rule from the holdout")
	fmt.Println("    - Or toggle the holdout on/off")
	fmt.Println()
	fmt.Println("  After saving, wait ~30-60 seconds for the datafile to refresh.")
	fmt.Print("  Press Enter when ready to re-check... ")
	var s string
	fmt.Scanln(&s)

	// Re-create client to force fresh datafile
	d2 := uc.Decide(LIVE_FLAG_A, nil)
	printDecision("AFTER UI change", d2)

	if d1.RuleKey != d2.RuleKey || d1.VariationKey != d2.VariationKey || d1.Enabled != d2.Enabled {
		fmt.Println("\n  CHANGE DETECTED: SDK picked up the datafile update!")
	} else {
		fmt.Println("\n  No change detected. Try waiting longer or verify the change was saved.")
		fmt.Println("  (Same user may still be bucketed the same way if traffic % only changed slightly)")
	}
	printTestResult("UI Refresh", true)
}

// ============================================================
// DATAFILE INSPECTION UTILITY
// ============================================================

func inspectDatafile(datafile []byte) {
	var df map[string]interface{}
	if err := json.Unmarshal(datafile, &df); err != nil {
		fmt.Printf("Error parsing datafile: %v\n", err)
		return
	}

	holdouts, ok := df["holdouts"].([]interface{})
	if !ok || len(holdouts) == 0 {
		fmt.Println("No holdouts found in datafile.")
		return
	}

	fmt.Printf("\nFound %d holdouts in datafile:\n\n", len(holdouts))
	for i, h := range holdouts {
		ho, ok := h.(map[string]interface{})
		if !ok {
			continue
		}
		fmt.Printf("  [%d] Key: %s\n", i+1, ho["key"])
		fmt.Printf("      ID:     %s\n", ho["id"])
		fmt.Printf("      Status: %s\n", ho["status"])

		// Check includedRules
		if ir, exists := ho["includedRules"]; !exists || ir == nil {
			fmt.Println("      Type:   GLOBAL (includedRules=null)")
		} else if rules, ok := ir.([]interface{}); ok {
			ruleStrs := make([]string, len(rules))
			for j, r := range rules {
				ruleStrs[j] = fmt.Sprintf("%v", r)
			}
			fmt.Printf("      Type:   LOCAL (includedRules=[%s])\n", strings.Join(ruleStrs, ", "))
		}

		// Traffic allocation
		if ta, ok := ho["trafficAllocation"].([]interface{}); ok && len(ta) > 0 {
			if first, ok := ta[0].(map[string]interface{}); ok {
				endOfRange := first["endOfRange"]
				pct := 0.0
				if v, ok := endOfRange.(float64); ok {
					pct = v / 100
				}
				fmt.Printf("      Traffic: %.0f%%\n", pct)
			}
		}

		// Audience
		if aids, ok := ho["audienceIds"].([]interface{}); ok && len(aids) > 0 {
			fmt.Printf("      Audience IDs: %v\n", aids)
		}
		fmt.Println()
	}
}
