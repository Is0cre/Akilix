package acquire

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var usbIDPattern = regexp.MustCompile(`^[0-9a-f]{4}$`)

// EnrichUSB reads udev properties and a local usb.ids file. It never contacts
// the network or opens the block device.
func EnrichUSB(ctx context.Context, runner Runner, device Device, databasePath string) (Device, error) {
	out, err := runner.Run(ctx, "udevadm", "info", "--query=property", "--name", device.Path)
	if err != nil {
		return device, fmt.Errorf("query udev device identity: %w", err)
	}
	properties := parseProperties(string(out))
	if strings.ToLower(properties["ID_BUS"]) != "usb" {
		return device, nil
	}
	vendorID := strings.ToLower(properties["ID_VENDOR_ID"])
	productID := strings.ToLower(properties["ID_MODEL_ID"])
	if !usbIDPattern.MatchString(vendorID) || !usbIDPattern.MatchString(productID) {
		return device, fmt.Errorf("USB device lacks valid numeric vendor/product IDs")
	}
	device.USBVendorID, device.USBProductID = vendorID, productID
	if serial := strings.TrimSpace(properties["ID_SERIAL_SHORT"]); serial != "" {
		device.Serial = serial
	}
	data, err := os.ReadFile(databasePath)
	if err != nil && !os.IsNotExist(err) {
		return device, err
	}
	if err == nil {
		device.USBVendorName, device.USBProductName = LookupUSBNames(data, vendorID, productID)
	}
	return device, nil
}

func parseProperties(value string) map[string]string {
	properties := make(map[string]string)
	for _, line := range strings.Split(value, "\n") {
		key, item, ok := strings.Cut(line, "=")
		if ok {
			properties[key] = strings.TrimSpace(item)
		}
	}
	return properties
}

func LookupUSBNames(data []byte, vendorID, productID string) (string, string) {
	vendorID, productID = strings.ToLower(vendorID), strings.ToLower(productID)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	insideVendor := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "C ") || strings.HasPrefix(line, "AT ") {
			continue
		}
		if line[0] != '\t' {
			fields := strings.Fields(line)
			insideVendor = len(fields) >= 2 && strings.ToLower(fields[0]) == vendorID
			if insideVendor {
				vendorName := strings.TrimSpace(strings.TrimPrefix(line, fields[0]))
				for scanner.Scan() {
					productLine := scanner.Text()
					if productLine == "" || strings.HasPrefix(productLine, "#") {
						continue
					}
					if productLine[0] != '\t' {
						return vendorName, ""
					}
					trimmed := strings.TrimPrefix(productLine, "\t")
					if strings.HasPrefix(trimmed, "\t") {
						continue
					}
					productFields := strings.Fields(trimmed)
					if len(productFields) >= 2 && strings.ToLower(productFields[0]) == productID {
						return vendorName, strings.TrimSpace(strings.TrimPrefix(trimmed, productFields[0]))
					}
				}
				return vendorName, ""
			}
		}
	}
	return "", ""
}
