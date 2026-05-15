package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"panzucha/internal/config"
	"sync"
	"time"
)

type Logger struct {
	slog    *slog.Logger
	sender  *LogstashSender
	service string
	env     string
}

type LogstashSender struct {
	client *http.Client
	url    string
	queue  chan LogEntry
	wg     sync.WaitGroup
	closed bool
	mu     sync.Mutex
}

func NewLogstashSender(url string) *LogstashSender {
	ls := &LogstashSender{
		client: &http.Client{Timeout: 5 * time.Second},
		url:    url,
		queue:  make(chan LogEntry, 10000),
	}
	ls.wg.Add(1)
	go ls.processQueue()
	return ls
}

func (ls *LogstashSender) processQueue() {
	defer ls.wg.Done()
	for entry := range ls.queue {
		data, err := json.Marshal(entry)
		if err != nil {
			slog.Error("failed to marshal log entry", "error", err)
			continue
		}
		// Non-blocking send with goroutine pool? For now, simple.
		go func(d []byte) {
			_, err := ls.client.Post(ls.url, "application/json", bytes.NewReader(d))
			if err != nil {
				slog.Error("failed to send log to logstash", "error", err)
			}
		}(data)
	}
}

func (ls *LogstashSender) Send(entry LogEntry) {
	ls.mu.Lock()
	if ls.closed {
		ls.mu.Unlock()
		return
	}
	ls.mu.Unlock()

	select {
	case ls.queue <- entry:
	default:
		slog.Error("log queue full, dropping entry", "category", entry.Category)
	}
}

func (ls *LogstashSender) Close() {
	ls.mu.Lock()
	ls.closed = true
	ls.mu.Unlock()
	close(ls.queue)
	ls.wg.Wait()
}

// New creates a new logger with async sending
func New(cfg *config.Config) *Logger {
	var sender *LogstashSender
	if cfg.LogstashURL != "" {
		sender = NewLogstashSender(cfg.LogstashURL)
		slog.Info("Logstash sender enabled", "url", cfg.LogstashURL)
	} else {
		slog.Warn("Logstash URL not set – logs will not be forwarded to ELK")
	}

	// Structured slog handler for stdout (JSON format for k8s)
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)

	return &Logger{
		slog:    slog.New(jsonHandler),
		sender:  sender,
		service: cfg.ServiceName,
		env:     cfg.Environment,
	}
}

func (l *Logger) Close() {
	l.sender.Close()
}
