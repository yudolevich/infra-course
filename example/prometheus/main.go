package main

import (
	"math/rand"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	reqTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "app_req_total"},
		[]string{"code"},
	)
	prometheus.MustRegister(reqTotal)
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rand.Intn(10) > 0 {
			w.WriteHeader(200)
			w.Write([]byte("OK\n"))
			reqTotal.WithLabelValues("200").Inc()

			return
		}

		w.WriteHeader(500)
		w.Write([]byte("NE OK\n"))
		reqTotal.WithLabelValues("500").Inc()
	}))
	http.ListenAndServe(":8080", nil)
}
