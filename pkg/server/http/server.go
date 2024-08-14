package http

import (
	"net/http"
	"waitq/api/pkg/service"
)

type Server struct {
	services   *service.Service
	apiName    string
	apiVersion string
}

func NewServer(services *service.Service, apiName, apiVersion string) *Server {
	return &Server{
		services:   services,
		apiName:    apiName,
		apiVersion: apiVersion,
	}
}

func (s *Server) Serve(port string) error {
	router := s.routes()

	return http.ListenAndServe(port, router)
}
