package fetch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type metadata struct {
	DownloadTokens string `json:"downloadTokens"`
}

type boardData struct {
	Cells [9][9]Cell `json:"cells,omitempty"`
	// Regions   [][][]int64   `json:"regions,omitempty"`
	// Cages     []interface{} `json:"cages,omitempty"`
	// Lines     []interface{} `json:"lines,omitempty"`
	// Arrows    []interface{} `json:"arrows,omitempty"`
	// Underlays []interface{} `json:"underlays,omitempty"`
	// Overlays  []interface{} `json:"overlays,omitempty"`
}

type Cell struct {
	Value string `json:"value,omitempty"`
}

const (
	baseUrl = "https://firebasestorage.googleapis.com/v0/b/sudoku-sandbox.appspot.com/o"
)

func FetchSudoku(id string) (boardData, error) {
	metadataUrl := fmt.Sprintf("%s/%s", baseUrl, id)

	resp, err := http.Get(metadataUrl)
	if err != nil {
		return boardData{}, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return boardData{}, fmt.Errorf("fetch metadata: %w", err)
	}

	var m metadata
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return boardData{}, fmt.Errorf("decode metadata: %w", err)
	}

	dataUrl, err := url.Parse(metadataUrl)
	if err != nil {
		return boardData{}, err
	}

	dataQuery := dataUrl.Query()
	dataQuery.Set("alt", "media")
	dataQuery.Set("token", m.DownloadTokens)
	dataUrl.RawQuery = dataQuery.Encode()

	resp, err = http.Get(dataUrl.String())
	if err != nil {
		return boardData{}, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return boardData{}, fmt.Errorf("fetch board data: %w", err)
	}

	var data boardData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return boardData{}, fmt.Errorf("decode metadata: %w", err)
	}

	return data, nil
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("HTTP %s", resp.Status)
	}
	return nil
}
