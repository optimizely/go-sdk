package listeners

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/optimizely/go-sdk/pkg/logging"
)

// DefaultLatency default
var DefaultLatency = flag.Duration("latency", 0*time.Millisecond, "latency offset")

// DefaultNatsCount default
var DefaultNatsCount = flag.Int("natsCount", 1, "number of nats subscribers")

// DefaultNatsURL default
var DefaultNatsURL = flag.String("natsUrl", nats.DefaultURL, "NATS server URL")

// DefaultSubject default
var DefaultSubject = flag.String("subject", "foo", "NATS subject to produce/consume")

// Listener handles control msgs
type Listener interface {
}

// AgentListener handles agent msgs
type AgentListener struct {
	latency   time.Duration
	natsCount int // max size of the queue before flush
	natsURL   string
	subject   string
	sdkKey    string
	logger    logging.OptimizelyLogProducer
}

// ListenerConfig - define config
type ListenerConfig func(al *AgentListener)

// WithLatency sets the latency
func WithLatency(latency time.Duration) ListenerConfig {
	return func(al *AgentListener) {
		al.latency = latency
	}
}

// WithNatsCount sets the natsCount
func WithNatsCount(natsCount int) ListenerConfig {
	return func(al *AgentListener) {
		al.natsCount = natsCount
	}
}

// WithNatsURL sets the natsUrl
func WithNatsURL(natsURL string) ListenerConfig {
	return func(al *AgentListener) {
		al.natsURL = natsURL
	}
}

// WithSubject sets the subject
func WithSubject(subject string) ListenerConfig {
	return func(al *AgentListener) {
		al.subject = subject
	}
}

// NewAgentListener processes msgs
func NewAgentListener(config ...ListenerConfig) *AgentListener {
	p := &AgentListener{}

	for _, opt := range config {
		opt(p)
	}

	p.logger = logging.GetLogger(p.sdkKey, "AgentListener")
	flag.Parse()

	if p.latency == 0 {
		p.latency = *DefaultLatency
	}

	if p.natsCount == 0 {
		p.natsCount = *DefaultNatsCount
	}

	if p.natsURL == "" {
		p.natsURL = *DefaultNatsURL
	}

	if p.subject == "" {
		p.subject = *DefaultSubject
	}

	return p
}

var wg = sync.WaitGroup{}

// Start starts nats listener
func (a *AgentListener) Start(ctx context.Context) {

	a.logger.Info("Agent Listener started")

	// Print out all the parameters
	a.logger.Info("Running sdk agent listener with the following parameters:")
	flag.VisitAll(func(f *flag.Flag) {
		str := fmt.Sprintf("\t%s: %s\n", f.Name, f.Value)
		a.logger.Info(str)
	})

	nc, err := nats.Connect(a.natsURL)
	if err != nil {
		a.logger.Error(err.Error(), err)
		a.logger.Error("failed to connect to nats: %s", a.natsURL)
	}

	a.addSubs(NewNatsConsumer(nc, a.subject), a.natsCount)

	wg.Wait()

	a.logger.Info("Demo complete.")
}

func (a *AgentListener) addSubs(c consumer, count int) {
	for i := 0; i < count; i++ {
		latency := time.Duration(i+1) * a.latency
		a.addSub(c, latency)
	}
}

func (a *AgentListener) addSub(c consumer, offset time.Duration) {
	wg.Add(1)
	go func() {
		c.Subscribe(offset)
	}()
}

type consumer interface {
	Subscribe(duration time.Duration)
	GetName() string
}

type natsConsumer struct {
	nc      *nats.Conn
	subject string
	logger  logging.OptimizelyLogProducer
}

// NewNatsConsumer is a nats listener
func NewNatsConsumer(nc *nats.Conn, subject string) *natsConsumer {
	return &natsConsumer{
		nc:      nc,
		subject: subject,
	}
}

func (c *natsConsumer) GetName() string {
	return "NATS"
}

func (c *natsConsumer) Subscribe(latency time.Duration) {
	_, _ = c.nc.Subscribe(c.subject, func(msg *nats.Msg) {
		time.Sleep(latency)
		c.logger.Info("Received msg!")

		// take action
		c.logger.Info("Listen Process complete!")
		wg.Done()
	})
}
