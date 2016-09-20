package main

import (
    "log"
    "io/ioutil"
    "regexp"
    "bytes"
    "fmt"
    "strconv"
)

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
    l.LastUpdated = now()

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
