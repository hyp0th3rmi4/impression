package cmd

import (
	"context"

	"github.com/hyp0th3rmi4/impression/runtime"
	"github.com/spf13/cobra"
)

// initImpression creates an instance of Impression and configures it with the supplied parameters.
// It assigns the project, topic, and context properties, and based on the value of outputPath it
// supplies a essageHandler implementation that writes to the standard output (outputPath is empty)
// or that will write messages as separate JSON files under the supplied path. The method also does
// configure a guard function that terminates the receiving loop once a maximum number of messages
// are retrieved or loops forever if the specified number of messages is zero or negative.
func initImpression(ctx context.Context, project string, topic string, outputPath string, messageLimit int) (*runtime.Impression, error) {

	var handler runtime.MessageHandler
	if len(outputPath) > 0 {
		handler = &runtime.FileHandler{
			OutputPath: outputPath,
		}
	} else {
		handler = &runtime.StdOutHandler{}
	}

	guard := runtime.MessageLimitGuard(messageLimit)

	// it is impossible here for the configuration
	// to be nil, therefore we don't need to check
	impression, err := runtime.NewImpression(
		runtime.WithProject(project),
		runtime.WithTopic(topic),
		runtime.WithContext(ctx),
		runtime.WithHandler(handler),
		runtime.WithGuard(guard),
	)
	if err != nil {
		return nil, err
	}

	return impression, nil
}

// NewExecuteCommand generates a cobra command that based on the command line parameters
// intialises an instance of Impression and then runns it. The method returns an error if
// any initialisation or error during the execution occurs.
func NewExecuteCommand() *cobra.Command {

	var outputPath string
	var messageLimit int
	var topic string
	var project string

	cmd := &cobra.Command{
		Use:     "impression",
		Short:   "Listens to a specified pubsub topic and dumps the messages published.",
		Example: "impression --project my-project --topic my-topic --message-limit 10 --output-path ./event-trace",
		RunE: func(cmd *cobra.Command, args []string) error {

			impression, err := initImpression(cmd.Context(), project, topic, outputPath, messageLimit)
			if err != nil {
				return err
			}

			return impression.Run()
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "Project identifier of the project where the pubsub topic is hosted.")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "Name of the topic to listen to.")
	cmd.Flags().StringVarP(&outputPath, "output-path", "o", "", "Path to the output directory that will be used to store messages.")
	cmd.Flags().IntVarP(&messageLimit, "message-limit", "m", -1, "Maximum number of messages to listen (if < 0 is unlimited).")

	cmd.MarkFlagRequired("topic")
	cmd.MarkFlagRequired("project")

	return cmd
}
