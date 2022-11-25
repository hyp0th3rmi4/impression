package runtime

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNopHandler(t *testing.T) {

	candidate := &NopHandler{}
	actual, err := candidate.Handler()
	assert.Nil(t, err)
	assert.NotNil(t, actual)

	testCases := []struct {
		Name    string
		Context context.Context
		Message *pubsub.Message
	}{
		{
			Name:    "Context(nil):Message(nil)",
			Context: nil,
			Message: nil,
		}, {
			Name:    "Context(valid):Message(nil)",
			Context: context.Background(),
			Message: nil,
		}, {
			Name:    "Context(nil):Message(valid)",
			Context: nil,
			Message: createMockMessage(),
		}, {
			Name:    "Context(nil):Message(nil)",
			Context: context.Background(),
			Message: createMockMessage(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {

			actual(testCase.Context, testCase.Message)
		})
	}
}

func createMockMessage() *pubsub.Message {

	data := []byte("this is an example")
	attributes := map[string]string{
		"attribute1": "value1",
		"attribute2": "value2",
	}

	return createMockMessageWith(attributes, data)
}

func createMockMessageWith(attributes map[string]string, data []byte) *pubsub.Message {

	one := 1

	return &pubsub.Message{
		ID:              uuid.New().String(),
		Data:            data,
		Attributes:      attributes,
		PublishTime:     time.Now(),
		DeliveryAttempt: &one,
	}
}
