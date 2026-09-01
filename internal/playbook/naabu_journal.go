package playbook

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Is0cre/Akilix/internal/journal"
	"github.com/Is0cre/Akilix/internal/scope"
)

type naabuResult struct {
	Host     string `json:"host"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

func IngestNaabuJournal(path, invocationID string, config scope.Config, log *journal.Journal, now func() time.Time) (found, dropped int, err error) {
	if invocationID == "" || log == nil || now == nil {
		return 0, 0, fmt.Errorf("invalid Naabu journal ingestion context")
	}
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		var result naabuResult
		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			return found, dropped, fmt.Errorf("decode Naabu JSONL: %w", err)
		}
		target := result.IP
		if target == "" {
			target = result.Host
		}
		if target == "" || result.Port < 1 || result.Port > 65535 {
			return found, dropped, fmt.Errorf("invalid Naabu result")
		}
		decision := scope.EvaluateDecision(config, target)
		eventName := "PORT_FOUND"
		if decision.Result != scope.Allow {
			eventName = "PORT_DROPPED_OUT_OF_SCOPE"
			dropped++
		} else {
			found++
		}
		payload := map[string]any{"invocation_id": invocationID, "target": target, "port": result.Port, "endpoint": target + ":" + strconv.Itoa(result.Port), "scope_result": decision.Result}
		if result.Protocol != "" {
			payload["protocol"] = result.Protocol
		}
		if decision.Rule != "" {
			payload["scope_rule"] = decision.Rule
		}
		event, err := journal.NewEvent(eventName, "RECON", payload, now())
		if err != nil {
			return found, dropped, err
		}
		if err := log.Append(event); err != nil {
			return found, dropped, err
		}
	}
	if err := scanner.Err(); err != nil {
		return found, dropped, err
	}
	return found, dropped, nil
}
