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

        // Fetching messages by batches
        limit := 1000
        for {
            if errors.Is(ctx.Err(), context.Canceled) {
                return
            }

            txn, err := b.Storage.GetMessagesForDispatch(time.Now(), limit)

            // Something wrong with database
            if err != nil {
                wait(2 * time.Second)
                continue
            }

            for _, msg := range txn.Messages() {
                err := b.Dispatcher.Dispatch(msg)

                // Protect from not dispatching messages
                if err != nil {
                    continue
                }

                txn.DeleteMsg(msg.ID)
            }

            txn.Commit()

            // If messages have finished, return from cycle
            if len(txn.Messages()) < limit {
                break
            }
        }

        nextMsg, err := b.Storage.GetNextMessage()
        if err != nil {
            wait(2 * time.Second)
            continue
        }

        nextMsgTimer := newInfiniteTimer()
        if nextMsg != nil {
            nextMsgTimer = time.NewTimer(time.Until(nextMsg.FiresAt))
            b.NextMsgUnix.Store(nextMsg.FiresAt.Unix())
        }

        // TODO: defer nextMsgTimer.Close()
        select {
        case <-nextMsgTimer.C: // Время до следующего события в базе
        case <-time.After(30 * time.Second): // Релоад по таймеру, на случай скейлинга
        case <-b.ReloadChan: // Релоад по событию нового сообщения
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

        go func() {
            b.ReloadChan <- struct{}{}
        }()
    }
}

func NewBackgroundService(s Storage, d Dispatcher) *Background {
    return &Background{
        Storage:    s,
        Dispatcher: d,
        ReloadChan: make(chan struct{}),
    }
}

func wait(d time.Duration) {
    timer := time.NewTimer(d)
    defer timer.Stop()

    <-timer.C
}

func newInfiniteTimer() *time.Timer {
    return &time.Timer{C: make(chan time.Time)}
}
