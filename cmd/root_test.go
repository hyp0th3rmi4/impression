package cmd

import (
	"context"
	"testing"

	"github.com/hyp0th3rmi4/impression/runtime"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitImpression(t *testing.T) {

	testCases := []struct {
		Name        string
		Topic       string
		Project     string
		OutputPath  string
		MaxMessages int
		Context     context.Context
		Expected    *runtime.Impression
	}{
		{
			Name:        "Empty Topic",
			Topic:       "",
			Project:     "project-id",
			MaxMessages: 10,
			OutputPath:  "",
			Context:     context.Background(),
			Expected:    nil,
		}, {
			Name:        "Empty Project",
			Topic:       "topic-id",
			Project:     "",
			MaxMessages: 10,
			OutputPath:  "",
			Context:     context.Background(),
			Expected:    nil,
		}, {
			Name:        "Nil Context",
			Topic:       "topic-id",
			Project:     "",
			MaxMessages: 10,
			OutputPath:  "",
			Context:     nil,
			Expected:    nil,
		}, {
			Name:        "Standard Output",
			Topic:       "topic-id",
			Project:     "project-id",
			MaxMessages: 0,
			OutputPath:  "",
			Context:     context.Background(),
			Expected: &runtime.Impression{
				Topic:   "topic-id",
				Project: "project-id",
				Context: context.Background(),
				Handler: &runtime.StdOutHandler{},
			},
		}, {
			Name:        "File Output",
			Topic:       "topic-id",
			Project:     "project-id",
			MaxMessages: 0,
			OutputPath:  "/path/to/file",
			Context:     context.Background(),
			Expected: &runtime.Impression{
				Topic:   "topic-id",
				Project: "project-id",
				Context: context.Background(),
				Handler: &runtime.FileHandler{
					OutputPath: "/path/to/file",
				},
			},
		}, {
			Name:        "Bounded Messages",
			Topic:       "topic-id",
			Project:     "project-id",
			MaxMessages: 10,
			OutputPath:  "/path/to/file",
			Context:     context.Background(),
			Expected: &runtime.Impression{
				Topic:   "topic-id",
				Project: "project-id",
				Context: context.Background(),
				Handler: &runtime.FileHandler{
					OutputPath: "/path/to/file",
				},
			},
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.Name, func(tt *testing.T) {

			actual, err := initImpression(testCase.Context, testCase.Project, testCase.Topic, testCase.OutputPath, testCase.MaxMessages)
			if testCase.Expected == nil {
				assert.NotNil(tt, err)
				assert.Nil(tt, actual)
			} else {
				assert.Nil(tt, err)
				require.NotNil(tt, actual)

				assert.Equal(tt, testCase.Expected.Topic, actual.Topic)
				assert.Equal(tt, testCase.Expected.Project, actual.Project)
				assert.Equal(tt, testCase.Expected.Context, actual.Context)
				assert.Equal(tt, testCase.Expected.Handler, actual.Handler)
				require.NotNil(tt, actual.Guard)

				if testCase.MaxMessages > 0 {
					// we test the guard as we it should be functioning
					// because we can't use equality tests in functions.
					for i := 0; i < testCase.MaxMessages-1; i++ {
						assert.False(tt, actual.Guard(nil, nil))
					}
					assert.True(tt, actual.Guard(nil, nil))
				}

			}
		})
	}
}

func TestNewExecuteCommand(t *testing.T) {

	actual := NewExecuteCommand()
	assert.NotNil(t, actual)

	// the only handle should be RunE
	assert.NotNil(t, actual.RunE)
	assert.Nil(t, actual.Run)
	assert.Nil(t, actual.PreRunE)
	assert.Nil(t, actual.PreRun)
	assert.Nil(t, actual.PostRun)
	assert.Nil(t, actual.PostRunE)
	assert.Nil(t, actual.PersistentPreRunE)
	assert.Nil(t, actual.PersistentPreRun)
	assert.Nil(t, actual.PersistentPostRun)
	assert.Nil(t, actual.PersistentPostRunE)

	assert.Equal(t, "impression", actual.Use)
	assert.Equal(t, "impression --project my-project --topic my-topic --message-limit 10 --output-path ./event-trace", actual.Example)

	flagsTestCases := []struct {
		Name      string
		ShortHand string
		Default   string
		Required  bool
	}{
		{
			Name:      "project",
			ShortHand: "p",
			Default:   "",
			Required:  true,
		}, {
			Name:      "topic",
			ShortHand: "t",
			Default:   "",
			Required:  true,
		}, {
			Name:      "output-path",
			ShortHand: "o",
			Default:   "",
			Required:  false,
		}, {
			Name:      "message-limit",
			ShortHand: "m",
			Default:   "-1",
			Required:  false,
		},
	}

	for _, flagTestCase := range flagsTestCases {

		t.Run(flagTestCase.Name, func(tt *testing.T) {

			actualFlag := actual.Flags().Lookup(flagTestCase.Name)
			assert.NotNil(tt, actualFlag)
			assert.NotEmpty(tt, actualFlag.Usage)
			assert.Equal(tt, flagTestCase.Default, actualFlag.DefValue)

			if flagTestCase.Required {

				require.NotNil(tt, actualFlag.Annotations)
				value, isPresent := actualFlag.Annotations[cobra.BashCompOneRequiredFlag]
				assert.True(tt, isPresent)
				assert.EqualValues(tt, []string{"true"}, value)

			} else {
				assert.Equal(tt, 0, len(actualFlag.Annotations))
			}
		})
	}

}
