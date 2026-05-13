package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
)

type LogstashSender struct {
	client *http.Client
	url    string
	queue  chan map[string]any
	wg     sync.WaitGroup
}

type CustomHandler struct {
	slog.Handler
	sender *LogstashSender
}

func NewLogstashSender(url string) *LogstashSender {
	ls := &LogstashSender{
		client: &http.Client{},
		url:    url,
		queue:  make(chan map[string]any, 1000),
	}

	ls.wg.Add(1)
	go ls.processQueue()
	return ls
}

func (ls *LogstashSender) processQueue() {
	defer ls.wg.Done()
	for entry := range ls.queue {
		data, _ := json.Marshal(entry)
		go func(d []byte) {
			http.Post(ls.url, "application/json", bytes.NewReader(d))
		}(data)
	}
}

func (ls *LogstashSender) Send(entry map[string]any) {
	select {
	case ls.queue <- entry:
	default: // drop if full – production should monitor queue length
	}
}

func (ls *LogstashSender) Close() {
	close(ls.queue)
	ls.wg.Wait()
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	// also send to external
	entry := map[string]any{
		"level":   r.Level,
		"message": r.Message,
		"time":    r.Time,
	}
	h.sender.Send(entry)
	return h.Handler.Handle(ctx, r)
}
