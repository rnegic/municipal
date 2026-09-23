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
	s = strings.TrimPrefix(s, prefix)
	id, err := strconv.ParseInt(s, 10, 64)
	return id, err == nil && id > 0
}
