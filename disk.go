package main

import (
    "fmt"
    "syscall"
    "bytes"
)

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
    l.LastUpdated = now()

    return l
}
