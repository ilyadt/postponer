package stdoutdispatcher

import (
    "fmt"
    "postponer/model"
)

type StdoutDispatcher struct{}

func (d *StdoutDispatcher) Dispatch(message *model.Message) error {
    //nolint:forbidigo
    fmt.Println(message.Queue + ":" + message.Body)

    return nil
}
