package spreadscore

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	sheets "google.golang.org/api/sheets/v4"
)

var ingameTimeRe = regexp.MustCompile("[^a-zA-Z0-9:]")

// Spreadsheet implements a remote scoring spreadsheet.
type Spreadsheet struct {
	// service is the sheets service
	service *sheets.Service
	// spreadsheetID is the spreadsheet ID
	spreadsheetID string
}

// NewSpreadsheet constructs a new spreadsheet api.
func NewSpreadsheet(service *sheets.Service, spreadsheetID string) *Spreadsheet {
	return &Spreadsheet{
		service:       service,
		spreadsheetID: spreadsheetID,
	}
}

// FetchSubmissions fetches the tailing submissions from the sheet.
func (s *Spreadsheet) FetchSubmissions(ctx context.Context, startRow int) ([]*Submission, error) {
	spreadsheetID := s.spreadsheetID
	service := s.service
	readRange := fmt.Sprintf("SUBMISSIONS!A%d:I", startRow+2)
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, err
	}

	var submissions []*Submission
	fixup := strings.TrimSpace
	getVal := func(i, j int) string {
		if j >= len(resp.Values[i]) {
			return ""
		}
		v, _ := resp.Values[i][j].(string) // prevents panic
		return fixup(v)
	}
	getIntVal := func(i, j int) int {
		if j >= len(resp.Values[i]) {
			return 0
		}
		valStr := fmt.Sprintf("%v", resp.Values[i][j])
		valInt, _ := strconv.Atoi(fixup(valStr))
		return valInt
	}
	getInt64Val := func(i, j int) int64 {
		if j >= len(resp.Values[i]) {
			return 0
		}
		valStr := fmt.Sprintf("%v", resp.Values[i][j])
		valInt, _ := strconv.ParseInt(fixup(valStr), 10, 64)
		return valInt
	}
	getFloatVal := func(i, j int) float32 {
		if j >= len(resp.Values[i]) {
			return 0
		}
		valStr := fmt.Sprintf("%v", resp.Values[i][j])
		valFloat, _ := strconv.ParseFloat(valStr, 32)
		return float32(valFloat)
	}
	_ = getFloatVal
	_ = getIntVal
	for i := 0; i < len(resp.Values); i += 1 {
		row := resp.Values[i]
		if row[0] == "" {
			continue
		}
		timestampStr := getVal(i, 0)
		if timestampStr == "" {
			continue
		}
		tt, err := time.Parse("1/2/2006 15:04:05", timestampStr)
		if err != nil {
			return nil, errors.Wrap(err, "parse timestamp")
		}
		submission := &Submission{
			Timestamp: tt,
			Name: fixup(
				getVal(i, 2),
			),
			MatchID:     getInt64Val(i, 1),
			Description: fixup(getVal(i, 3)),
			ShowName:    fixup(getVal(i, 4)),
			Hero:        fixup(getVal(i, 5)),
			Imported:    strings.ToLower(fixup(getVal(i, 7))) == "true",
		}

		// filter to digits and trim
		igt := getVal(i, 6)
		igt = ingameTimeRe.ReplaceAllString(igt, "")
		igt = strings.TrimSpace(igt)
		submission.IngameTime = igt
		submission.MatchTime = 0 // not used anymore

		submissions = append(submissions, submission)
	}

	return submissions, nil
}

// SetSubmissionImported sets the imported cell to a value.
func (s *Spreadsheet) SetSubmissionsImported(rowIndexes []int, imported bool) error {
	if len(rowIndexes) == 0 {
		return nil
	}
	valStr := "TRUE"
	if !imported {
		valStr = "FALSE"
	}
	var req sheets.BatchUpdateValuesRequest
	req.Data = make([]*sheets.ValueRange, 0, len(rowIndexes))
	for _, rowIndex := range rowIndexes {
		updateRange := fmt.Sprintf("SUBMISSIONS!H%d", rowIndex+2)
		req.Data = append(req.Data, &sheets.ValueRange{
			Range:  updateRange,
			Values: [][]interface{}{[]interface{}{valStr}},
		})
	}
	req.ValueInputOption = "RAW"
	_, err := s.service.Spreadsheets.Values.BatchUpdate(s.spreadsheetID, &req).Do()
	return err
}
