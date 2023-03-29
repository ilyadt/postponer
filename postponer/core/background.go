package core

import (
    "context"
    "errors"
    "postponer/model"
    "sync/atomic"
    "time"
)

type Background struct {
    Storage     Storage
    Dispatcher  Dispatcher
    ReloadChan  chan struct{}
    NextMsgUnix atomic.Int64
}

func (b *Background) Do(ctx context.Context) {
    for {
        // Init
        b.NextMsgUnix.Store(0)

        txn := b.Storage.GetMessagesForDispatch(time.Now(), 1000)

        for len(txn.Messages()) > 0 {
            for _, msg := range txn.Messages() {
                err := b.Dispatcher.Dispatch(msg)

                // Protect from not dispatching messages
                if err != nil {
                    continue
                }

                txn.DeleteMsg(msg.ID)
            }

            txn.Commit()

            txn = b.Storage.GetMessagesForDispatch(time.Now(), 1000)
        }

        txn.Commit()

        nextMsg, err := b.Storage.GetNextMessage()
        if err != nil && !errors.Is(err, ErrNoMsg) { // Unexpected error
            select {
            case <-time.After(1 * time.Second): // Ожидание, что база оживет
                continue
            case <-ctx.Done():
                return
            }
        }

        nextMsgTimer := &time.Timer{C: make(chan time.Time)} // infinite timer
        if nextMsg != nil {
            nextMsgTimer = time.NewTimer(time.Until(nextMsg.FiresAt))
            b.NextMsgUnix.Store(nextMsg.FiresAt.Unix())
        }

        // TODO: defer nextMsgTimer.Close()
        select {
        case <-nextMsgTimer.C: // Время до следующего события в базе
        case <-time.After(2 * time.Minute): // Релоад по таймеру, на случай скейлинга
        case <-b.ReloadChan: // Релоад по событию
            continue
        case <-ctx.Done():
            return // exit start function
        }
    }
}

func (b *Background) Add(msg *model.Message) {
    nextFiresAt := b.NextMsgUnix.Load()

    if msg.FiresAt.Unix() < nextFiresAt || nextFiresAt == 0 {
        // If service is already waiting for reload, cleaning it
        select {
        case <-b.ReloadChan:
        default:
        }

        b.ReloadChan <- struct{}{}
    }
}

func NewBackgroundService(s Storage, d Dispatcher) *Background {
    return &Background{
        Storage:    s,
        Dispatcher: d,
        ReloadChan: make(chan struct{}),
    }
}
