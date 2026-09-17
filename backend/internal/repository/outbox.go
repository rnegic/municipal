package repository

import (
	"context"
	"encoding/json"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"

	"ukapp/gen/db/ukapp/public/model"
	. "ukapp/gen/db/ukapp/public/table"
)

type OutboxPayload struct {
	Text string `json:"text"`
}

// Enqueue writes one outbox row per recipient inside the caller's transaction (at-least-once).
func (s *Store) Enqueue(ctx context.Context, tx qrm.DB, targetMaxIDs []int64, kind string, p OutboxPayload) error {
	if len(targetMaxIDs) == 0 {
		return nil
	}
	raw, _ := json.Marshal(p)
	stmt := OutboxMessage.INSERT(OutboxMessage.TargetMaxUserID, OutboxMessage.Kind, OutboxMessage.Payload)
	for _, id := range targetMaxIDs {
		stmt = stmt.VALUES(id, kind, string(raw))
	}
	_, err := stmt.ExecContext(ctx, tx)
	return err
}

// FetchPendingOutbox returns up to limit due messages, oldest first.
func (s *Store) FetchPendingOutbox(ctx context.Context, limit int64) ([]model.OutboxMessage, error) {
	var items []model.OutboxMessage
	err := SELECT(OutboxMessage.ID, OutboxMessage.TargetMaxUserID, OutboxMessage.Payload, OutboxMessage.Attempts).
		FROM(OutboxMessage).
		WHERE(OutboxMessage.Status.EQ(String("pending")).AND(OutboxMessage.NextAttemptAt.LT_EQ(NOW()))).
		ORDER_BY(OutboxMessage.ID).LIMIT(limit).
		QueryContext(ctx, s.db, &items)
	return items, err
}

func (s *Store) MarkOutboxSent(ctx context.Context, id int64, attempts int32) error {
	_, err := OutboxMessage.UPDATE(OutboxMessage.Status, OutboxMessage.SentAt, OutboxMessage.Attempts).
		SET("sent", NOW(), attempts).
		WHERE(OutboxMessage.ID.EQ(Int64(id))).ExecContext(ctx, s.db)
	return err
}

// MarkOutboxAttemptFailed bumps attempts and schedules a retry (backoff: attempts² minutes),
// or marks the message failed once maxAttempts is reached.
func (s *Store) MarkOutboxAttemptFailed(ctx context.Context, id int64, prevAttempts int32, maxAttempts int32) error {
	attempts := prevAttempts + 1
	status := "pending"
	if attempts >= maxAttempts {
		status = "failed"
	}
	_, err := OutboxMessage.UPDATE(OutboxMessage.Attempts, OutboxMessage.Status, OutboxMessage.NextAttemptAt).
		SET(attempts, status, NOW().ADD(INTERVALd(time.Duration(attempts*attempts)*time.Minute))).
		WHERE(OutboxMessage.ID.EQ(Int64(id))).ExecContext(ctx, s.db)
	return err
}
