package core

import (
    "errors"
    "postponer/model"
    "time"
)

var ErrNoMsg = errors.New("no messages in queue")

type DispatchMessagesTxn interface {
    Messages() []*model.Message
    DeleteMsg(messageID string)
    Commit()
}

type Storage interface {
    SaveNewMessage(message *model.Message) error
    GetNextMessage() (*model.Message, error)
    GetMessagesForDispatch(firesAt time.Time, limit int) DispatchMessagesTxn
}
