package sqlstorage

import (
	"database/sql"
	"errors"
	"postponer/core"
	"postponer/model"
	"time"
)

type sqlStorage struct {
	Db     *sql.DB
	Logger model.Logger
}

func (m *sqlStorage) SaveNewMessage(message model.Message) error {

	_, err := m.Db.Exec(
		"INSERT INTO postponer_queue(id, queue, body, fires_at, created_at) VALUES ($1, $2, $3, $4, $5)",
		message.ID,
		message.Queue,
		message.Body,
		message.FiresAt.Unix(),
		time.Now().Unix(),
	)

	if err != nil {
		m.Logger.Errorf("Mysql save msg: %s, error: %s", message.ID, err.Error())
		return err
	}

	return nil
}

func (m *sqlStorage) GetNextMessage() (*model.Message, error) {

	row := m.Db.QueryRow(`SELECT "id", "queue", "body", "fires_at" FROM "postponer_queue" ORDER BY "fires_at" LIMIT 1 FOR SHARE SKIP LOCKED`)

	res := model.Message{}
	firesAtUnix := int64(0)
	err := row.Scan(&res.ID, &res.Queue, &res.Body, &firesAtUnix)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrNoMsg
		}

		m.Logger.Errorf("Mysql getNextMsg error: %s\n", err.Error())

		return nil, err
	}

	res.FiresAt = time.Unix(firesAtUnix, 0)

	return &res, nil
}

func (m *sqlStorage) GetMessagesForDispatch(firesAt time.Time, limit int) core.DispatchMessagesTxn {
	tx, err := m.Db.Begin()

	if err != nil {
		m.Logger.Error("Mysql cannot start tx, error: " + err.Error())

		return &DispatchMsgsTxn{msgs: []model.Message{}}
	}

	rows, err := tx.Query(
		`SELECT "id", "queue", "body", "fires_at" FROM "postponer_queue" WHERE "fires_at" < $1 LIMIT $2 FOR UPDATE SKIP LOCKED`,
		firesAt.Unix(),
		limit)

	if err != nil {
		_ = tx.Rollback()
		m.Logger.Error("Cannot getMessages " + err.Error())
		return &DispatchMsgsTxn{msgs: []model.Message{}}
	}

	defer rows.Close()

	result := make([]model.Message, 0)

	for rows.Next() {
		msg := model.Message{}
		firesAtUnix := int64(0)

		if err := rows.Scan(&msg.ID, &msg.Queue, &msg.Body, &firesAtUnix); err != nil {
			_ = tx.Rollback()
			m.Logger.Error("Cannot getMessages scan" + err.Error())
			return &DispatchMsgsTxn{msgs: []model.Message{}}
		}
		msg.FiresAt = time.Unix(firesAtUnix, 0)

		result = append(result, msg)
	}

	return &DispatchMsgsTxn{
		tx:     tx,
		logger: m.Logger,
		msgs:   result,
	}
}

func NewStorage(db *sql.DB, l model.Logger) *sqlStorage {
	return &sqlStorage{
		Db:     db,
		Logger: l,
	}
}
