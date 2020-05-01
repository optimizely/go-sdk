---
title: "Customize error handler"
slug: "customize-error-handler-go"
hidden: true
createdAt: "2019-09-12T13:58:25.299Z"
updatedAt: "2019-09-12T14:09:26.510Z"
---
[block:callout]
{
  "type": "danger",
  "title": "Update for Go SDK",
  "body": "This content is copied from C# and needs to be updated for Go"
}
[/block]
You can provide your own custom **error handler** logic to standardize across your production environment. 

This error handler is called when SDK is not executed as expected, it may be because of arguments provided to the SDK or running in an environment where network or any other disruptions occur.

See the code example below. If the error handler is not overridden, a no-op error handler is used by default.
[block:code]
{
  "codes": [
    {
      "code": "using System;\nusing OptimizelySDK.ErrorHandler;\n\n/**\n * Creates a CustomErrorHandler and calls HandleError when exception is raised by the SDK. \n **/\n/** CustomErrorHandler should be inherited by IErrorHandler, namespace of OptimizelySDK.ErrorHandler.\n **/\npublic class CustomErrorHandler : IErrorHandler\n{\n    /// <summary>\n    /// Handle exceptions when raised by the SDK.\n    /// </summary>\n    /// <param name=\"exception\">object of Exception raised by the SDK.</param>\n    public void HandleError(Exception exception)\n    {\n        throw new NotImplementedException();\n    }\n}\n\n",
      "language": "csharp"
    }
  ]
}
[/block]