package acquire

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const TrustSchema = "akilix.device-trust.v1"

type TrustEntry struct {
	ID             string    `json:"id"`
	Identity       string    `json:"identity"`
	Kind           string    `json:"kind"`
	Vendor         string    `json:"vendor,omitempty"`
	Model          string    `json:"model,omitempty"`
	Serial         string    `json:"serial,omitempty"`
	WWN            string    `json:"wwn,omitempty"`
	USBVendorID    string    `json:"usb_vendor_id,omitempty"`
	USBProductID   string    `json:"usb_product_id,omitempty"`
	USBVendorName  string    `json:"usb_vendor_name,omitempty"`
	USBProductName string    `json:"usb_product_name,omitempty"`
	Label          string    `json:"label,omitempty"`
	AddedAt        time.Time `json:"added_at"`
}

type TrustRegistry struct {
	Schema      string            `json:"schema"`
	Entries     []TrustEntry      `json:"entries"`
	Revocations []TrustRevocation `json:"revocations"`
}

type TrustRevocation struct {
	Entry     TrustEntry `json:"entry"`
	RevokedAt time.Time  `json:"revoked_at"`
}

func TrustIdentity(device Device) (string, string, error) {
	candidates := trustIdentityCandidates(device)
	if len(candidates) == 0 {
		return "", "", fmt.Errorf("device lacks a stable WWN or serial plus hardware identity")
	}
	identity := candidates[0]
	sum := sha256.Sum256([]byte(identity))
	return "TD-" + hex.EncodeToString(sum[:8]), identity, nil
}

func trustIdentityCandidates(device Device) []string {
	wwn := strings.ToLower(strings.TrimSpace(device.WWN))
	serial := strings.ToLower(strings.TrimSpace(device.Serial))
	vendorID := strings.ToLower(strings.TrimSpace(device.USBVendorID))
	productID := strings.ToLower(strings.TrimSpace(device.USBProductID))
	vendor := strings.ToLower(strings.TrimSpace(device.Vendor))
	model := strings.ToLower(strings.TrimSpace(device.Model))
	identities := []string{}
	if wwn != "" {
		identities = append(identities, "block:wwn:"+wwn)
	}
	if serial != "" && usbIDPattern.MatchString(vendorID) && usbIDPattern.MatchString(productID) {
		identities = append(identities, "block:usb:"+vendorID+":"+productID+"|serial:"+serial)
	}
	// Retain the original identity as a compatibility candidate for registries
	// created before numeric USB enrichment was available.
	if serial != "" && (vendor != "" || model != "") {
		identities = append(identities, "block:serial:"+serial+"|vendor:"+vendor+"|model:"+model)
	}
	return identities
}

func LoadTrust(path string) (TrustRegistry, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return TrustRegistry{Schema: TrustSchema, Entries: []TrustEntry{}, Revocations: []TrustRevocation{}}, nil
	}
	if err != nil {
		return TrustRegistry{}, err
	}
	var registry TrustRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return TrustRegistry{}, fmt.Errorf("decode device trust registry: %w", err)
	}
	if registry.Schema != TrustSchema || registry.Entries == nil {
		return TrustRegistry{}, fmt.Errorf("invalid device trust registry")
	}
	if registry.Revocations == nil {
		registry.Revocations = []TrustRevocation{}
	}
	return registry, nil
}

func (r *TrustRegistry) Add(device Device, label string, now time.Time) (TrustEntry, error) {
	id, identity, err := TrustIdentity(device)
	if err != nil {
		return TrustEntry{}, err
	}
	for _, candidate := range trustIdentityCandidates(device) {
		for _, entry := range r.Entries {
			if entry.Identity == candidate {
				return TrustEntry{}, fmt.Errorf("device is already trusted as %s", entry.ID)
			}
		}
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	entry := TrustEntry{ID: id, Identity: identity, Kind: "block", Vendor: device.Vendor, Model: device.Model, Serial: device.Serial, WWN: device.WWN, USBVendorID: device.USBVendorID, USBProductID: device.USBProductID, USBVendorName: device.USBVendorName, USBProductName: device.USBProductName, Label: strings.TrimSpace(label), AddedAt: now.UTC()}
	r.Schema = TrustSchema
	r.Entries = append(r.Entries, entry)
	sort.Slice(r.Entries, func(i, j int) bool { return r.Entries[i].ID < r.Entries[j].ID })
	return entry, nil
}

func (r *TrustRegistry) Remove(id string, now time.Time) (TrustEntry, error) {
	for i, entry := range r.Entries {
		if entry.ID == id {
			if now.IsZero() {
				now = time.Now().UTC()
			}
			r.Entries = append(r.Entries[:i], r.Entries[i+1:]...)
			r.Revocations = append(r.Revocations, TrustRevocation{Entry: entry, RevokedAt: now.UTC()})
			return entry, nil
		}
	}
	return TrustEntry{}, fmt.Errorf("trusted device %q not found", id)
}

func (r TrustRegistry) Match(device Device) (TrustEntry, bool) {
	for _, identity := range trustIdentityCandidates(device) {
		for _, entry := range r.Entries {
			if entry.Identity == identity {
				return entry, true
			}
		}
	}
	return TrustEntry{}, false
}

func SaveTrust(path string, registry TrustRegistry) error {
	if registry.Schema != TrustSchema || registry.Entries == nil || registry.Revocations == nil {
		return fmt.Errorf("invalid device trust registry")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("invalid device trust directory %q", dir)
	}
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(dir, ".trusted-devices-")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(data)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = d.Sync()
	if closeErr := d.Close(); err == nil {
		err = closeErr
	}
	return err
}

func FindWholeDisk(report Inspection, path string) (Device, error) {
	for _, device := range report.Devices {
		if device.Path == path {
			return device, nil
		}
	}
	return Device{}, fmt.Errorf("device %q is not an inspected whole disk", path)
}
