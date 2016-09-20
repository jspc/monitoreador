package main

import (
    "bytes"
    "fmt"
    "github.com/guillermo/go.procmeminfo"
)

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
    l.LastUpdated = now()

    return l
}
