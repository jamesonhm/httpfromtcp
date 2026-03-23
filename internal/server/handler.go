package server

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

//type HandlerError struct {
//	StatusCode response.StatusCode
//	Message    string
//}
//
//func (h *HandlerError) Write(w io.Writer) error {
//
//	response.WriteStatusLine(w, h.StatusCode)
//	hdrs := response.GetDefaultHeaders(len(h.Message))
//	response.WriteHeaders(w, hdrs)
//	w.Write([]byte(h.Message))
//	return nil
//}

// type Handler func(w io.Writer, req *request.Request) *HandlerError
type Handler func(w *response.Writer, req *request.Request)
