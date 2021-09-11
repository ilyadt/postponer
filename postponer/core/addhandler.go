package core

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"postponer/model"
	"strconv"
	"time"
)

type handler struct {
	Dispatcher Dispatcher
	Storage    Storage
	Background *background
}

// Ex. /add?queue=my.favorite.queue&body=this_is_message&delay=5
func (h *handler) Request(res http.ResponseWriter, req *http.Request) {

	var queue string
	queueQueryParams, exists := req.URL.Query()["queue"]
	if !exists {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte("queue is mandatory"))

		return
	}
	queue = queueQueryParams[0]

	// TODO: check alphanumeric
	if len(queue) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte("queue name is empty"))

		return
	}

	var msgBody string

	bodyQueryParams, exists := req.URL.Query()["body"]
	if !exists {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte("body param is mandatory"))

		return
	}

	msgBody = bodyQueryParams[0]

	// Min length 1, Max Length 256 KB
	if len(msgBody) == 0 || len(msgBody) > 256*1024*8 {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte("body must be from 1 byte to 256KB"))

		return
	}

	var delay int // default timeout value
	delayQueryParams, exists := req.URL.Query()["delay"]
	if !exists {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte("delay is mandatory"))

		return
	}

	var err error
	delay, err = strconv.Atoi(delayQueryParams[0])

	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte("delay must be integer"))
		return
	}

	// Constraint for message delay
	if delay > 7*24*60*60 {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte("delay must be less than 7 days"))
		return
	}

	// ---- validated ----

	firesAt := time.Now().Add(time.Duration(delay) * time.Second)

	msgModel := model.Message{
		ID:      uuid.New().String(),
		Queue:   queue,
		Body:    msgBody,
		FiresAt: firesAt,
	}

	if delay == 0 {
		if err := h.Dispatcher.Dispatch(msgModel); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
		}

		responseBody := fmt.Sprintf(`{"messageId":"%s"}`, msgModel.ID)
		_, _ = res.Write([]byte(responseBody))

		return
	}

	if err := h.Storage.SaveNewMessage(msgModel); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	go h.reloadBackground(msgModel.FiresAt)

	// Returning MessageId to client
	responseBody := fmt.Sprintf(`{"messageId":"%s"}`, msgModel.ID)
	_, _ = res.Write([]byte(responseBody))
}

func (h *handler) reloadBackground(newMsgFiresAt time.Time) {
	nextMsg, err := h.Storage.GetNextMessage()

	if err != nil {
		return
	}

	if nextMsg == nil || newMsgFiresAt.Before(nextMsg.FiresAt) {
		// Reloading background service
		h.Background.Reload()
	}
}

func NewAddHandler(d Dispatcher, s Storage, b *background) *handler {
	return &handler{
		Dispatcher: d,
		Storage:    s,
		Background: b,
	}
}
