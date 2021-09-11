package core

import (
	"context"
	"time"
)

type background struct {
	Storage    Storage
	Dispatcher Dispatcher
	StopCtx    context.Context
	ReloadChan chan struct{}
}

func (b *background) Do() {
	for {
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
		if err != nil {
			select {
			case <-time.After(1 * time.Second): // Ожидание, что база оживет
				continue
			case <-b.StopCtx.Done():
				return
			}
		}

		var nextMsgTimer *time.Timer

		if nextMsg == nil {
			nextMsgTimer = &time.Timer{C: make(chan time.Time)} // infinite timer
		} else {
			nextMsgTimer = time.NewTimer(time.Until(nextMsg.FiresAt))
		}

		// TODO: defer nextMsgTimer.Close()
		select {
		case <-nextMsgTimer.C: // Время до следующего события в базе
		case <-time.After(2 * time.Minute): // Дополнительный релоад по таймеру, на случай скейлинга
		case <-b.ReloadChan: // Релоад по событию
			continue
		case <-b.StopCtx.Done():
			return // exit start function
		}
	}
}

func (b *background) Reload() {
	// If service is already waiting for reload, cleaning it
	select {
	case <-b.ReloadChan:
	default:
	}

	b.ReloadChan <- struct{}{}
}

func NewBackgroundService(s Storage, d Dispatcher, sCtx context.Context) *background {
	return &background{
		Storage:    s,
		Dispatcher: d,
		StopCtx:    sCtx,
		ReloadChan: make(chan struct{}),
	}
}
