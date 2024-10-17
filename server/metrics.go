package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"runtime/debug"
	"time"

	"gitlab.com/mstarongitlab/goutils/other"
)

func buildMetricsHandler() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/", profilingRootHandler)
	router.HandleFunc("GET /current-goroutines", metricActiveGoroutinesHandler)
	router.HandleFunc("GET /memory", metricMemoryStatsHandler)
	router.HandleFunc("GET /pprof/cpu", pprof.Profile)
	router.Handle("GET /pprof/memory", pprof.Handler("heap"))
	router.Handle("GET /pprof/goroutines", pprof.Handler("goroutine"))
	router.Handle("GET /pprof/blockers", pprof.Handler("block"))

	return profilingAuthenticationMiddleware(router)
}

func setupProfilingHandler() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/", profilingRootHandler)
	router.HandleFunc("GET /current-goroutines", metricActiveGoroutinesHandler)
	router.HandleFunc("GET /memory", metricMemoryStatsHandler)
	router.HandleFunc("GET /pprof/cpu", pprof.Profile)
	router.Handle("GET /pprof/memory", pprof.Handler("heap"))
	router.Handle("GET /pprof/goroutines", pprof.Handler("goroutine"))
	router.Handle("GET /pprof/blockers", pprof.Handler("block"))

	return router
}

func isAliveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "yup")
}

func profilingRootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(
		w,
		"Endpoints: /, /{memory,current-goroutines}, /pprof/{cpu,memory,goroutines,blockers}",
	)
}

func metricActiveGoroutinesHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{\"goroutines\": %d}", runtime.NumGoroutine())
}

func metricMemoryStatsHandler(w http.ResponseWriter, r *http.Request) {
	type OutData struct {
		CollectedAt          time.Time `json:"collected_at"`
		HeapUsed             uint64    `json:"heap_used"`
		HeapIdle             uint64    `json:"heap_idle"`
		StackUsed            uint64    `json:"stack_used"`
		GCLastFired          time.Time `json:"gc_last_fired"`
		GCNextTargetHeapSize uint64    `json:"gc_next_target_heap_size"`
	}
	stats := runtime.MemStats{}
	gcStats := debug.GCStats{}
	runtime.ReadMemStats(&stats)
	debug.ReadGCStats(&gcStats)
	outData := OutData{
		CollectedAt:          time.Now(),
		HeapUsed:             stats.HeapInuse,
		HeapIdle:             stats.HeapIdle,
		StackUsed:            stats.StackInuse,
		GCLastFired:          gcStats.LastGC,
		GCNextTargetHeapSize: stats.NextGC,
	}

	jsonData, err := json.Marshal(&outData)
	if err != nil {
		other.HttpErr(
			w,
			HttpErrIdJsonMarshalFail,
			"Failed to encode return data",
			http.StatusInternalServerError,
		)
		return
	}
	fmt.Fprint(w, string(jsonData))
}
