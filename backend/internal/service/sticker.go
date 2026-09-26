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

const (
	stickerQRLeft = 94.0
	stickerQRTop  = 198.0
	stickerQRSize = 232.0
)

func qrPath(link string) (string, error) {
	q, err := qrcode.New(link, qrcode.Medium)
	if err != nil {
		return "", err
	}
	q.DisableBorder = true
	bm := q.Bitmap()
	cell := stickerQRSize / float64(len(bm))
	var path strings.Builder
	for y, row := range bm {
		for x, dark := range row {
			if !dark {
				continue
			}
			fmt.Fprintf(&path, "M%.2f %.2fh%.2fv%.2fh-%.2fz",
				stickerQRLeft+float64(x)*cell, stickerQRTop+float64(y)*cell, cell, cell, cell)
		}
	}
	return path.String(), nil
}

const (
	stickerAddressWidth    = 300.0
	stickerAddressBaseline = 534.0
	stickerAddressLead     = 22.0
)

func fitFontSize(runes int, maxSize float64) float64 {
	return min(maxSize, stickerAddressWidth/(0.55*float64(runes+1)))
}

func splitAddress(address string) []string {
	words := strings.Fields(address)
	total := utf8.RuneCountInString(address)
	split, bestDiff := 0, total
	for i := 1; i < len(words); i++ {
		diff := total - 2*utf8.RuneCountInString(strings.Join(words[:i], " "))
		if diff < 0 {
			diff = -diff
		}
		if diff < bestDiff {
			split, bestDiff = i, diff
		}
	}
	if split == 0 {
		return []string{address}
	}
	return []string{strings.Join(words[:split], " "), strings.Join(words[split:], " ")}
}

func addressText(address string) string {
	if size := fitFontSize(utf8.RuneCountInString(address), 18); size >= 13 {
		return fmt.Sprintf(`<text x="210" y="%.0f" font-size="%.1f" font-weight="700" text-anchor="middle" fill="#101828">%s</text>`,
			stickerAddressBaseline, size, html.EscapeString(address))
	}
	lines := splitAddress(address)
	longest := 0
	for _, line := range lines {
		longest = max(longest, utf8.RuneCountInString(line))
	}
	size := max(10, fitFontSize(longest, 16))
	var out strings.Builder
	top := stickerAddressBaseline - stickerAddressLead*float64(len(lines)-1)/2
	for i, line := range lines {
		fmt.Fprintf(&out, `<text x="210" y="%.0f" font-size="%.1f" font-weight="700" text-anchor="middle" fill="#101828">%s</text>`+"\n",
			top+stickerAddressLead*float64(i), size, html.EscapeString(line))
	}
	return strings.TrimRight(out.String(), "\n")
}

func StickerSVG(address, link string) ([]byte, error) {
	path, err := qrPath(link)
	if err != nil {
		return nil, err
	}
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="105mm" height="148mm" viewBox="0 0 420 594" font-family="'Segoe UI', Roboto, Arial, Helvetica, sans-serif">
<defs>
<linearGradient id="brand" x1="0" y1="0" x2="0" y2="1">
<stop offset="0" stop-color="#3f5a9e"/>
<stop offset="1" stop-color="#334d8f"/>
</linearGradient>
<clipPath id="card"><rect x="3" y="3" width="414" height="588" rx="22"/></clipPath>
</defs>
<rect x="3" y="3" width="414" height="588" rx="22" fill="#ffffff" stroke="#dde5f2" stroke-width="2"/>
<g clip-path="url(#card)">
<rect x="3" y="3" width="414" height="149" fill="url(#brand)"/>
<circle cx="386" cy="14" r="66" fill="#ffffff" opacity="0.08"/>
<circle cx="44" cy="142" r="40" fill="#ffffff" opacity="0.06"/>
</g>
<text x="210" y="44" font-size="11" font-weight="600" letter-spacing="2.4" text-anchor="middle" fill="#ffffff" opacity="0.75">ПРИЛОЖЕНИЕ ДОМА В MAX</text>
<text x="210" y="90" font-size="34" font-weight="700" text-anchor="middle" fill="#ffffff">Что-то сломалось?</text>
<text x="210" y="124" font-size="15" text-anchor="middle" fill="#ffffff" opacity="0.88">Сообщите в УК за минуту — без звонков</text>
<path d="M62 196V178a12 12 0 0 1 12-12h18M324 166h18a12 12 0 0 1 12 12v18M62 432v18a12 12 0 0 0 12 12h18M324 462h18a12 12 0 0 0 12-12v-18" fill="none" stroke="#3f5a9e" stroke-width="4" stroke-linecap="round"/>
<path d="%s" fill="#101828" shape-rendering="crispEdges"/>
<text x="210" y="486" font-size="14" font-weight="600" text-anchor="middle" fill="#475467">Наведите камеру телефона на код</text>
<rect x="40" y="498" width="340" height="58" rx="16" fill="#f1f5fb"/>
%s
<text x="210" y="578" font-size="11.5" text-anchor="middle" fill="#667085">Заявку увидят соседи и УК, о ремонте придёт уведомление</text>
</svg>`, path, addressText(address))
	return []byte(svg), nil
}
