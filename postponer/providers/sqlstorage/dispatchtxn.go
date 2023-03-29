package sqlstorage

import (
    "database/sql"
    "postponer/model"
)

type DispatchMsgsTxn struct {
    tx     *sql.Tx
    msgs   []*model.Message
    logger model.Logger
}

func (txm *DispatchMsgsTxn) Messages() []*model.Message {
    return txm.msgs
}

func (txm *DispatchMsgsTxn) DeleteMsg(messageID string) {
    // dummy transaction
    if txm.tx == nil {
        return
    }

    _, err := txm.tx.Exec(`DELETE FROM "postponer_queue" WHERE "id" = $1`, messageID)

    if err != nil {
        txm.logger.Errorf("Sql delete msg: %s, error: %s", messageID, err.Error())
    }
}

// TODO: error trace
func (txm *DispatchMsgsTxn) Commit() {
    // dummy transaction
    if txm.tx == nil {
        return
    }

    if err := txm.tx.Commit(); err != nil {
        txm.logger.Errorf("Sql commit error: %s", err.Error())
    }
}
