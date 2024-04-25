package responder

import (
	"fmt"
	"net/http"

	"github.com/ptflp/godecoder"

	"go.uber.org/zap"
)

type Responder interface {
	OutputJSON(w http.ResponseWriter, responseData interface{})

	ErrorUserNotFound(w http.ResponseWriter)
	ErrorUserConflict(w http.ResponseWriter)
	ErrorNotAllowed(w http.ResponseWriter)
	ErrorBadRequest(w http.ResponseWriter, err error)
	ErrorForbidden(w http.ResponseWriter)
	ErrorInternal(w http.ResponseWriter, err error)
}

type Respond struct {
	log *zap.Logger
	godecoder.Decoder
}

func NewResponder(decoder godecoder.Decoder, logger *zap.Logger) Responder {
	return &Respond{log: logger, Decoder: decoder}
}

func (r *Respond) OutputJSON(w http.ResponseWriter, responseData interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	_ = r.Encode(w, responseData)
	//if err := r.Encode(w, responseData); err != nil {
	//	r.log.Error("responder json encode error", zap.Error(err))
	//}
}

func (r *Respond) ErrorBadRequest(w http.ResponseWriter, err1 error) {
	r.log.Info("http response bad request status code", zap.Error(err1))
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_ = r.Encode(w, ErrorResponse{
		Message: fmt.Sprintf("400 bad request, err: %v", err1),
	})
	//if err := r.Encode(w, ErrorResponse{
	//	Message: fmt.Sprintf("400 bad request, err: %v", err1),
	//}); err != nil {
	//	r.log.Info("response writer error on write", zap.Error(err))
	//}
}

func (r *Respond) ErrorForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_ = r.Encode(w, ErrorResponse{
		Message: "403 Forbidden",
	})
	//if err := r.Encode(w, ErrorResponse{
	//	Message: "403 Forbidden",
	//}); err != nil {
	//	r.log.Error("response writer error on write", zap.Error(err))
	//}
}

func (r *Respond) ErrorInternal(w http.ResponseWriter, err error) {
	r.log.Info("http response internal server error", zap.Error(err))
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_ = r.Encode(w, ErrorResponse{
		Message: "500 Internal Server Error",
	})
	//if err := r.Encode(w, ErrorResponse{
	//	Message: "500 Internal Server Error",
	//}); err != nil {
	//	r.log.Error("response writer error on write", zap.Error(err))
	//}
}

func (r *Respond) ErrorNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusMethodNotAllowed)
	_ = r.Encode(w, ErrorResponse{
		Message: "405 Method not allowed",
	})
	//if err := r.Encode(w, ErrorResponse{
	//	Message: "405 Method not allowed",
	//}); err != nil {
	//	r.log.Error("response writer error on write", zap.Error(err))
	//}
}

func (r *Respond) ErrorUserConflict(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusConflict)
	_ = r.Encode(w, ErrorResponse{
		Message: "409 User already exists",
	})
	//if err := r.Encode(w, ErrorResponse{
	//	Message: "409 User already exists",
	//}); err != nil {
	//	r.log.Error("response writer error on write", zap.Error(err))
	//}
}

func (r *Respond) ErrorUserNotFound(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_ = r.Encode(w, ErrorResponse{
		Message: "404 Login or password is incorrect",
	})
	//if err := r.Encode(w, ErrorResponse{
	//	Message: "404 Login or password is incorrect",
	//}); err != nil {
	//	r.log.Error("response writer error on write", zap.Error(err))
	//}
}
