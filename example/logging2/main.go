package main

import (
	"io"
	"log/slog"
	"log/syslog"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func main() {
	logstash, err := syslog.Dial("udp", "logstash:5044", syslog.LOG_INFO, "test")
	if err != nil {
		slog.Error("error syslog dial", "error", err)
		return
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.MultiWriter(os.Stdout, logstash), nil)))
	slog.Info("start")

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if rand.Intn(10) > 0 {
			w.WriteHeader(200)
			w.Write([]byte("OK\n"))
			slog.Info("OK", "user", r.UserAgent(), "path", r.URL.Path, "duration", time.Since(start), "code", 200)

			return
		}

		w.WriteHeader(500)
		w.Write([]byte("NE OK\n"))
		slog.Error("NE OK", "user", r.UserAgent(), "path", r.URL.Path,  "duration", time.Since(start),"code", 500)
	}))

	http.ListenAndServe(":8080", nil)
}
