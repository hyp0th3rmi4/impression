package runtime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/pubsub"
)

// Impression represents the entity controlling
// the main receive and dispatch logic to listen
// and consume messages from a pubsub topic
type Impression struct {
	// The name of the topic to subscribe to
	Topic string
	// The identifier of the project hosting the topic
	Project string
	// The context
	Context context.Context
	// Handler implementation used to consume messages
	// received from the pubsub topic via the subscription
	Handler MessageHandler
	// Guard function that controls the termination
	// of the receiving loop
	Guard GuardFunction
}

// MessageHandler defines the contract for all the handlers
// that can be configured with an Impression instance. The
// contract defines a method that generates a callback function
// that can be used to process the messages received.
type MessageHandler interface {
	// Handler creates a callback used to process the messages
	// received via the subscription, or errors if it cannot
	// produce a valid callback.
	Handler() (ReceiveFunction, error)
}

// ReceiveFunction represents the signature of the callback that
// is expected by the Subscrition.Receive(....) method.
type ReceiveFunction func(context.Context, *pubsub.Message)

// GuardFunction represents the signature of the function that is
// used to check whether the receiving loop needs to be terminated.
// The expectation is that this function returns true if the loop
// needs to be terminated
type GuardFunction func(context.Context, *pubsub.Message) bool

// ImpressionOpyion defines the signature of a function that can be
// used to configure one of the parameter of an Impression instance.
type ImpressionOption func(impr *Impression)

// WithContext returns a function that configures a given instance
// of Impression with the given context.
func WithContext(ctx context.Context) ImpressionOption {
	return func(impr *Impression) {
		impr.Context = ctx
	}
}

// WithProject returns a function that configures a given instance
// of Impression with the given project identifier.
func WithProject(project string) ImpressionOption {
	return func(impression *Impression) {
		impression.Project = project
	}
}

// WithTopic returns a function that configures a given instance
// of Impression with the given topic name.
func WithTopic(topic string) ImpressionOption {
	return func(impression *Impression) {
		impression.Topic = topic
	}
}

// WithHandler returns a function that configures a given instance
// of Impression with the given message handler implementation.
func WithHandler(handler MessageHandler) ImpressionOption {
	return func(impression *Impression) {
		impression.Handler = handler
	}
}

// WithGuard returns a function that configures a given instance
// of Impression with the given guard function.
func WithGuard(guard GuardFunction) ImpressionOption {
	return func(impression *Impression) {
		impression.Guard = guard
	}
}

// ForeverGuard is a dummy implementation that never returns true and
// keeps the receiving loop active forever.
func ForeverGuard(context context.Context, message *pubsub.Message) bool {
	return false
}

// NewImpression creates an instance of the impression message dispatcher and
// configures it according to the given option functions. The implementation
// also checks that a default handler and a default guard are added if they
// have not been supplied with the configuration option function. The method
// returns an error if any of the following conditions are detected:
//
// - the project is an empty string
// - the topic is an empty string
// - the context is set to nil
//
// In all other cases it returns a valid instance of Impression that can be
// used to susbcribe to the specific topic.
func NewImpression(options ...ImpressionOption) (*Impression, error) {

	impression := &Impression{
		Context: context.Background(),
	}
	for _, option := range options {
		option(impression)
	}

	// once we have run all the options we check whether
	// we need to complete the configuration with default
	// values.

	// add the default guard on number of messages.
	if impression.Guard == nil {
		impression.Guard = ForeverGuard
	}

	if impression.Handler == nil {
		impression.Handler = &NopHandler{}
	}

	if len(impression.Project) == 0 {
		return nil, errors.New("[init] project cannot be empty, and must be equal to the id of project containing the topic")
	}

	if len(impression.Topic) == 0 {
		return nil, errors.New("[init] topic name cannot be empty")
	}

	if impression.Context == nil {
		return nil, errors.New("[init] configuration set context to nil")
	}

	return impression, nil
}

// Run subscribes to the specified topic and dispatches the message to the
// configured handler. The implementation initialises a context that can
// cancelled via an operating system interrupt (CTRL+C) or by reaching the
// natural termination defined by the configured guard.
func (impr *Impression) Run() error {

	// step 1 initalise context
	var cancel context.CancelFunc
	impr.Context, cancel = context.WithCancel(impr.Context)
	defer cancel()

	// step 3 initialising the signal channel
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	interruptHandler := func() {
		<-interruptChannel
		cancel()
	}
	go interruptHandler()

	// step 4 initialise subscription
	sub, err := impr.subscribe()
	if err != nil {
		return err
	}
	defer sub.Delete(impr.Context)

	// step 5 listen and dispatch messages
	err = impr.dispatch(sub, cancel)

	return err
}

// subscribe initialises a pubsub client for the specified project and
// sets up a subscription with a pseudo random name to receive the
// messages from the specified topic. The method returns either a valid
// pointer to a subscription or an error explaining what went wrong.
func (impr *Impression) subscribe() (*pubsub.Subscription, error) {

	client, err := pubsub.NewClient(impr.Context, impr.Project)
	if err != nil {
		return nil, fmt.Errorf("[subscribe] - failed to create pubsub client: %w", err)
	}
	hexString, err := randomHexString(5)
	if err != nil {
		return nil, fmt.Errorf("[subscribe] - failed to generate random identifier: %w", err)
	}
	sub, err := client.CreateSubscription(
		impr.Context, fmt.Sprintf("impression-%s", hexString),
		pubsub.SubscriptionConfig{
			Topic: client.Topic(impr.Topic),
		},
	)
	return sub, err
}

// dispatch executes a blocking call to subscription.Receive to process
// the messages with the configured handler. The receiving loop is either
// terminated when the configured guard function returns false of the
// process receives an interrupt. If the configured handler returns an
// error while configuring the handler function the method terminates
// prematurely without entering the receive loop.
func (impr *Impression) dispatch(subscription *pubsub.Subscription, cancel context.CancelFunc) error {

	handler, err := impr.Handler.Handler()
	if err != nil {
		return err
	}

	err = subscription.Receive(impr.Context, func(ctx context.Context, message *pubsub.Message) {
		handler(impr.Context, message)
		message.Ack()

		done := impr.Guard(impr.Context, message)
		if done {
			cancel()
		}
	})

	return err
}

// randomHexString is a function that generates a random string of the given size
// that is used as a suffix for the subscription name to register with the topic.
func randomHexString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
