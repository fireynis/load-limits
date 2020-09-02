package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseLoad(t *testing.T) {
	app := newTestApplication(t)

	ts := httptest.NewTLSServer(app.routes())
	defer ts.Close()

	tests := []struct {
		name       string
		urlPath    string
		payload    string
		wantCode   int
		wantString string
	}{
		{"Valid ID", "/", "{\"id\":\"4\",\"customer_id\":\"1\",\"load_amount\":\"$250.00\",\"time\":\"2000-01-01T00:00:00Z\"}", http.StatusOK, "{\"id\":4,\"customer_id\":1,\"accepted\":false}"},
		{"Valid ID", "/", "{\"id\":\"2\",\"customer_id\":\"2\",\"load_amount\":\"$250.00\",\"time\":\"2000-01-01T00:00:00Z\"}", http.StatusOK, "{\"id\":2,\"customer_id\":2,\"accepted\":false}"},
		{"Valid ID", "/", "{\"id\":\"2\",\"customer_id\":\"3\",\"load_amount\":\"$250.00\",\"time\":\"2000-01-02T00:00:00Z\"}", http.StatusOK, "{\"id\":2,\"customer_id\":3,\"accepted\":false}"},
		{"Valid ID", "/", "{\"id\":\"2\",\"customer_id\":\"4\",\"load_amount\":\"$250.00\",\"time\":\"2000-01-01T00:00:00Z\"}", http.StatusOK, "{\"id\":2,\"customer_id\":4,\"accepted\":true}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := ts.Client().Post(ts.URL+tt.urlPath, "application/json", strings.NewReader(tt.payload))

			if err != nil {
				t.Errorf("Unexepcted error %v", err)
			}

			defer response.Body.Close()

			data, err := ioutil.ReadAll(response.Body)
			dataString := string(data)

			if response.StatusCode != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, response.StatusCode)
			}

			if dataString != tt.wantString {
				t.Errorf("want %s; got %s", tt.wantString, dataString)
			}
		})
	}
}
