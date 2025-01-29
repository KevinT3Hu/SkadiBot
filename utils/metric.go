package utils

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartMetric() {
	metric_port := os.Getenv("METRIC_PORT")
	if metric_port == "" {
		metric_port = "2112"
	}
	servmux := http.NewServeMux()

	servmux.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+metric_port, servmux)
}

var (
	latencyBucket   = []float64{1, 3, 5, 7, 10, 20, 50, 100, 500, 1000}
	aiLatencyBucket = []float64{1000, 2000, 4000, 5000, 7500, 10000, 15000, 20000, 30000, 50000, 100000}
)

var (
	MessageRecCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "message_total",
		Help: "The total number of messages accepted",
	})
	MessageHitCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "message_hit_total",
		Help: "The total number of messages hit",
	})
	MessageMissCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "message_miss_total",
		Help: "The total number of messages missed",
	})
	MessageHitReplyCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "message_hit_reply_total",
		Help: "The total number of messages hit and replied",
	})
	MessageMissReplyCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "message_miss_reply_total",
		Help: "The total number of messages missed and replied",
	})

	Doc2vecLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "doc2vec_latency",
		Help:    "Latency of doc2vec",
		Buckets: latencyBucket,
	})
	TotalLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "total_latency",
		Help:    "Total latency",
		Buckets: latencyBucket,
	})

	AIRequestCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ai_request_total",
		Help: "The total number of AI requests",
	})

	AIRequestLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "ai_request_latency",
		Help:    "Latency of AI request",
		Buckets: aiLatencyBucket,
	})
)
