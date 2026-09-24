package http

import (
	"strconv"
	"strings"
)

func formatHouseID(id int64) string    { return "h_" + strconv.FormatInt(id, 10) }
func formatIncidentID(id int64) string { return "inc_" + strconv.FormatInt(id, 10) }
func formatUserID(id int64) string     { return "u_" + strconv.FormatInt(id, 10) }
func formatPhotoID(id int64) string    { return "ph_" + strconv.FormatInt(id, 10) }

func parseID(prefix, s string) (int64, bool) {
	digits, ok := strings.CutPrefix(s, prefix)
	if !ok || digits == "" || strings.TrimLeft(digits, "0123456789") != "" {
		return 0, false
	}
	id, err := strconv.ParseInt(digits, 10, 64)
	return id, err == nil && id > 0
}
