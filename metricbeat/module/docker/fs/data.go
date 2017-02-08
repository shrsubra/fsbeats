package fs

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/metricbeat/mb"
)

func eventsMapping(usageList []FileSystemUsage) []common.MapStr {
	events := []common.MapStr{}
	for _, usage := range usageList {
		events = append(events, eventMapping(&usage))
	}
	return events
}

func eventMapping(stats *FileSystemUsage) common.MapStr {
	event := common.MapStr{
		mb.ModuleData: common.MapStr{
			"container": stats.Container.ToMapStr(),
		},
		"total": stats.Total,
		"free": stats.Free,
		"used": common.MapStr{
			"bytes": stats.Used,
			"pct":   stats.UsedPct,
		},
	}
	return event
}
