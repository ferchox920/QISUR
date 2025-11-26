package logger

import (
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	formatJSON = "json"
	formatText = "text"
)

type config struct {
	level     slog.Level
	writer    io.Writer
	addSource bool
	format    string
}

// Option permite modificar la configuracion del logger.
type Option func(*config)

// WithLevel ajusta el nivel minimo; por defecto toma LOG_LEVEL o info.
func WithLevel(level slog.Level) Option {
	return func(c *config) {
		c.level = level
	}
}

// WithWriter define un writer destino; por defecto stdout.
func WithWriter(w io.Writer) Option {
	return func(c *config) {
		if w != nil {
			c.writer = w
		}
	}
}

// WithoutSource desactiva el origen de archivo/linea en las entradas de log.
func WithoutSource() Option {
	return func(c *config) {
		c.addSource = false
	}
}

// WithTextFormat fuerza formato de texto; por defecto JSON.
func WithTextFormat() Option {
	return func(c *config) {
		c.format = formatText
	}
}

// New crea un slog.Logger estructurado con valores razonables:
// - Salida JSON (usa LOG_FORMAT=text o WithTextFormat para texto plano)
// - Timestamps RFC3339
// - AddSource=true
// - Nivel desde LOG_LEVEL (debug|info|warn|error), por defecto info.
func New(opts ...Option) *slog.Logger {
	cfg := config{
		level:     levelFromEnv(),
		writer:    os.Stdout,
		addSource: true,
		format:    formatFromEnv(),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	handler := buildHandler(cfg)
	return slog.New(handler)
}

func buildHandler(cfg config) slog.Handler {
	options := &slog.HandlerOptions{
		Level:       cfg.level,
		AddSource:   cfg.addSource,
		ReplaceAttr: replaceAttr,
	}

	if cfg.format == formatText {
		return slog.NewTextHandler(cfg.writer, options)
	}
	return slog.NewJSONHandler(cfg.writer, options)
}

func replaceAttr(_ []string, attr slog.Attr) slog.Attr {
	switch attr.Key {
	case slog.TimeKey:
		if attr.Value.Kind() == slog.KindTime {
			attr.Value = slog.StringValue(attr.Value.Time().Format(time.RFC3339))
		}
	case slog.SourceKey:
		if src, ok := attr.Value.Any().(*slog.Source); ok {
			attr.Value = slog.StringValue(src.File + ":" + strconv.Itoa(src.Line))
		}
	}
	return attr
}

func levelFromEnv() slog.Level {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func formatFromEnv() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT"))) {
	case formatText:
		return formatText
	default:
		return formatJSON
	}
}
