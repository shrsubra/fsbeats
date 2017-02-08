package fs

import (
    "github.com/elastic/beats/libbeat/common"
    "github.com/elastic/beats/metricbeat/mb"
    "github.com/elastic/beats/libbeat/logp"
    dc "github.com/fsouza/go-dockerclient"
    "github.com/elastic/beats/metricbeat/module/docker"
    sigar "github.com/elastic/gosigar"
)

// init registers the MetricSet with the central registry.
// The New method will be called after the setup of the module and before starting to fetch data
func init() {
    if err := mb.Registry.AddMetricSet("docker", "fs", New, docker.HostParser); err != nil {
        panic(err)
    }
}

// MetricSet type defines all fields of the MetricSet
// As a minimum it must inherit the mb.BaseMetricSet fields, but can be extended with
// additional entries. These variables can be used to persist data or configuration between
// multiple fetch calls.
type MetricSet struct {
    mb.BaseMetricSet
    dockerClient *dc.Client
    rootDir string
}

// New create a new instance of the MetricSet
// Part of new is also setting up the configuration by processing additional
// configuration entries if needed.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {

    config := docker.Config{}

    if err := base.Module().UnpackConfig(&config); err != nil {
        return nil, err
    }
    client, err := docker.NewDockerClient(base.HostData().URI, config)
    if err != nil {
        return nil, err
    }

    info, err:= client.Info()
    if err != nil {
        return nil, err
    }
    dockerRootDir := info.DockerRootDir + "/" + info.Driver + "/mnt/"

    return &MetricSet{
        BaseMetricSet: base,
        dockerClient: client,
        rootDir: dockerRootDir,
    }, nil
}

type FileSystemUsage struct {
    Container *docker.Container
    Total uint64
    Free uint64
    Used uint64
    UsedPct float64
}

func DiskUsage(path string) (*FileSystemUsage, error) {
    stat := sigar.FileSystemUsage{}
    if err := stat.Get(path); err != nil {
        return nil, err
    }
    usedPct := float64(stat.Used) / float64(stat.Total)
    usage := FileSystemUsage{
        Total: stat.Total,
        Free: stat.Free,
        Used: stat.Used,
        UsedPct: usedPct,
    }
    return &usage, nil

}

// Fetch methods implements the data gathering and data conversion to the right format
// It returns the event which is then forward to the output. In case of an error, a
// descriptive error must be returned.
func (m *MetricSet) Fetch() ([]common.MapStr, error) {

    containers, err := m.dockerClient.ListContainers(dc.ListContainersOptions{})
    if err != nil {
        return nil, err
    }
    usages := []FileSystemUsage{}
    for _, container := range containers {
        p := m.rootDir + container.ID
        du, err := DiskUsage(p)
        if err != nil {
            logp.Warn("Stat failed for %v with %v", p, err)
            continue
        }
        du.Container = docker.NewContainer(&container)
        usages = append(usages, *du)
    }
    return eventsMapping(usages), nil
}

