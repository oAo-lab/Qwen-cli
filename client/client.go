package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func Client(apiURL, apiKey string, params []byte, callBack func(data []byte)) error {
	reader := bytes.NewReader(params)

	req, err := http.NewRequest("POST", apiURL, reader)
	if err != nil {
		return fmt.Errorf("error creating request: %s", err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading error response: %s, status code: %d", err.Error(), resp.StatusCode)
		}
		return fmt.Errorf("API error: %s, status code: %d", string(bodyBytes), resp.StatusCode)
	}

	// Handle streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				break
			}
			callBack([]byte(data))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %s", err.Error())
	}

	return nil
}
