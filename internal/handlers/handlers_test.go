package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMetricHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	type args struct {
		metricType string
	}
	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name   string
		method string
		url    string
		want   want
		args   args
	}{
		{
			name:   "casual gauge",
			method: "POST",
			url:    "/update/gauge/asd/123",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
			args: args{
				metricType: "gauge",
			},
		},
		{
			name:   "casual counter (tba)",
			method: "POST",
			url:    "/update/counter/asd/123",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
			args: args{
				metricType: "counter",
			},
		},
		{
			name:   "wrong type",
			method: "POST",
			url:    "/update/gaug/asd/123",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
			args: args{
				metricType: "gauge",
			},
		},
		{
			name:   "no name",
			method: "POST",
			url:    "/update/gauge/123",
			want: want{
				statusCode:  http.StatusNotFound,
				contentType: "text/plain",
			},
			args: args{
				metricType: "gauge",
			},
		},
		{
			name:   "wrong value",
			method: "POST",
			url:    "/update/gauge/asd/asd",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
			args: args{
				metricType: "gauge",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			r.POST("/update/:metricType/:metricName/:metricValue", URLUpdate())
			request := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			result.Body.Close()
		})
	}
}

func TestJSONValue(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		body        string
	}

	tests := []struct {
		name        string
		metric      string
		sendBody    string
		getBody     string
		want        want
		description string
	}{
		{
			name:   "retrieve gauge metric",
			metric: "gauge",
			sendBody: `{
				"id": "cpu_load",
				"type": "gauge",
				"value": 42.5
			}`,
			getBody: `{
				"id": "cpu_load",
				"type": "gauge"
			}`,
			want: want{
				contentType: "application/json; charset=utf-8",
				statusCode:  200,
				body:        `{"id":"cpu_load","type":"gauge","value":42.5}`,
			},
			description: "Store and retrieve gauge metric",
		},
		{
			name:   "retrieve counter metric",
			metric: "counter",
			sendBody: `{
				"id": "requests",
				"type": "counter",
				"delta": 10
			}`,
			getBody: `{
				"id": "requests",
				"type": "counter"
			}`,
			want: want{
				contentType: "application/json; charset=utf-8",
				statusCode:  200,
				body:        `{"id":"requests","type":"counter","delta":10}`,
			},
			description: "Store and retrieve counter metric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			r.POST("/update", JSONUpdate())
			r.POST("/value", JSONValue())

			if tt.sendBody != "" {
				sendRequest := httptest.NewRequest("POST", "/update", bytes.NewBuffer([]byte(tt.sendBody)))
				sendRequest.Header.Set("Content-Type", "application/json")

				sendRecorder := httptest.NewRecorder()
				r.ServeHTTP(sendRecorder, sendRequest)

				assert.Equal(t, 200, sendRecorder.Code, "Failed to store metric")
			}

			request := httptest.NewRequest("POST", "/value", bytes.NewBuffer([]byte(tt.getBody)))
			request.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			bodyBytes, _ := io.ReadAll(result.Body)
			bodyString := string(bodyBytes)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Contains(t, result.Header.Get("Content-Type"), "application/json")
			assert.JSONEq(t, tt.want.body, bodyString, "Response body does not match expected")
		})
	}
}
