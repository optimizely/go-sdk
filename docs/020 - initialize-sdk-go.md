---
title: "Initialize SDK"
slug: "initialize-sdk-go"
hidden: false
createdAt: "2019-08-21T21:13:33.317Z"
updatedAt: "2019-08-21T21:13:45.319Z"
---
To initialize the **OptimizelyClient** you will need either the SDK key or hard-coded JSON datafile.
[block:code]
{
  "codes": [
    {
      "code": "import \"github.com/optimizely/go-sdk/optimizely/client\"\n\noptimizelyFactory := &client.OptimizelyFactory{\n          SDKKey: \"[SDK_KEY_HERE]\",\n          Datafile: []byte(\"DATAFILE_JSON_STRING_HERE\")\n}\n\n// Instantiates a static client (no datafile polling)\noptlyClient, err := optimizelyFactory.StaticClient()\n\n// Instantiates a client that syncs the datafile in the background\noptlyClient, err := optimizelyFactory.Client()\n\n",
      "language": "go"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": ""
}
[/block]