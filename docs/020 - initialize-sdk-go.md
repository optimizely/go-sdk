---
title: "Initialize SDK"
excerpt: ""
slug: "initialize-sdk-go"
hidden: false
createdAt: "2019-08-21T21:13:33.317Z"
updatedAt: "2019-08-21T21:13:45.319Z"
---
To initialize the **OptimizelyClient** you will need either the SDK key or hard-coded JSON datafile.
```go
import "github.com/optimizely/go-sdk/pkg/client"

optimizelyFactory := &client.OptimizelyFactory{
            SDKKey: "[SDK_KEY_HERE]",         
            Datafile: []byte("DATAFILE_JSON_STRING_HERE")
}

// Instantiate a static client (no datafile polling)
optlyClient, err := optimizelyFactory.StaticClient()

// Instantiates a client that syncs the datafile in the background
optlyClient, err := optimizelyFactory.Client()
```