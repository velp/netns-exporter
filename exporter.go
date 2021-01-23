package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
)

const exporterName = "netns_exporter"

func init() { //nolint:gochecknoinits
	prometheus.MustRegister(version.NewCollector(exporterName))
}

type APIServer struct {
	config *NetnsExporterConfig
	server *http.Server
	logger logrus.FieldLogger
}

func NewAPIServer(config *NetnsExporterConfig, logger *logrus.Logger) (*APIServer, error) {
	apiServer := APIServer{
		config: config,
		logger: logger.WithField("component", "api-server"),
	}
	// Try to register new Prometheus exporter.
	err := prometheus.Register(NewCollector(config, logger))
	if err != nil {
		apiServer.logger.Errorf("Registering netns exporter failed: %s", err)

		return nil, err
	}

	// Configure and start HTTP server.
	httpMux := http.NewServeMux()
	timeout := time.Duration(config.APIServer.RequestTimeout) * time.Second
	address := strings.Join([]string{
		config.APIServer.ServerAddress,
		strconv.Itoa(config.APIServer.ServerPort),
	}, ":")
	apiServer.server = &http.Server{
		Addr:              address,
		Handler:           httpMux,
		ReadHeaderTimeout: timeout,
		WriteTimeout:      timeout,
		IdleTimeout:       timeout,
	}

	httpMux.HandleFunc("/", apiServer.indexPage)
	httpMux.Handle(config.APIServer.TelemetryPath, apiServer.middlewareLogging(promhttp.Handler()))

	return &apiServer, nil
}

// StartAPIServer starts Exporter's HTTP server.
func (s *APIServer) Start() error {
	s.logger.Infof("Starting API server on %s", s.server.Addr)

	return s.server.ListenAndServe()
}

func (s *APIServer) middlewareLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.WithFields(logrus.Fields{
			"addr":   r.RemoteAddr,
			"method": r.Method,
			"agent":  r.UserAgent(),
		}).Debugf("%s", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (s *APIServer) indexPage(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte(`<html>
<head><title>Network nemespace Exporter</title></head>
<body>
<h1>Network nemespace Exporter</h1>
<p><a href='` + s.config.APIServer.TelemetryPath + `'>Metrics</a></p>
</body>
</html>`))
	if err != nil {
		s.logger.Errorf("error handling index page: %s", err)
	}
}
