package httpapi

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"eventloop/internal/httpapi/handler"
)

func readResponse(resp *http.Response) (result string, err error) {
	scanner := bufio.NewScanner(resp.Body)
	defer resp.Body.Close()
	for i := 0; scanner.Scan() && i < 5; i++ {
		if result != "" {
			result += " "
		}
		result += scanner.Text()
	}
	if err = scanner.Err(); err != nil {
		return "", err
	}
	return result, nil
}

func handleRequest(t *testing.T, resp *http.Response, err error) string {
	if err != nil {
		t.Fatal(err)
	}
	result, errResp := readResponse(resp)
	if errResp != nil {
		t.Log(errResp)
	}
	t.Log(result)
	return result
}

func handleJsonRequest[T any](t *testing.T, resp *http.Response, err error) (result T) {
	if err != nil {
		t.Fatal(err)
	}
	errDecode := json.NewDecoder(resp.Body).Decode(&result)
	if errDecode != nil {
		t.Log(errDecode)
	}
	t.Log(result)
	return
}

func createEvent(t *testing.T, eventName string) (*http.Response, string) {
	requestURL := fmt.Sprintf("http://localhost:8090/events/1/%v", eventName)
	resp, err := http.PostForm(requestURL, url.Values{})
	return resp, handleRequest(t, resp, err)
}

func getEvents(t *testing.T, eventName string) (*http.Response, string) {
	requestURL := fmt.Sprintf("http://localhost:8090/events/%v", eventName)
	resp, err := http.Get(requestURL)
	return resp, handleRequest(t, resp, err)
}

func triggerEvents(t *testing.T, eventName string) string {
	requestURL := fmt.Sprintf("http://localhost:8090/trigger/%v", eventName)
	resp, err := http.PostForm(requestURL, url.Values{})
	return handleRequest(t, resp, err)
}

func toggleEvents(t *testing.T, body *bytes.Buffer) string {
	resp, err := http.Post("http://localhost:8090/toggle/", "text/plain", body)
	return handleRequest(t, resp, err)
}

func removeEvent(t *testing.T, ids []string) string {
	client := &http.Client{}

	encodedJson, errJson := json.Marshal(ids)
	if errJson != nil {
		t.Errorf("can't unmarshal json: %v", errJson)
		return ""
	}
	reader := bytes.NewReader(encodedJson)
	req, err := http.NewRequest("DELETE", "http://localhost:8090/events/", reader)
	if err != nil {
		t.Errorf("error creting request: %v", err)
		return ""
	}

	// Fetch Request
	resp, errDo := client.Do(req)
	if errDo != nil {
		t.Errorf("error fetching request: %v", errDo)
		return errDo.Error()
	}

	// Read Response Body
	result, errResp := readResponse(resp)
	if errResp != nil {
		t.Errorf("error reading response: %v", errResp)
		return errResp.Error()
	}

	return result
}

func subscribeEvents(
	t *testing.T, eventName string, inputJSON struct {
		Listeners []int `json:"listeners"`
		Triggers  []int `json:"triggers"`
	},
) string {
	requestURL := fmt.Sprintf("http://localhost:8090/subscribe/%v", eventName)

	encodedJSON, errJson := json.Marshal(inputJSON)
	if errJson != nil {
		t.Errorf("can't unmarshal json: %v", errJson)
		return errJson.Error()
	}

	reader := bytes.NewReader(encodedJSON)

	resp, errPost := http.Post(requestURL, "application/json", reader)
	result := handleRequest(t, resp, errPost)

	return result
}

func scheduleEvent(t *testing.T, body *bytes.Buffer) (*http.Response, string) {
	resp, err := http.Post("http://localhost:8090/scheduler/1", "text/plain", body)
	return resp, handleRequest(t, resp, err)
}

func scheduleStop(t *testing.T) (resp *http.Response, JSON handler.ScheduleResponse) {
	var (
		err error
	)

	data := bytes.NewBufferString("STOP")
	resp, err = http.Post("http://localhost:8090/scheduler/", "text/plain", data)

	JSON = handleJsonRequest[handler.ScheduleResponse](t, resp, err)

	return
}
