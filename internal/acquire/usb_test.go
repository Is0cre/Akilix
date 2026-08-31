package acquire

import (
	"context"
	"testing"
)

func TestLookupUSBNames(t *testing.T) {
	data := []byte("1234  Example Vendor\n\t0001  First Device\n\t0002  Evidence Reader\n\t\t01  Interface\n5678  Other\n")
	vendor, product := LookupUSBNames(data, "1234", "0002")
	if vendor != "Example Vendor" || product != "Evidence Reader" {
		t.Fatalf("vendor=%q product=%q", vendor, product)
	}
}

func TestEnrichUSBUsesExactUdevArgumentsAndLocalDatabase(t *testing.T) {
	runner := &fakeRunner{out: []byte("ID_BUS=usb\nID_VENDOR_ID=1234\nID_MODEL_ID=0002\nID_SERIAL_SHORT=SERIAL\n")}
	path := t.TempDir() + "/missing-usb.ids"
	device, err := EnrichUSB(context.Background(), runner, Device{Path: "/dev/sdb"}, path)
	if err != nil {
		t.Fatal(err)
	}
	if runner.name != "udevadm" || len(runner.args) != 4 || runner.args[0] != "info" || runner.args[3] != "/dev/sdb" {
		t.Fatalf("command=%s args=%v", runner.name, runner.args)
	}
	if device.USBVendorID != "1234" || device.USBProductID != "0002" || device.Serial != "SERIAL" {
		t.Fatalf("device=%+v", device)
	}
}

func TestEnrichUSBLeavesNonUSBDeviceAlone(t *testing.T) {
	runner := &fakeRunner{out: []byte("ID_BUS=nvme\n")}
	original := Device{Path: "/dev/nvme0n1", Serial: "N1"}
	device, err := EnrichUSB(context.Background(), runner, original, "/missing")
	if err != nil || device.Serial != original.Serial || device.USBVendorID != "" {
		t.Fatalf("device=%+v err=%v", device, err)
	}
}
