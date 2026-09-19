package http

import (
	"strconv"
	"strings"
)

// Wire IDs are opaque prefixed strings ("h_123", "inc_42", "u_7"); internally everything is
// still a bigint primary key. Prefixes are cosmetic — parseID tolerates a missing prefix too.

func formatHouseID(id int64) string    { return "h_" + strconv.FormatInt(id, 10) }
func formatIncidentID(id int64) string { return "inc_" + strconv.FormatInt(id, 10) }
func formatUserID(id int64) string     { return "u_" + strconv.FormatInt(id, 10) }
func formatEventID(id int64) string    { return "evt_" + strconv.FormatInt(id, 10) }
func formatPhotoID(id int64) string    { return "ph_" + strconv.FormatInt(id, 10) }

func parseID(prefix, s string) (int64, bool) {
	s = strings.TrimPrefix(s, prefix)
	id, err := strconv.ParseInt(s, 10, 64)
	return id, err == nil && id > 0
}
