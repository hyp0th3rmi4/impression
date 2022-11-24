package runtime

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithProject(t *testing.T) {

	candidate := &Impression{}
	expected := "project-id-xcgf45d"
	option := WithProject(expected)

	require.NotNil(t, option)
	option(candidate)
	assert.Equal(t, expected, candidate.Project)

}

func TestWithTopic(t *testing.T) {

	candidate := &Impression{}
	expected := "my-topic-id"
	option := WithTopic(expected)

	require.NotNil(t, option)
	option(candidate)
	assert.Equal(t, expected, candidate.Topic)

}

func TestWithContext(t *testing.T) {

	candidate := &Impression{}
	expected := context.Background()
	option := WithContext(expected)

	require.NotNil(t, option)
	option(candidate)
	assert.Equal(t, expected, candidate.Context)

}

func TestWithHandler(t *testing.T) {

	candidate := &Impression{}
	expected := &NopHandler{}
	option := WithHandler(expected)

	require.NotNil(t, option)
	option(candidate)
	assert.Equal(t, expected, candidate.Handler)

}

func TestWithGuard(t *testing.T) {

	state := struct {
		Called bool
	}{
		Called: false,
	}

	candidate := &Impression{}
	expected := func(_ context.Context, _ *pubsub.Message) bool {
		state.Called = true
		return false
	}

	option := WithGuard(expected)
	require.NotNil(t, option)
	option(candidate)
	require.NotNil(t, candidate.Guard)

	assert.False(t, candidate.Guard(nil, nil))
	assert.True(t, state.Called)
}

func TestNewImpression(t *testing.T) {

	testCases := []struct {
		Name     string
		Options  []ImpressionOption
		Expected *Impression
	}{
		{
			Name:     "Empty Options",
			Options:  nil,
			Expected: nil,
		}, {
			Name: "No Project",
			Options: []ImpressionOption{
				WithContext(context.Background()),
				WithTopic("topicId"),
			},
			Expected: nil,
		}, {
			Name: "Empty Project",
			Options: []ImpressionOption{
				WithContext(context.Background()),
				WithProject(""),
				WithTopic("topicId"),
			},
			Expected: nil,
		}, {
			Name: "No Topic",
			Options: []ImpressionOption{
				WithContext(context.Background()),
				WithProject("project-a3dfzc"),
			},
			Expected: nil,
		}, {
			Name: "Empty Topic",
			Options: []ImpressionOption{
				WithContext(context.Background()),
				WithProject("project-a3dfzc"),
				WithTopic(""),
			},
			Expected: nil,
		}, {
			Name: "Nil Context",
			Options: []ImpressionOption{
				WithContext(nil),
				WithProject("project-a3dfzc"),
				WithTopic("topicId"),
			},
			Expected: nil,
		}, {
			Name: "Baseline",
			Options: []ImpressionOption{
				WithTopic("topicId"),
				WithProject("project-a3dfzc"),
			},
			Expected: &Impression{
				Topic:   "topicId",
				Project: "project-a3dfzc",
				Context: context.Background(),
				Handler: &NopHandler{},
				Guard:   ForeverGuard,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {

			var actual *Impression
			var err error
			if len(testCase.Options) == 0 {
				actual, err = NewImpression()
			} else {
				actual, err = NewImpression(testCase.Options...)
			}

			if testCase.Expected == nil {
				assert.NotNil(tt, err)
				assert.Nil(tt, actual)
			} else {
				assert.Nil(tt, err)
				require.NotNil(tt, actual)

				assert.Equal(tt, testCase.Expected.Project, actual.Project)
				assert.Equal(tt, testCase.Expected.Topic, actual.Topic)
				assert.Equal(tt, testCase.Expected.Handler, actual.Handler)
				// we cannot compare functions for equality the best we got
				// is to check that the guard is assigned.
				assert.NotNil(tt, actual.Guard)
				assert.Equal(tt, testCase.Expected.Guard(nil, nil), actual.Guard(nil, nil))
			}
		})
	}
}
