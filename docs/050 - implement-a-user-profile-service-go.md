---
title: "Implement a user profile service"
excerpt: ""
slug: "implement-a-user-profile-service-go"
hidden: true
createdAt: "2019-09-12T13:56:27.183Z"
updatedAt: "2019-12-03T20:10:31.031Z"
---
Use a **User Profile Service** to persist information about your users and ensure variation assignments are sticky. The User Profile Service implementation you provide will override Optimizely's default bucketing behavior in cases when an experiment assignment has been saved. 

When implementing in a multi-server or stateless environment, we suggest using this interface with a backend like Cassandra or Redis. You can decide how long you want to keep your sticky bucketing around by configuring these services.

Implementing a User Profile Service is optional and is only necessary if you want to keep variation assignments sticky even when experiment conditions are changed while it is running (for example, audiences, attributes, variation pausing, and traffic distribution). Otherwise, the Go SDK is stateless and relies on deterministic bucketing to return consistent assignments. See [How bucketing works](doc:how-bucketing-works) for more information.
### Implement a user profile service
Refer to the code samples below to provide your own User Profile Service. It should expose two functions with the following signatures:

  * lookup: Takes a user ID string and returns a user profile matching the schema below.
  * save: Takes a user profile and persists it.

```go
import ( "github.com/optimizely/go-sdk/pkg/decision" )

// CustomUserProfileService is custom implementation of the UserProfileService interface
type CustomUserProfileService struct {
}

// Lookup is used to retrieve past bucketing decisions for users
func (s *CustomUserProfileService) Lookup(userID string) decision.UserProfile {
   return decision.UserProfile{}
}

// Save is used to save bucketing decisions for users
func (s *CustomUserProfileService) Save(userProfile decision.UserProfile) {
}

```
**The UserProfile struct** 
```go
type UserProfile struct {
   ID                  string
   ExperimentBucketMap map[UserDecisionKey]string
}

// UserDecisionKey is used to access the saved decisions in a user profile
type UserDecisionKey struct {
   ExperimentID string
   Field        string
}

// Sample user profile with a saved variation
userProfile := decision.UserProfile{
		ID: "optly_user_1",
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decision.UserDecisionKey{ExperimentID: "experiment_1", Field: "variation_id" 
    }: "variation_1234",
	},
}
```

Use `experiment_bucket_map` from the `UserProfile` struct to override the default bucketing behavior and define an alternate experiment variation for a given user. For each experiment that you want to override, add an object to the map. Use the experiment ID as the key and include a variation_id property that specifies the desired variation. If there isn't an entry for an experiment, then the default bucketing behavior persists.

The Go SDK uses the field `variation_id` by default to create a decision key. Create a decision key manually with the method `decision.NewUserDecisionKey`: 

```go
decisionKey := decision.NewUserDecisionKey("experiment_id")
```
**Passing a User Profile Service Implementation to the OptimizelyClient:** 
```go
userProfileService := new(CustomUserProfileService)
optimizelyClient, err := factory.Client(
       client.WithUserProfileService(userProfileService),
)

```