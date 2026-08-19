package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/alexg/my-estate/internal/config"
	"github.com/alexg/my-estate/internal/repository/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	store     *postgres.Store
	log       *slog.Logger
	enableDev bool
	origins   []string
}

func New(store *postgres.Store, cfg config.Config, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{
		store:     store,
		log:       log,
		enableDev: cfg.EnableDevEndpoints,
		origins:   cfg.CORSOrigins,
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(s.logRequests)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.health)
		r.Get("/sources", s.listSources)
		r.Get("/properties", s.searchProperties)
		r.Get("/properties/{id}", s.getProperty)
		r.Get("/properties/{id}/history", s.getPropertyHistory)
		r.Post("/dev/collect", s.devCollect)
	})
	return r
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		s.log.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"ms", time.Since(start).Milliseconds(),
			"request_id", middleware.GetReqID(r.Context()),
		)
	})
}
