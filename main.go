package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "github.com/guillermo/go.procmeminfo"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/signal"
    "regexp"
    "strconv"
    "syscall"
    "time"
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

func LoadConfig(){
    panicGuide = os.Getenv("SYSTEM_GUIDE")

    controlDir = os.Getenv("CONTROL_DIR")
    if len(controlDir) == 0 {
        controlDir = "/"
    }

    healthcheck.SchemaVersion = 1
    healthcheck.SystemCode = os.Getenv("SYSTEM_CODE")
    healthcheck.Name = os.Getenv("SYSTEM_NAME")
    healthcheck.Description = os.Getenv("SYSTEM_DESCRIPTION")
}

func BuildHealthcheck(w http.ResponseWriter, r *http.Request){
    LogRequest(r)

    var metrics []Metric
    metrics = append(metrics, LoadAvg())
    metrics = append(metrics, Memory())
    metrics = append(metrics, DiskUsage())

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

func LoadAvg() Metric{
    var l Metric
    l.Id = "1"
    l.Name = "Load Average"
    l.BusinessImpact = "A high load could lower the quality of the service"
    l.Severity = 2
    l.PanicGuide = panicGuide
    l.Ok = true

    loadavg, err := ioutil.ReadFile("/proc/loadavg")
    if err != nil {
        log.Fatalf("Could not read /proc/loadavg: %s\n", err)
    }

    cpuInfo, err := ioutil.ReadFile("/proc/cpuinfo")
    if err != nil {
        log.Fatalf("Could not read /proc/cpuinfo: %s\n", err)
    }

    r,_ := regexp.Compile(`processor[\s]*:[\s]*[\d]`)
    cpuCount := len(r.FindAll(cpuInfo, -1))

    loadFields := bytes.Fields(loadavg)
    var output bytes.Buffer

    if cmp(loadFields[0], cpuCount) {
        output.WriteString(fmt.Sprintf("1 minute value is too high: %s. ", loadFields[0]))
        l.Ok = false
    }

    if cmp(loadFields[1], cpuCount) {
        output.WriteString(fmt.Sprintf("5 minute value is too high: %s. ", loadFields[1]))
        l.Ok = false
    }

    if cmp(loadFields[2], cpuCount) {
        output.WriteString(fmt.Sprintf("15 minute value is too high: %s. ", loadFields[2]))
        l.Ok = false
    }


    l.CheckOutput = output.String()
    l.LastUpdated = fmt.Sprintf("%s", time.Now())

    return l
}

func Memory() Metric{
    var l Metric
    l.Id = "2"
    l.Name = "Memory Usage"
    l.BusinessImpact = "Running low on memory could cause the operating system to close important processes"
    l.Severity = 2
    l.PanicGuide = panicGuide
    l.Ok = true

    meminfo := &procmeminfo.MemInfo{}
    meminfo.Update()

    var output bytes.Buffer

    percent := (meminfo.Used() / meminfo.Total()) * 100
    if  percent > 90.0 {
        output.WriteString(fmt.Sprintf("%d percent used", percent))
        l.Ok = false
    }

    l.CheckOutput = output.String()
    l.LastUpdated = fmt.Sprintf("%s", time.Now())

    return l
}

func DiskUsage() Metric{
    var l Metric
    l.Id = "3"
    l.Name = "Disk Usage"
    l.BusinessImpact = "Running low on disk space could cause assets to disappear"
    l.Severity = 2
    l.PanicGuide = panicGuide
    l.Ok = true

    var stat syscall.Statfs_t
    syscall.Statfs(controlDir, &stat)

    // Available blocks * size per block = available space in bytes
    available := stat.Bavail * uint64(stat.Bsize)

    var output bytes.Buffer
    if available < (1024*1024*1024) {
        output.WriteString(fmt.Sprintf("%d mb free on disk", available/1024/1024))
        l.Ok = false
    }

    l.CheckOutput = output.String()
    l.LastUpdated = fmt.Sprintf("%s", time.Now())

    return l
}

func cmp(a []byte, b int)(bool){
    s := string(a)
    f,err := strconv.ParseFloat(s,64)

    if err != nil {
        log.Fatalf("err: %v\n", err)
    } else {
        if int(f) > b {
            return true
        }
    }
    return false
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
