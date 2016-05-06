# Goptimizely

Optimizely Server Side SDK.... in Go.

## Installation / Basic Usage

TL;DR: just `yy` and `p` from the `examples` directory.

1. Install the Optimizely Go SDK with `go get`:

        go get github.com/optimizely/go-sdk

2. Import the Optimizely SDK:

        import optimizely "github.com/optimizely/go-sdk/optimizely"

3. Create a new `OptimizelyClient` object with your Account ID:

        optimizely_client , err := optimizely.GetOptimizelyClient(OPTIMIZELY_ACCOUNT_ID)

4. Do something cool with the Optimizely client!

        optimizely_client.Track(.....)
