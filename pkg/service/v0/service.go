package svc

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/owncloud/ocis-wopiserver/pkg/assets"
	"github.com/owncloud/ocis-wopiserver/pkg/config"
	"github.com/owncloud/ocis/ocis-pkg/log"
	ocsm "github.com/owncloud/ocis/ocis-pkg/middleware"
)

// Service defines the extension handlers.
type Service interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// NewService returns a service implementation for Service.
func NewService(opts ...Option) Service {
	options := newOptions(opts...)

	m := chi.NewMux()
	m.Use(options.Middleware...)

	m.Use(ocsm.Static(
		options.Config.HTTP.Root,
		assets.New(
			assets.Logger(options.Logger),
			assets.Config(options.Config),
		),
		options.Config.HTTP.CacheTTL,
	))

	svc := WopiServer{
		serviceID: options.Config.HTTP.Namespace + "." + options.Config.Server.Name,
		logger:    options.Logger,
		config:    options.Config,
		mux:       m,
	}

	m.Route(options.Config.HTTP.Root, func(r chi.Router) {
		r.NotFound(svc.NotFound)
		r.Use(middleware.StripSlashes)
	})

	return svc
}

// WopiServer defines implements the business logic for Service.
type WopiServer struct {
	serviceID string
	logger    log.Logger
	config    *config.Config
	mux       *chi.Mux
}

// ServeHTTP implements the Service interface.
func (p WopiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mux.ServeHTTP(w, r)
}

func (p WopiServer) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
