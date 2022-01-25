package ingest

import "strings"

// ShowNameToID maps a show name to a show ID.
// Lowercase.
var ShowNameToID = map[string]string{
	"fails of the week":  "fail",
	"rampage":            "rampage",
	"headshot":           "headshot",
	"symphony of skills": "sos",
	"sos":                "sos",
	"top plays":          "top",
}

// GetIDFromShowName returns the ID to use for a show name.
func GetIDFromShowName(name string) string {
	sn, snOK := ShowNameToID[strings.ToLower(name)]
	if snOK {
		return sn
	}
	return "top"
}
