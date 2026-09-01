package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Is0cre/Akilix/internal/deviceevent"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: akilix-udev-handler <add|remove> /dev/DEVICE")
		os.Exit(2)
	}
	properties := map[string]string{}
	for _, key := range []string{"ID_SERIAL_SHORT", "ID_VENDOR", "ID_MODEL", "ID_BUS"} {
		properties[key] = os.Getenv(key)
	}
	kernelRO := os.Args[1] == "add" && readOnly(filepath.Base(os.Args[2]))
	event, err := deviceevent.New(os.Args[1], os.Args[2], properties, kernelRO, time.Now().UTC())
	if err == nil {
		_, err = deviceevent.Append("/run/akilix/device-events", event)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func readOnly(device string) bool {
	data, err := os.ReadFile(filepath.Join("/sys/class/block", device, "ro"))
	return err == nil && strings.TrimSpace(string(data)) == "1"
}
