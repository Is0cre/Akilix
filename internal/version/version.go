package version

import "runtime"

const (
	Name    = "Akilix"
	Version = "0.0.1-m0"
	Base    = "openSUSE Leap 16"
)

type Info struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Base         string `json:"base"`
	Architecture string `json:"architecture"`
}

func Current() Info {
	architecture := runtime.GOARCH
	if architecture == "amd64" {
		architecture = "x86_64"
	}
	return Info{Name: Name, Version: Version, Base: Base, Architecture: architecture}
}
