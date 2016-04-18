package main

import (
    "bytes"
    "code.google.com/p/gcfg"
    "encoding/json"
    "flag"
    "fmt"
    "github.com/guillermo/go.procmeminfo"
    "github.com/zenazn/goji"
    "github.com/zenazn/goji/web"
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

type Modes struct {
    Mode map[string]*struct {
        SystemCode string
        ApplicationName string
        PanicGuide string
    }
}

type HealthCheck struct {
    SchemaVersion int
    SystemCode string
    Name string
    Description string
    Checks []Metric
}

type Metric struct {
    Id int
    Name string
    Ok bool
    Severity int
    BusinessImpact string
    PanicGuide string
    CheckOutput string
    LastUpdated string
}

var config string
var err error
var healthcheck HealthCheck
var hostname string
var mode string
var panicGuide string

func init(){
    flag.StringVar(&config, "file", "/etc/monitoreador/config.ini", "Config file for monitoreador")
    flag.StringVar(&config, "f", "/etc/monitoreador/config.ini", "Config file for monitoreador (Shorthand)")

    flag.StringVar(&mode, "mode", "default", "Mode from config file to use")
    flag.StringVar(&mode, "m", "default", "Mode from config file to use (Shorthand)")
}

func LoadConfig(){
    var m Modes
    err := gcfg.ReadFileInto(&m, config)
    if err != nil {
        log.Fatalf("Failed to parse gcfg data: %s", err)
    }

    modeObj := m.Mode[mode]
    log.Printf("Loading configuration")
    log.Printf("Monitoring for: %s, '%s'", mode, modeObj.ApplicationName)
    log.Printf("PCode: %s\n", modeObj.SystemCode)

    panicGuide = modeObj.PanicGuide

    healthcheck.SchemaVersion = 1
    healthcheck.SystemCode = modeObj.SystemCode
    healthcheck.Name = modeObj.ApplicationName
    healthcheck.Description = modeObj.ApplicationName // HAHAHAHA FUCK YOU EVERYBODY
}

func BuildHealthcheck(c web.C, w http.ResponseWriter, r *http.Request){
    var metrics []Metric
    metrics = append(metrics, LoadAvg())
    metrics = append(metrics, Memory())
    metrics = append(metrics, DiskUsage())

    healthcheck.Checks = metrics

    j, err := json.Marshal(healthcheck)
    if err != nil {
        log.Fatalf("Could not marshal projects data: %s\n", err)
    }

    fmt.Fprintf(w, string(j))
}

func LoadAvg() Metric{
    var l Metric
    l.Id = 1
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
    l.Id = 2
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
    l.Id = 3
    l.Name = "Disk Usage"
    l.BusinessImpact = "Running low on disk space could cause assets to disappear"
    l.Severity = 2
    l.PanicGuide = panicGuide
    l.Ok = true

    var stat syscall.Statfs_t
    wd,_ := os.Getwd()

    syscall.Statfs(wd, &stat)

    // Available blocks * size per block = available space in bytes
    available := stat.Bavail * uint64(stat.Bsize)
    var output bytes.Buffer
    if available < (1024*1024*1024) {
        output.WriteString(fmt.Sprintf("%d mb free on disk", available*1024*1024))
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

    goji.Get("/", BuildHealthcheck)
    goji.Serve()
}
