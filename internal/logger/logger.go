package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type (
	ContextKey string
)

var (
	logKey ContextKey = "logger"

	gcpMetadataServerURL = "http://metadata.google.internal/computeMetadata/v1/project/project-id"
	gcpProjectID         = ""
	gcpSlogMapping       = map[string]string{
		"level": "severity",
		"msg":   "message",
	}
)

func isGCP() bool {
	return os.Getenv("K_SERVICE") != ""
}

func NewLogger() *slog.Logger {
	opts := &slog.HandlerOptions{}

	if isGCP() {
		opts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
			if v, ok := gcpSlogMapping[a.Key]; ok {
				return slog.Attr{Key: v, Value: a.Value}
			}
			return a
		}
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

func WithLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := NewLogger()
		logger = logger.With("method", r.Method, "path", r.URL.Path)

		if isGCP() {
			traceparent := r.Header.Get("Traceparent")
			parsedTraceparent := strings.Split(traceparent, "-")

			traceID := parsedTraceparent[1]
			traceResource := fmt.Sprintf("projects/%s/traces/%s", gcpProjectID, traceID)
			spanID := parsedTraceparent[2]

			logger = logger.With("logging.googleapis.com/trace", traceResource)
			logger = logger.With("logging.googleapis.com/spanId", spanID)
		}

		ctx := context.WithValue(r.Context(), logKey, logger)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(logKey).(*slog.Logger); ok {
		return logger
	}

	return NewLogger()
}

func init() {
	if !isGCP() {
		return
	}

	req, err := http.NewRequest("GET", gcpMetadataServerURL, nil)
	if err != nil {
		return
	}

	req.Header.Add("Metadata-Flavor", "Google")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		FromContext(context.Background()).Error("Failed to get GCP project ID", "cause", err)
		return
	}
	defer resp.Body.Close()

	projectID, err := io.ReadAll(resp.Body)
	if err != nil {
		FromContext(context.Background()).Error("Failed to read GCP project ID", "cause", err)
		return
	}

	gcpProjectID = string(projectID)
}
