package optlyplugins

import (
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/decision"
)

type NotificationCallbacks struct {
	decisionService *decision.CompositeService
	client          *client.OptimizelyClient
	listenersCalled interface{}
}

(NotificationCallbacks *n) SubscribeNotification(listeners map[string]int) {

	addNotificationCallback := func (notificationType string)  {
		switch notificationType {
		case KeyDecision:
			// Add deicision callback
			break
		case TrackKey:
			// Add track callback
			break
		}
	}
	if len(listeners) < 1 {
		return
	}
	
	for key, count := range(listeners) {
		for i := 0; i < count; i++ {
			addNotificationCallback(key)
		}
	}
}

(NotificationCallbacks *n) decisionCallback() {

}

(NotificationCallbacks *n) trackCallback() {

}

