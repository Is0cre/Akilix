package playbook

import (
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/Is0cre/Akilix/internal/journal"
	"github.com/Is0cre/Akilix/internal/scope"
)

const maxNmapXMLSize = 64 * 1024 * 1024

type nmapHost struct {
	Status struct {
		State string `xml:"state,attr"`
	} `xml:"status"`
	Addresses []struct {
		Address string `xml:"addr,attr"`
		Type    string `xml:"addrtype,attr"`
	} `xml:"address"`
	Hostnames []struct {
		Name string `xml:"name,attr"`
	} `xml:"hostnames>hostname"`
	Ports []struct {
		Protocol string `xml:"protocol,attr"`
		PortID   string `xml:"portid,attr"`
		State    struct {
			Value string `xml:"state,attr"`
		} `xml:"state"`
		Service struct {
			Name string `xml:"name,attr"`
		} `xml:"service"`
	} `xml:"ports>port"`
}

// IngestNmapJournal reads a completed managed invocation's XML artifact. It
// never rewrites the artifact and re-evaluates every reported address through
// the canonical workbook scope engine before creating journal observations.
func IngestNmapJournal(path, invocationID string, config scope.Config, log *journal.Journal, now func() time.Time) (found, dropped int, err error) {
	if invocationID == "" || log == nil || now == nil {
		return 0, 0, fmt.Errorf("invalid Nmap journal ingestion context")
	}
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_CLOEXEC|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return 0, 0, err
	}
	file := os.NewFile(uintptr(fd), path)
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxNmapXMLSize {
		return 0, 0, fmt.Errorf("invalid Nmap XML artifact")
	}
	decoder := xml.NewDecoder(io.LimitReader(file, maxNmapXMLSize+1))
	for {
		token, decodeErr := decoder.Token()
		if decodeErr == io.EOF {
			break
		}
		if decodeErr != nil {
			return found, dropped, fmt.Errorf("decode Nmap XML: %w", decodeErr)
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "host" {
			continue
		}
		var host nmapHost
		if err := decoder.DecodeElement(&host, &start); err != nil {
			return found, dropped, fmt.Errorf("decode Nmap host: %w", err)
		}
		if host.Status.State != "up" {
			continue
		}
		for _, address := range host.Addresses {
			if address.Type != "ipv4" && address.Type != "ipv6" {
				continue
			}
			decision := scope.EvaluateDecision(config, address.Address)
			eventName := "HOST_DISCOVERED"
			if decision.Result != scope.Allow {
				eventName = "HOST_DROPPED_OUT_OF_SCOPE"
				dropped++
			} else {
				found++
			}
			payload := map[string]any{"invocation_id": invocationID, "address": address.Address, "address_type": address.Type, "scope_result": decision.Result}
			if len(host.Hostnames) > 0 && host.Hostnames[0].Name != "" {
				payload["hostname"] = host.Hostnames[0].Name
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
			for _, port := range host.Ports {
				if port.State.Value != "open" {
					continue
				}
				portNumber, err := strconv.Atoi(port.PortID)
				if err != nil || portNumber < 1 || portNumber > 65535 || port.Protocol == "" {
					return found, dropped, fmt.Errorf("invalid Nmap port result")
				}
				portEventName := "PORT_FOUND"
				if decision.Result != scope.Allow {
					portEventName = "PORT_DROPPED_OUT_OF_SCOPE"
				}
				portPayload := map[string]any{"invocation_id": invocationID, "target": address.Address, "port": portNumber, "protocol": port.Protocol, "endpoint": net.JoinHostPort(address.Address, port.PortID), "scope_result": decision.Result}
				if port.Service.Name != "" {
					portPayload["service"] = port.Service.Name
				}
				if decision.Rule != "" {
					portPayload["scope_rule"] = decision.Rule
				}
				portEvent, err := journal.NewEvent(portEventName, "RECON", portPayload, now())
				if err != nil {
					return found, dropped, err
				}
				if err := log.Append(portEvent); err != nil {
					return found, dropped, err
				}
			}
		}
	}
	return found, dropped, nil
}
