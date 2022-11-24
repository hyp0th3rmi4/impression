package runtime

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/pubsub"
)

// Message is a stripped-down version of the pubsub.Message
// struct that is designed to facilitate the serilisation
// of the messages into JSON format.
type Message struct {
	Id          string            `json:"id"`
	PublishTime time.Time         `json:"publishTime"`
	Attributes  map[string]string `json:"attributes"`
	Data        string            `json:"data"`
}

// NopHandler is an implementation of MessageHandler that
// returns a "nop" function when requested for an handler.
type NopHandler struct {
}

// FileHandler is an implementation of MessageHandler that
// returns a function that writes each message in a separate
// file in JSON format in the prescribed output path.
type FileHandler struct {
	OutputPath string
}

// StdOutHandler is an implementation MessageHandler that
// returns a function that dumps each message to the standard
// output.
type StdOutHandler struct {
}

// Handler returns an empty function.
func (nop *NopHandler) Handler() (ReceiveFunction, error) {

	return func(context.Context, *pubsub.Message) {}, nil
}

// Handler ensures that the directory supplied with the
// given FileHandler is created and valid and then it
// returns function that serialises into JSON each message
// passed as argument. One file per message is created and
// named with the unique identifier of the file.
func (handler *FileHandler) Handler() (ReceiveFunction, error) {

	if len(handler.OutputPath) != 0 {

		info, err := os.Stat(handler.OutputPath)
		if err != nil {

			if errors.Is(err, fs.ErrNotExist) {
				err = os.MkdirAll(handler.OutputPath, os.ModeDir)
				if err != nil {
					return nil, err
				}

			} else {
				return nil, fmt.Errorf("output path could not be read (value: %s, error: %v)", handler.OutputPath, err)
			}
		} else if !info.IsDir() {

			return nil, fmt.Errorf("output path exists and it is not a directory (value: %s)", handler.OutputPath)
		}
	} else {
		return nil, errors.New("output path is empty")
	}

	return func(ctx context.Context, message *pubsub.Message) {

		msg := &Message{
			Id:          message.ID,
			PublishTime: message.PublishTime,
			Attributes:  message.Attributes,
			// capturing the byte array to serialise it into a JSON document
			Data: base64.StdEncoding.EncodeToString(message.Data),
		}

		filePath := filepath.Join(handler.OutputPath, fmt.Sprintf("%s.json", msg.Id))
		bytes, _ := json.Marshal(msg)
		_ = os.WriteFile(filePath, bytes, 0644)
	}, nil
}

// Handler returns a function that formats the message for the console
// and writes it to the standard output. Each message is preceded by
// a separator line and followed by two blank lines.
func (handler *StdOutHandler) Handler() (ReceiveFunction, error) {
	return func(ctx context.Context, message *pubsub.Message) {

		fmt.Println("----------------------------")
		fmt.Printf("Id: %s", message.ID)
		fmt.Printf("PublishTime: %s", message.PublishTime)
		fmt.Println("Attributes:")
		for k, v := range message.Attributes {
			fmt.Printf("- %s: %s", k, v)
		}
		fmt.Println("Data:")
		fmt.Println(message.Data)
		fmt.Println()
		fmt.Println()

	}, nil
}
