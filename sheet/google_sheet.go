package sheet

import (
	"context"
	"fmt"

	"github.com/devasherr/contest-bot/parser"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const ContestProblemCount = 5

type Service struct {
	sheetSvc      *sheets.Service
	spreadSheetID string
}

func NewService(spreadSheetID string) (*Service, error) {
	srv, err := sheets.NewService(context.TODO(), option.WithCredentialsFile("credentials.json"), option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return nil, err
	}

	return &Service{
		sheetSvc:      srv,
		spreadSheetID: spreadSheetID,
	}, nil
}

func (s *Service) CreateSheetIfNotExists(sheetName string) error {
	resp, err := s.sheetSvc.Spreadsheets.Get(s.spreadSheetID).Do()
	if err != nil {
		return err
	}

	for _, sh := range resp.Sheets {
		if sh.Properties.Title == sheetName {
			// sheet already exists
			return nil
		}
	}

	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: sheetName,
					},
				},
			},
		},
	}

	_, err = s.sheetSvc.Spreadsheets.BatchUpdate(s.spreadSheetID, req).Do()
	return err
}

func (s *Service) WriteContestData(sheetName string, contestants []parser.Contestant, stats map[string]string) error {
	var values [][]interface{}
	headers := []interface{}{"Rank", "Name", "Solved", "Completion rate"}
	values = append(values, headers)

	for _, c := range contestants {
		completionRate := float32(c.Solved) / float32(ContestProblemCount) * 100
		row := []interface{}{c.Rank, c.Name, c.Solved, fmt.Sprintf("%.2f%%", completionRate)}
		values = append(values, row)
	}

	// append data to sheet
	writeRange := sheetName + "!A1"
	_, err := s.sheetSvc.Spreadsheets.Values.Append(s.spreadSheetID, writeRange, &sheets.ValueRange{
		Values: values,
	}).ValueInputOption("RAW").Do()
	if err != nil {
		return err
	}

	// stats below contestants
	statStartRow := len(contestants) + 3
	var statValues [][]interface{}
	statValues = append(statValues, []interface{}{"Statistics"})
	for k, v := range stats {
		statValues = append(statValues, []interface{}{k, v})
	}

	statRange := fmt.Sprintf("%s!A%d", sheetName, statStartRow)
	_, err = s.sheetSvc.Spreadsheets.Values.Update(s.spreadSheetID, statRange, &sheets.ValueRange{
		Values: statValues,
	}).ValueInputOption("RAW").Do()

	return err
}
