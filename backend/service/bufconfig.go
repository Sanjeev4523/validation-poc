package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"validation-service/backend/logger"
)

// ParseBSRModuleFromBufYAML parses buf.yaml to extract the default BSR module
// Returns org and module name, or error if not found
func ParseBSRModuleFromBufYAML(basePath string) (org, module string, err error) {
	bufYAMLPath := filepath.Join(basePath, "buf.yaml")
	logger.Debug("Parsing BSR module from buf.yaml at: %s", bufYAMLPath)

	data, err := os.ReadFile(bufYAMLPath)
	if err != nil {
		logger.Error("Failed to read buf.yaml from %s: %v", bufYAMLPath, err)
		return "", "", fmt.Errorf("failed to read buf.yaml: %w", err)
	}

	// Look for the module name pattern: name: buf.build/{org}/{module}
	// Pattern should match both formats:
	//   name: buf.build/{org}/{module}
	//   - name: buf.build/{org}/{module} (in array)
	pattern := regexp.MustCompile(`name:\s*buf\.build/([^/\s\n]+)/([^\s\n]+)`)
	matches := pattern.FindStringSubmatch(string(data))

	if len(matches) < 3 {
		logger.Error("Could not find module name pattern in buf.yaml")
		return "", "", fmt.Errorf("could not find module name in buf.yaml")
	}

	org = strings.TrimSpace(matches[1])
	module = strings.TrimSpace(matches[2])
	logger.Debug("Successfully parsed BSR module from buf.yaml: org=%s, module=%s", org, module)

	return org, module, nil
}

// GetBSRConfig extracts BSR org and module, with fallback to defaults
func GetBSRConfig(basePath string) (org, module string) {
	org, module, err := ParseBSRModuleFromBufYAML(basePath)
	if err != nil {
		// Fallback to default values from buf.yaml
		// Default: buf.build/sanjeev-personal/validation
		logger.Warn("Failed to parse buf.yaml, using default BSR config: org=sanjeev-personal, module=validation (error: %v)", err)
		return "sanjeev-personal", "validation"
	}
	logger.Info("Using BSR config from buf.yaml: org=%s, module=%s", org, module)
	return org, module
}
