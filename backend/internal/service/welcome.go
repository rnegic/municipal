package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"ukapp/internal/maxclient"
	"ukapp/internal/repository"
)

func (s *Service) WelcomeText(name string) string {
	hello := "Здравствуйте!"
	if name = strings.TrimSpace(name); name != "" {
		hello = "Здравствуйте, " + name + "!"
	}
	return hello + ` 👋

Я помогаю жителям быстро сообщать управляющей компании о проблемах в доме.

• Нет воды, не работает лифт, грязно во дворе — оформите заявку за минуту, можно с фото.
• Если соседи уже сообщили — присоединитесь к их заявке: чем больше голосов, тем быстрее реакция.
• Я напишу, когда УК возьмёт проблему в работу и когда её решат, а вы подтвердите, что всё починили.

Откройте приложение и укажите свой адрес: ` + fmt.Sprintf("https://max.ru/%s?startapp=welcome", s.botName)
}

func (s *Service) handleUpdates(ctx context.Context, ups []maxclient.Update) error {
	for _, u := range ups {
		if u.UpdateType != "bot_started" || u.User.UserID == 0 {
			continue
		}
		name := u.User.FirstName
		if name == "" {
			name = u.User.Name
		}
		if err := s.repo.Enqueue(ctx, s.repo.DB(), []int64{u.User.UserID}, "welcome", repository.OutboxPayload{Text: s.WelcomeText(name)}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) RunBotUpdates(ctx context.Context) {
	var marker *int64
	for ctx.Err() == nil {
		ups, next, err := s.maxc.Updates(ctx, marker, "bot_started")
		if err != nil {
			slog.Warn("max updates", "err", err)
			select {
			case <-ctx.Done():
			case <-time.After(5 * time.Second):
			}
			continue
		}
		if err := s.handleUpdates(ctx, ups); err != nil {
			slog.Warn("welcome enqueue", "err", err)
			continue
		}
		marker = next
	}
}
