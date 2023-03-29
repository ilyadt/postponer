package core

import (
    "fmt"
    "net/http"
    "postponer/model"
    "strconv"
    "time"

    "github.com/google/uuid"
)

type Handler struct {
    Dispatcher Dispatcher
    Storage    Storage
    Background *Background
}

// Ex. /add?queue=my.favorite.queue&body=this_is_message&delay=5
func (h *Handler) Request(res http.ResponseWriter, req *http.Request) {
    queue := req.URL.Query().Get("queue")

    // TODO: check alphanumeric
    if queue == "" {
        res.WriteHeader(http.StatusBadRequest)
        _, _ = res.Write([]byte("queue name is empty"))

        return
    }

    msgBody := req.URL.Query().Get("body")

    // Min length 1, Max Length 256 KB
    if len(msgBody) == 0 || len(msgBody) > 256*1024*8 {
        res.WriteHeader(http.StatusBadRequest)
        _, _ = res.Write([]byte("body must be from 1 byte to 256KB"))

        return
    }

    // Message delay value
    delay, _ := strconv.Atoi(req.URL.Query().Get("delay"))

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

    go h.reloadBackground(msgModel.ID)

    // Returning MessageId to client
    responseBody := fmt.Sprintf(`{"messageId":"%s"}`, msgModel.ID)
    _, _ = res.Write([]byte(responseBody))
}

func (h *Handler) reloadBackground(newMsgID string) {
    nextMsg, err := h.Storage.GetNextMessage()
    // No New Messages | DB error
    if err != nil {
        return
    }

    // Reloading background service if first message in queue is newMsg
    if newMsgID == nextMsg.ID {
        h.Background.Reload()
    }
}

func NewAddHandler(d Dispatcher, s Storage, b *Background) *Handler {
    return &Handler{
        Dispatcher: d,
        Storage:    s,
        Background: b,
    }
}
