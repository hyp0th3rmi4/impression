package runtime

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNopHandler(t *testing.T) {

	candidate := &NopHandler{}
	actual, err := candidate.Handler()
	assert.Nil(t, err)
	assert.NotNil(t, actual)

	// note the contract defines context and message not nil,
	// but these tests are an attempt do demonstrate that the
	// function does not thing and disregard any parameter
	// being passed.
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

func TestStdOutHandler(t *testing.T) {

	candidate := &StdOutHandler{}

	handler, err := candidate.Handler()
	assert.Nil(t, err)
	require.NotNil(t, handler)

	message := createMockMessage()

	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handler(context.Background(), message)
	w.Close()

	buffer, _ := io.ReadAll(r)
	actual := string(buffer)

	// loose check, we don't guarantee that the format is exactly there but
	// that the information is represented as we think, we may want to revise
	// this test later.
	assert.True(t, strings.Contains(actual, fmt.Sprintf("Id: %s\n", message.ID)))
	assert.True(t, strings.Contains(actual, fmt.Sprintf("PublishTime: %s\n", message.PublishTime)))
	for k, v := range message.Attributes {
		assert.True(t, strings.Contains(actual, fmt.Sprintf("- %s: %s\n", k, v)))
	}
	assert.True(t, strings.Contains(actual, fmt.Sprint(message.Data)))
	os.Stdout = old

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
