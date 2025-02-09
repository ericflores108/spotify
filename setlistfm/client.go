package setlistfm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const BaseURL = "https://api.setlist.fm/rest/1.0"

type SetlistFM struct {
	Client *http.Client
	APIKey string
}

func (s *SetlistFM) GetSetlist(ID string) (*Setlist, error) {
	URL := fmt.Sprintf("%s/setlist/%s", BaseURL, ID)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", s.APIKey)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var setlist Setlist
	if err := json.Unmarshal(body, &setlist); err != nil {
		return nil, err
	}

	return &setlist, nil
}
