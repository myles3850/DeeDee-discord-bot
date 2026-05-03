package googleplatform

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Sheet struct {
	service *sheets.Service
	id      string
}

func NewSheet(ctx context.Context) (*Sheet, error) {
	creds := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if creds == "" {
		return nil, fmt.Errorf("GOOGLE_CREDENTIALS_JSON is not set")
	}
	id := os.Getenv("GOOGLE_SPREADSHEET_ID")
	if id == "" {
		return nil, fmt.Errorf("GOOGLE_SPREADSHEET_ID is not set")
	}
	svc, err := sheets.NewService(ctx, option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(creds)))
	if err != nil {
		return nil, fmt.Errorf("creating sheets service: %w", err)
	}
	return &Sheet{service: svc, id: id}, nil
}

// GetWorksheet returns all values from the named worksheet tab.
func (s *Sheet) GetWorksheet(ctx context.Context, sheetName string) ([][]interface{}, error) {
	resp, err := s.service.Spreadsheets.Values.Get(s.id, sheetName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("reading worksheet %q: %w", sheetName, err)
	}
	return resp.Values, nil
}
