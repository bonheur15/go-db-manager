package utils

import (
	"runtime"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

type ServerInfo struct {
	CPUInfo      []cpu.InfoStat         `json:"cpu_info"`
	MemInfo      *mem.VirtualMemoryStat `json:"mem_info"`
	LoadInfo     *load.AvgStat          `json:"load_info"`
	HostInfo     *host.InfoStat         `json:"host_info"`
	GoVersion    string                 `json:"go_version"`
	NumCPU       int                    `json:"num_cpu"`
	NumGoroutine int                    `json:"num_goroutine"`
	OS           string                 `json:"os"`
	Arch         string                 `json:"arch"`
}

func GetServerInfo() (*ServerInfo, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	loadInfo, err := load.Avg()
	if err != nil {
		return nil, err
	}

	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	serverInfo := &ServerInfo{
		CPUInfo:      cpuInfo,
		MemInfo:      memInfo,
		LoadInfo:     loadInfo,
		HostInfo:     hostInfo,
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
	}

	return serverInfo, nil
}
