This branch (`russell-demo-gosdk-cmab`) contains a working CMAB demonstration that shows the full prediction flow from the Go SDK.

## Quick Setup

1. **Switch to the demo branch:**
   ```bash
   git checkout russell-demo-gosdk-cmab
   ```

2. **Run the CMAB demo:**
   ```bash
   cd examples
   go run main.go
   ```

## What You'll See

The demo will show detailed logs including:

1. **Datafile fetch** from develrc environment
2. **CMAB prediction endpoint URL** with hardcoded ruleID 1304618
3. **CMAB raw response** from the prediction service:
   ```json
   {"predictions":[{"variation_id":"838409"}]}
   ```
4. **Final decision result** showing `Enabled: true`, `Variation: on`

## Expected Output

You should see something like:
```
[DefaultCmabClient] CMAB prediction endpoint URL: https://inte.prediction.cmab.optimizely.com/predict/1304618 (original ruleID: 9000003705922)
[DefaultCmabClient] CMAB request body: {"instances":[{"visitorId":"user123","experimentId":"1304618","attributes":[],"cmabUUID":"0ef64db8-affb-479b-87c6-d5be0b882c36"}]}
[DefaultCmabClient] CMAB raw response: {"predictions":[{"variation_id":"838409"}]}
[DefaultCmabClient] CMAB parsed variation ID: 838409
=== DECISION RESULT ===
Enabled: true
Variation: on
Rule: cmab-for-matjaz
```

## Key Technical Details

- **SDK Key**: `JgzFaGzGXx6F1ocTbMTmn` (Matjaz's develrc project)
- **Feature Flag**: `flag-matjaz-editor`
- **CMAB Experiment**: `cmab-for-matjaz`
- **Test Rule ID**: `1304618` (hardcoded for demo)
- **Environment**: inte (integration) CMAB prediction service
- **JSON Format**: camelCase (`visitorId`, `experimentId`, `cmabUUID`)
