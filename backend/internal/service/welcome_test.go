package service

import (
	"context"
	"strings"
	"testing"

	"ukapp/internal/maxclient"
)

func TestHandleUpdates_WelcomesOnBotStarted(t *testing.T) {
	s := testStore(t)
	svc := New(s, nil, nil, &fakeUk{}, "testbot")
	var started, other maxclient.Update
	started.UpdateType, started.User.UserID, started.User.FirstName = "bot_started", 777, "Анна"
	other.UpdateType, other.User.UserID = "message_created", 778
	if err := svc.handleUpdates(context.Background(), []maxclient.Update{started, other}); err != nil {
		t.Fatal(err)
	}
	var target int64
	var text string
	if err := s.DB().QueryRow(`SELECT target_max_user_id, payload->>'text' FROM outbox_message WHERE kind = 'welcome'`).Scan(&target, &text); err != nil {
		t.Fatal(err)
	}
	if target != 777 || !strings.HasPrefix(text, "Здравствуйте, Анна!") || !strings.Contains(text, "https://max.ru/testbot?startapp=welcome") {
		t.Fatalf("target=%d text=%q", target, text)
	}
	if n := outboxCount(t, s, "pending"); n != 1 {
		t.Fatalf("only bot_started must be welcomed, got %d messages", n)
	}
}
