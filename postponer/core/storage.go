package core

import (
    "postponer/model"
    "time"
)

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
