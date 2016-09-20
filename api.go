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

    if checks.shouldBuild("loadavg") {
        metrics = append(metrics, LoadAvg())
    }

    if checks.shouldBuild("memory") {
        metrics = append(metrics, Memory())
    }

    if checks.shouldBuild("disk") {
        metrics = append(metrics, DiskUsage())
    }

    if checks.shouldBuild("consul") {
        metrics = append(metrics, ConsulNode())

        for _,s := range ConsulServices() {
            metrics = append(metrics, s)
        }
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

func (c Checks) shouldBuild(checkName string) (bool) {
    if c.List[0] == "all" {
        return true
    }

    for _,v := range c.List {
        if v == checkName { return true }
    }

    return false
}
