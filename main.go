package main

import (
    "flag"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
)

type HealthCheck struct {
    SchemaVersion int     `json:"schemaVersion"`
    SystemCode string     `json:"systemCode"`
    Name string           `json:"name"`
    Description string    `json:"description"`
    Checks []Metric       `json:"checks"`
}

type Metric struct {
    Id string             `json:"id"`
    Name string           `json:"name"`
    Ok bool               `json:"ok"`
    Severity int          `json:"severity"`
    BusinessImpact string `json:"businessImpact"`
    PanicGuide string     `json:"panicGuide"`
    CheckOutput string    `json:"checkOutput"`
    LastUpdated string    `json:"lastUpdated"`
}

var config string
var err error
var healthcheck HealthCheck
var hostname string
var panicGuide string

var controlDir string
var checkNodeName string
var consulAddress string

func LoadConfig(){
    panicGuide = os.Getenv("SYSTEM_GUIDE")

    controlDir = os.Getenv("CONTROL_DIR")
    if len(controlDir) == 0 {
        controlDir = "/"
    }

    checkNodeName = os.Getenv("CONSUL_NODENAME")
    if len(checkNodeName) == 0 {
        checkNodeName,_ = os.Hostname()
    }

    healthcheck.SchemaVersion = 1
    healthcheck.SystemCode = os.Getenv("SYSTEM_CODE")
    healthcheck.Name = os.Getenv("SYSTEM_NAME")
    healthcheck.Description = os.Getenv("SYSTEM_DESCRIPTION")
}

func main(){
    flag.Parse()
    hostname, err = os.Hostname()
    if err != nil {
        log.Fatalf("Could not get hostname: %s", err)
    }

    log.Printf( "Starting logging on %s\n", hostname )

    LoadConfig()

    reload := make(chan os.Signal, 1)
    signal.Notify(reload, syscall.SIGHUP)
    go func(){
        for sig := range reload {
            log.Printf("Received %s - reloading profile\n", sig)
            LoadConfig()
        }
    }()

    http.HandleFunc("/", BuildHealthcheck)
    http.ListenAndServe(":8000", nil)
}
