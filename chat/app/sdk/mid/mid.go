package mid

import (
	"github.com/omer1998/chat-app-go.git/chat/foundation/web"
)

func IsError(dataModel web.Encoder) error {

	err, isError := dataModel.(error)
	if isError {
		return err
	}
	return nil
}
