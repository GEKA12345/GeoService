package responder

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ptflp/godecoder"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestRespond_OutputJSON(t *testing.T) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := NewResponder(decoder, logger)

	type args struct {
		w            *httptest.ResponseRecorder
		responseData interface{}
	}
	tests := []struct {
		name string
		resp Responder
		args args
		want int
	}{
		{"1", respond, args{httptest.NewRecorder(), SearchResponse{Addresses: []*Address{{Address: "test", Lat: 1.0, Lon: 2.0}}}}, http.StatusOK},
		//{"2", respond, args{httptest.NewRecorder(), 123}, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.resp.OutputJSON(tt.args.w, tt.args.responseData)
			fmt.Println(tt.args.w.Body)
			assert.Equal(t, tt.args.w.Code, tt.want)
		})
	}
}

func TestRespond_ErrorBadRequest(t *testing.T) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := NewResponder(decoder, logger)

	type args struct {
		w   *httptest.ResponseRecorder
		err error
	}
	tests := []struct {
		name string
		resp Responder
		args args
		want int
	}{
		{"1", respond, args{httptest.NewRecorder(), errors.New("test")}, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.resp.ErrorBadRequest(tt.args.w, tt.args.err)
			assert.Equal(t, tt.args.w.Code, tt.want)
		})
	}
}

func TestRespond_ErrorInternal(t *testing.T) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := NewResponder(decoder, logger)

	type args struct {
		w   *httptest.ResponseRecorder
		err error
	}
	tests := []struct {
		name string
		resp Responder
		args args
		want int
	}{
		{"1", respond, args{httptest.NewRecorder(), errors.New("test")}, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.resp.ErrorInternal(tt.args.w, tt.args.err)
			assert.Equal(t, tt.args.w.Code, tt.want)
		})
	}
}

func TestRespond_ErrorUserNotFound(t *testing.T) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := NewResponder(decoder, logger)

	type args struct {
		w *httptest.ResponseRecorder
	}
	tests := []struct {
		name string
		resp Responder
		args args
		want int
	}{
		{"1", respond, args{httptest.NewRecorder()}, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.resp.ErrorUserNotFound(tt.args.w)
			assert.Equal(t, tt.args.w.Code, tt.want)
		})
	}
}

func TestRespond_ErrorUserConflict(t *testing.T) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := NewResponder(decoder, logger)

	type args struct {
		w *httptest.ResponseRecorder
	}
	tests := []struct {
		name string
		resp Responder
		args args
		want int
	}{
		{"1", respond, args{httptest.NewRecorder()}, http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.resp.ErrorUserConflict(tt.args.w)
			assert.Equal(t, tt.args.w.Code, tt.want)
		})
	}
}

func TestRespond_ErrorNotAllowed(t *testing.T) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := NewResponder(decoder, logger)

	type args struct {
		w *httptest.ResponseRecorder
	}
	tests := []struct {
		name string
		resp Responder
		args args
		want int
	}{
		{"1", respond, args{httptest.NewRecorder()}, http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.resp.ErrorNotAllowed(tt.args.w)
			assert.Equal(t, tt.args.w.Code, tt.want)
		})
	}
}

func TestRespond_ErrorForbidden(t *testing.T) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := NewResponder(decoder, logger)

	type args struct {
		w *httptest.ResponseRecorder
	}
	tests := []struct {
		name string
		resp Responder
		args args
		want int
	}{
		{"1", respond, args{httptest.NewRecorder()}, http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.resp.ErrorForbidden(tt.args.w)
			assert.Equal(t, tt.args.w.Code, tt.want)
		})
	}
}
