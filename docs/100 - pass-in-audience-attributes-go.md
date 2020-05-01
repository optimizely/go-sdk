---
title: "Pass in audience attributes"
slug: "pass-in-audience-attributes-go"
hidden: false
createdAt: "2019-09-12T13:58:32.804Z"
updatedAt: "2019-10-29T23:40:24.261Z"
---
You can pass strings, numbers, Booleans, and null as user attribute values. Attributes are part of the UserContext object. The example below shows how to pass in attributes.
[block:code]
{
  "codes": [
    {
      "code": "import \"github.com/optimizely/go-sdk/pkg/entities\"\n\nattributes := map[string]interface{}{\n        \"DEVICE\": \"iPhone\",\n        \"hey\":    2,\n}\n\nuser := entities.UserContext{\n        ID:         \"userId\",\n        Attributes: attributes,\n}\n\n",
      "language": "go"
    }
  ]
}
[/block]

[block:callout]
{
  "type": "warning",
  "title": "Important",
  "body": "During audience evaluation, note that if you don't pass a valid attribute value for a given audience condition—for example, if you pass a string when the audience condition requires a Boolean, or if you simply forget to pass a value—then that condition will be skipped. The [SDK logs](doc:customize-logger-go) will include warnings when this occurs."
}
[/block]