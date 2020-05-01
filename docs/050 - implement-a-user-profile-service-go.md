---
title: "Implement a user profile service"
slug: "implement-a-user-profile-service-go"
hidden: true
createdAt: "2019-09-12T13:56:27.183Z"
updatedAt: "2019-12-03T20:10:31.031Z"
---
Use a **User Profile Service** to persist information about your users and ensure variation assignments are sticky. The User Profile Service implementation you provide will override Optimizely's default bucketing behavior in cases when an experiment assignment has been saved. 

When implementing in a multi-server or stateless environment, we suggest using this interface with a backend like Cassandra or Redis. You can decide how long you want to keep your sticky bucketing around by configuring these services.

Implementing a User Profile Service is optional and is only necessary if you want to keep variation assignments sticky even when experiment conditions are changed while it is running (for example, audiences, attributes, variation pausing, and traffic distribution). Otherwise, the Go SDK is stateless and relies on deterministic bucketing to return consistent assignments. See [How bucketing works](doc:how-bucketing-works) for more information.
[block:api-header]
{
  "title": "Implement a user profile service"
}
[/block]
Refer to the code samples below to provide your own User Profile Service. It should expose two functions with the following signatures:

  * lookup: Takes a user ID string and returns a user profile matching the schema below.
  * save: Takes a user profile and persists it.

[block:code]
{
  "codes": [
    {
      "code": "import ( \"github.com/optimizely/go-sdk/pkg/decision\" )\n\n// CustomUserProfileService is custom implementation of the UserProfileService interface\ntype CustomUserProfileService struct {\n}\n\n// Lookup is used to retrieve past bucketing decisions for users\nfunc (s *CustomUserProfileService) Lookup(userID string) decision.UserProfile {\n   return decision.UserProfile{}\n}\n\n// Save is used to save bucketing decisions for users\nfunc (s *CustomUserProfileService) Save(userProfile decision.UserProfile) {\n}\n",
      "language": "go"
    }
  ]
}
[/block]
**The UserProfile struct** 
[block:code]
{
  "codes": [
    {
      "code": "type UserProfile struct {\n   ID                  string\n   ExperimentBucketMap map[UserDecisionKey]string\n}\n\n// UserDecisionKey is used to access the saved decisions in a user profile\ntype UserDecisionKey struct {\n   ExperimentID string\n   Field        string\n}\n\n// Sample user profile with a saved variation\nuserProfile := decision.UserProfile{\n\t\tID: \"optly_user_1\",\n\t\tExperimentBucketMap: map[decision.UserDecisionKey]string{\n\t\t\tdecision.UserDecisionKey{ExperimentID: \"experiment_1\", Field: \"variation_id\" \n    }: \"variation_1234\",\n\t},\n}",
      "language": "go"
    }
  ]
}
[/block]
Use `experiment_bucket_map` from the `UserProfile` struct to override the default bucketing behavior and define an alternate experiment variation for a given user. For each experiment that you want to override, add an object to the map. Use the experiment ID as the key and include a variation_id property that specifies the desired variation. If there isn't an entry for an experiment, then the default bucketing behavior persists.

The Go SDK uses the field `variation_id` by default to create a decision key. Create a decision key manually with the method `decision.NewUserDecisionKey`: 

[block:code]
{
  "codes": [
    {
      "code": "decisionKey := decision.NewUserDecisionKey(\"experiment_id\")",
      "language": "go"
    }
  ]
}
[/block]
**Passing a User Profile Service Implementation to the OptimizelyClient:** 
[block:code]
{
  "codes": [
    {
      "code": "userProfileService := new(CustomUserProfileService)\noptimizelyClient, err := factory.Client(\n       client.WithUserProfileService(userProfileService),\n)\n",
      "language": "go"
    }
  ]
}
[/block]