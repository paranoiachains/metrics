package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricHandler(t *testing.T) {
	type args struct {
		metricType string
	}
	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		url  string
		want want
		args args
	}{
		{
			name: "casual gauge",
			url:  "/update/gauge/asd/123",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain",
			},
			args: args{
				metricType: "gauge",
			},
		},
		{
			name: "casual counter (tba)",
			url:  "/update/counter/asd/123",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain",
			},
			args: args{
				metricType: "counter",
			},
		},
		{
			name: "wrong type",
			url:  "/update/gaug/asd/123",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain",
			},
			args: args{
				metricType: "gauge",
			},
		},
		{
			name: "no name",
			url:  "/update/gauge/123",
			want: want{
				statusCode:  http.StatusNotFound,
				contentType: "text/plain",
			},
			args: args{
				metricType: "gauge",
			},
		},
		{
			name: "wrong value",
			url:  "/update/gauge/asd/asd",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain",
			},
			args: args{
				metricType: "gauge",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(MetricHandler(tt.args.metricType))
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

		})
	}
}
