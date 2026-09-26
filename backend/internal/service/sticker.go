package service

import (
	"context"
	"fmt"
	"html"
	"strings"
	"unicode/utf8"

	"github.com/skip2/go-qrcode"

	"ukapp/internal/domain"

	"ukapp/gen/db/ukapp/public/model"
)

func (s *Service) HouseLink(houseID int64) string {
	return fmt.Sprintf("https://max.ru/%s?startapp=h_%d", s.botName, houseID)
}

func (s *Service) BindHouseFromSticker(ctx context.Context, u model.AppUser, houseID int64) (model.AppUser, error) {
	if u.Role != domain.RoleResident || u.HouseID != nil {
		return u, nil
	}
	bound, err := s.repo.SetUserHouse(ctx, u.ID, houseID)
	if err != nil || !bound {
		return u, err
	}
	u.HouseID = &houseID
	return u, nil
}

func (s *Service) HouseSticker(ctx context.Context, houseID int64) ([]byte, error) {
	h, err := s.repo.FindHouse(ctx, houseID)
	if err != nil {
		return nil, err
	}
	return StickerSVG(h.AddressRaw, s.HouseLink(houseID))
}

func StickerSVG(address, link string) ([]byte, error) {
	q, err := qrcode.New(link, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	q.DisableBorder = true
	bm := q.Bitmap()
	const size, top = 300.0, 150.0
	cell := size / float64(len(bm))
	var path strings.Builder
	for y, row := range bm {
		for x, dark := range row {
			if dark {
				fmt.Fprintf(&path, "M%.2f %.2fh%.2fv%.2fh-%.2fz", 60+float64(x)*cell, top+float64(y)*cell, cell, cell, cell)
			}
		}
	}
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="105mm" height="148mm" viewBox="0 0 420 594" font-family="Arial, Helvetica, sans-serif">
<rect width="420" height="594" rx="18" fill="#fff" stroke="#1f6feb" stroke-width="6"/>
<text x="210" y="70" font-size="30" font-weight="700" text-anchor="middle" fill="#111">Что-то сломалось?</text>
<text x="210" y="110" font-size="18" text-anchor="middle" fill="#333">Наведите камеру — сообщите в УК за минуту</text>
<path d="%s" fill="#111"/>
<text x="210" y="498" font-size="%.1f" text-anchor="middle" fill="#111">%s</text>
<text x="210" y="535" font-size="15" text-anchor="middle" fill="#555">Приложение дома в MAX: заявки видят соседи и УК,</text>
<text x="210" y="557" font-size="15" text-anchor="middle" fill="#555">а о ремонте придёт уведомление</text>
</svg>`, path.String(), min(16, 380/(0.55*float64(utf8.RuneCountInString(address)+1))), html.EscapeString(address))
	return []byte(svg), nil
}
