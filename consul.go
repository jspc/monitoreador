package main

import (
    "bytes"
    "fmt"
    "log"
    "github.com/hashicorp/consul/api"
)

func ConsulNode() (m Metric) {
    client, err := api.NewClient(api.DefaultConfig())
    if err != nil {
        log.Fatal(err)
    }

    m.Id = "consul-status"
    m.Name = "Consul"
    m.BusinessImpact = "Containers on this host will not be able to access up to date consul data"
    m.Severity = 1
    m.PanicGuide = panicGuide
    m.Ok = true

    var output bytes.Buffer

    checks, _,_ := client.Health().Node(checkNodeName, nil)
    for _, hc := range checks {
        if hc.Status == "critical" {
            m.Ok = false
        }

        output.WriteString(fmt.Sprintf("%s: %s. ", hc.Name, hc.Output))
    }

    m.CheckOutput = output.String()
    m.LastUpdated = now()

    return
}

func ConsulServices() (ms []Metric) {
    client, err := api.NewClient(api.DefaultConfig())
    if err != nil {
        log.Fatal(err)
    }

    checks,_,_ := client.Catalog().Services(nil)
    for c,_ := range checks {
        services,_,_ := client.Health().Service(c, "", false, nil)

        for _,s := range services {
            if s.Node.Node == checkNodeName {
                var m Metric

                name := fmt.Sprintf("consul-service-%s-%s", c, s.Checks[0].CheckID)

                m.Id = name
                m.Name = name
                m.BusinessImpact = "The cluster probably wont do stuff"
                m.Severity = 1
                m.PanicGuide = panicGuide
                m.Ok = s.Checks[0].Status == "passing"
                m.CheckOutput = s.Checks[0].Output
                m.LastUpdated = now()

                ms = append(ms, m)
            }
        }
    }

    return
}
