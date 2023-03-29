package sqlstorage

import (
	"database/sql"
	"errors"
	"postponer/core"
	"postponer/model"
	"time"
)

type SQLStorage struct {
	DB     *sql.DB
	Logger model.Logger
}

func (m *SQLStorage) SaveNewMessage(message *model.Message) error {
	_, err := m.DB.Exec(
		"INSERT INTO postponer_queue(id, queue, body, fires_at, created_at) VALUES ($1, $2, $3, $4, $5)",
		message.ID,
		message.Queue,
		message.Body,
		message.FiresAt.Unix(),
		time.Now().Unix(),
	)

	if err != nil {
		m.Logger.Errorf("save msg: %s, error: %s", message.ID, err.Error())

		return err
	}

	return nil
}

func (m *SQLStorage) GetNextMessage() (*model.Message, error) {
	row := m.DB.QueryRow(`SELECT "id", "queue", "body", "fires_at" FROM "postponer_queue" ORDER BY "fires_at" LIMIT 1 FOR SHARE SKIP LOCKED`)

	res := model.Message{}
	firesAtUnix := int64(0)
	err := row.Scan(&res.ID, &res.Queue, &res.Body, &firesAtUnix)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		m.Logger.Errorf("GetNextMessage error: %s\n", err.Error())

		return nil, err
	}

	res.FiresAt = time.Unix(firesAtUnix, 0)

	return &res, nil
}

func (m *SQLStorage) GetMessagesForDispatch(firesAt time.Time, limit int) core.DispatchMessagesTxn {
	tx, err := m.DB.Begin()

	if err != nil {
		m.Logger.Error("cannot start transaction, error: " + err.Error())

		return &DispatchMsgsTxn{}
	}

	rows, err := tx.Query(
		`SELECT "id", "queue", "body", "fires_at" FROM "postponer_queue" WHERE "fires_at" <= $1 ORDER BY "fires_at" LIMIT $2 FOR UPDATE SKIP LOCKED`,
		firesAt.Unix(),
		limit,
	)

	if err != nil {
		_ = tx.Rollback()
		m.Logger.Error("Cannot getMessages " + err.Error())

		return &DispatchMsgsTxn{}
	}

	defer rows.Close()

	var result []*model.Message

	for rows.Next() {
		msg := model.Message{}
		firesAtUnix := int64(0)

		if err := rows.Scan(&msg.ID, &msg.Queue, &msg.Body, &firesAtUnix); err != nil {
			_ = tx.Rollback()
			m.Logger.Error("Cannot getMessages scan" + err.Error())

			return &DispatchMsgsTxn{}
		}

		msg.FiresAt = time.Unix(firesAtUnix, 0)

		result = append(result, &msg)
	}

	return &DispatchMsgsTxn{
		tx:     tx,
		logger: m.Logger,
		msgs:   result,
	}
}

func NewStorage(db *sql.DB, l model.Logger) *SQLStorage {
	return &SQLStorage{
		DB:     db,
		Logger: l,
	}
}
