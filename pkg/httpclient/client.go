package httpclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

func PostJSON(url string, payload interface{}, headers map[string]string) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, errors.New("auth service returned error")
	}

	return resp, nil
}
