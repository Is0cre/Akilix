package workbookview

import (
	"fmt"
	"sort"

	"github.com/Is0cre/Akilix/internal/journal"
	"github.com/Is0cre/Akilix/internal/workbook"
)

const maxDiscoveryKeys = 100000

type Discovery struct {
	Kind              string `json:"kind"`
	Value             string `json:"value"`
	Hostname          string `json:"hostname,omitempty"`
	FirstSeen         string `json:"first_seen"`
	LastSeen          string `json:"last_seen"`
	Occurrences       int    `json:"occurrences"`
	FirstProvenanceID string `json:"first_provenance_id"`
	LastProvenanceID  string `json:"last_provenance_id"`
	LastInvocationID  string `json:"last_invocation_id,omitempty"`
}

// Discoveries builds a read-only, deduplicated projection from canonical
// journal observations. It does not persist an index or modify the workbook.
func Discoveries(root, name string) ([]Discovery, error) {
	if err := workbook.ValidateLayout(root, name); err != nil {
		return nil, err
	}
	workbookRoot, err := Path(root, name, "root")
	if err != nil {
		return nil, err
	}
	items := map[string]*Discovery{}
	err = journal.Visit(workbookRoot, func(event journal.Event) error {
		kind, key := "", ""
		switch event.Event {
		case "HOST_DISCOVERED":
			kind, key = "host", stringPayload(event, "address")
		case "PORT_FOUND":
			kind, key = "port", stringPayload(event, "endpoint")
		default:
			return nil
		}
		if key == "" {
			return fmt.Errorf("discovery journal event lacks identity")
		}
		mapKey := kind + "\x00" + key
		item := items[mapKey]
		if item == nil {
			if len(items) >= maxDiscoveryKeys {
				return fmt.Errorf("discovery projection exceeds %d unique records", maxDiscoveryKeys)
			}
			item = &Discovery{Kind: kind, Value: key, FirstSeen: event.Timestamp, FirstProvenanceID: event.ProvenanceID}
			items[mapKey] = item
		}
		item.LastSeen, item.LastProvenanceID, item.Occurrences = event.Timestamp, event.ProvenanceID, item.Occurrences+1
		if hostname := stringPayload(event, "hostname"); hostname != "" {
			item.Hostname = hostname
		}
		item.LastInvocationID = stringPayload(event, "invocation_id")
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := make([]Discovery, 0, len(items))
	for _, item := range items {
		result = append(result, *item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Kind == result[j].Kind {
			return result[i].Value < result[j].Value
		}
		return result[i].Kind < result[j].Kind
	})
	return result, nil
}

func stringPayload(event journal.Event, key string) string {
	value, _ := event.Payload[key].(string)
	return value
}
