package main

import (
    "net/http"
    "log"
    "fmt"
    "encoding/json"
)

func BuildHealthcheck(w http.ResponseWriter, r *http.Request){
    LogRequest(r)

    var metrics []Metric
    metrics = append(metrics, LoadAvg())
    metrics = append(metrics, Memory())
    metrics = append(metrics, DiskUsage())
    metrics = append(metrics, ConsulNode())

    for _,s := range ConsulServices() {
        metrics = append(metrics, s)
    }

    healthcheck.Checks = metrics

    j, err := json.Marshal(healthcheck)
    if err != nil {
        log.Fatalf("Could not marshal projects data: %s\n", err)
    }

    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, string(j))
}

func LogRequest(r *http.Request) {
    log.Printf( "%s :: %s %s",
        r.RemoteAddr,
        r.Method,
        r.URL.Path)
}
