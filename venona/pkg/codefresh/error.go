package codefresh

import "fmt"

type (
	// Error is an error that may be thrown from Codefresh API
	Error struct {
		Message       string
		APIStatusCode int
	}
)

func (c Error) Error() string {
	return fmt.Sprintf("HTTP request to Codefresh API rejected. Status-Code: %d. Message: %s", c.APIStatusCode, c.Message)
}
