package spin

import (
	"encoding/json"
	"fmt"
	"os"
)

type Package struct {
	Os     string `json:"os"`
	Arch   string `json:"arch"`
	URL    string `json:"url"`
	Sha256 string `json:"sha256"`
}

type Manifest struct {
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Homepage          string    `json:"homepage"`
	Version           string    `json:"version"`
	SpinCompatibility string    `json:"spinCompatibility"`
	License           string    `json:"license"`
	Packages          []Package `json:"packages"`
}

// ValidatePlugin validates the plugin spec
func ValidatePlugin(name, file string) error {
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	manifest := &Manifest{}
	err = json.Unmarshal(raw, manifest)
	if err != nil {
		return err
	}

	//TODO: add more validations?
	if manifest.Name == "" {
		return fmt.Errorf("name is required for plugin")
	}

	return nil
}

// GetPluginName gets the plugin name from template .krew.yaml file
func GetPluginName(spec []byte) (string, error) {
	manifest := &Manifest{}
	err := json.Unmarshal(spec, manifest)
	if err != nil {
		return "", err
	}

	return manifest.Name, nil
}

// PluginFileName returns the plugin file with extension
func PluginFileName(name string) string {
	return fmt.Sprintf("%s%s", name, ".json")
}
