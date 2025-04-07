package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paranoiachains/metrics/internal/collector"
	"github.com/paranoiachains/metrics/internal/middleware"
	"github.com/paranoiachains/metrics/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMetricHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	type args struct {
		metricType  string
		metricName  string
		metricValue interface{}
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
				metricType:  "gauge",
				metricName:  "asd",
				metricValue: float64(123),
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
				metricType:  "counter",
				metricName:  "asd",
				metricValue: int64(123),
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
				metricType:  "gauge",
				metricName:  "asd",
				metricValue: float64(123),
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
				metricType:  "gauge",
				metricName:  "",
				metricValue: float64(123),
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
				metricType:  "gauge",
				metricName:  "asd",
				metricValue: "asd",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mocks.NewMockDatabase(ctrl)

			if tt.want.statusCode == http.StatusOK {
				mockStorage.EXPECT().Update(gomock.Any(), tt.args.metricType, tt.args.metricName, gomock.Any())
			}

			r := gin.New()
			r.Use(gin.Recovery(), middleware.LoggerMiddleware(), middleware.GzipMiddleware())

			r.POST("/update/:metricType/:metricName/:metricValue", func(c *gin.Context) {
				if c.Param("metricType") != "gauge" && c.Param("metricType") != "counter" {
					c.String(http.StatusBadRequest, "invalid metric type")
					return
				}
				urlHandle(c, c.Param("metricType"), mockStorage)
			})

			request := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
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
		{
			name:   "pollcount 0",
			metric: "counter",
			sendBody: `{
				"id": "PollCount",
				"type": "counter",
				"delta": 0
			}`,
			getBody: `{
				"id": "PollCount",
				"type": "counter"
			}`,
			want: want{
				contentType: "application/json; charset=utf-8",
				statusCode:  200,
				body:        `{"id":"PollCount","type":"counter","delta":0}`,
			},
			description: "pollcount with zero value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mocks.NewMockDatabase(ctrl)

			if tt.want.statusCode == http.StatusOK {
				var sendMetric collector.Metric
				err := json.Unmarshal([]byte(tt.sendBody), &sendMetric)
				if err != nil {
					t.Fatalf("failed to unmarshal sendBody: %v", err)
				}

				var getMetric collector.Metric
				err = json.Unmarshal([]byte(tt.getBody), &getMetric)
				if err != nil {
					t.Fatalf("failed to unmarshal getBody: %v", err)
				}
				var wantMetric collector.Metric
				err = json.Unmarshal([]byte(tt.want.body), &wantMetric)
				if err != nil {
					t.Fatalf("failed to unmarshal want.body: %v", err)
				}

				if sendMetric.MType == "gauge" {
					mockStorage.EXPECT().
						Update(gomock.Any(), sendMetric.MType, sendMetric.ID, *sendMetric.Value).
						Return(nil)
				} else {
					mockStorage.EXPECT().
						Update(gomock.Any(), sendMetric.MType, sendMetric.ID, *sendMetric.Delta).
						Return(nil)
				}

				mockStorage.EXPECT().
					Return(gomock.Any(), getMetric.MType, getMetric.ID).
					Return(&wantMetric, nil)
			}

			r := gin.New()
			r.Use(gin.Recovery(), middleware.LoggerMiddleware(), middleware.GzipMiddleware())

			r.POST("/update/", func(c *gin.Context) {
				jsonHandle(c, mockStorage)
			})
			r.POST("/value/", func(c *gin.Context) {
				returnValue(c, mockStorage)
			})

			if tt.sendBody != "" {
				sendRequest := httptest.NewRequest("POST", "/update/", bytes.NewBuffer([]byte(tt.sendBody)))
				sendRequest.Header.Set("Content-Type", "application/json")

				sendRecorder := httptest.NewRecorder()
				r.ServeHTTP(sendRecorder, sendRequest)

				assert.Equal(t, 200, sendRecorder.Code, "Failed to store metric")
			}

			request := httptest.NewRequest("POST", "/value/", bytes.NewBuffer([]byte(tt.getBody)))
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
