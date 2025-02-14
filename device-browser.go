package main

import (
	"fmt"
	"regexp"
	"strings"
)

// DeviceInfo represents detailed information about the user's device
type DeviceInfo struct {
    OS              string  `json:"os"`              // Operating system name
    Platform        string  `json:"platform"`        // Platform (desktop/mobile/tablet)
    Version         *string `json:"version"`         // OS version (nullable)
    Browser         *string `json:"browser"`         // Browser name (nullable)
    BrowserVersion  *string `json:"browserVersion"`  // Browser version (nullable)
    Engine          *string `json:"engine"`          // Browser engine (nullable)
    EngineVersion   *string `json:"engineVersion"`   // Engine version (nullable)
    Architecture    *string `json:"architecture"`    // CPU architecture (nullable)
}

// String implements the Stringer interface for pretty printing
func (d DeviceInfo) String() string {
    // Helper function to safely get string value from pointer
    getValue := func(s *string) string {
        if s == nil {
            return "null"
        }
        return *s
    }

    return fmt.Sprintf(
        "DeviceInfo{\n"+
            "  OS: %s\n"+
            "  Platform: %s\n"+
            "  Version: %s\n"+
            "  Browser: %s\n"+
            "  BrowserVersion: %s\n"+
            "  Engine: %s\n"+
            "  EngineVersion: %s\n"+
            "  Architecture: %s\n"+
            "}",
        d.OS,
        d.Platform,
        getValue(d.Version),
        getValue(d.Browser),
        getValue(d.BrowserVersion),
        getValue(d.Engine),
        getValue(d.EngineVersion),
        getValue(d.Architecture),
    )
}

// ToMap converts DeviceInfo to a map for easy access
func (d DeviceInfo) ToMap() map[string]interface{} {
    getValue := func(s *string) interface{} {
        if s == nil {
            return nil
        }
        return *s
    }

    return map[string]interface{}{
        "os":             d.OS,
        "platform":       d.Platform,
        "version":        getValue(d.Version),
        "browser":        getValue(d.Browser),
        "browserVersion": getValue(d.BrowserVersion),
        "engine":         getValue(d.Engine),
        "engineVersion":  getValue(d.EngineVersion),
        "architecture":   getValue(d.Architecture),
    }
}

func getDeviceInfo(userAgent string) DeviceInfo {
    // Convert to lowercase for case-insensitive matching
    ua := strings.ToLower(userAgent)
    
    // Initialize device info with default values
    info := DeviceInfo{
        OS:       "Unknown",
        Platform: "Unknown",
    }

    // OS detection patterns
    osPatterns := map[string][]string{
        "Windows": {
            `windows nt 10\.0`,
            `windows nt 6\.3`,
            `windows nt 6\.2`,
            `windows nt 6\.1`,
            `windows nt 6\.0`,
            `windows nt 5\.2`,
            `windows nt 5\.1`,
            `windows`,
        },
        "MacOS": {
            `mac os x`,
            `macintosh`,
            `darwin`,
        },
        "iOS": {
            `iphone`,
            `ipad`,
            `ipod`,
            `ios`,
        },
        "Android": {
            `android`,
        },
        "Linux": {
            `linux`,
            `ubuntu`,
            `fedora`,
            `debian`,
            `sunos`,
        },
    }

    // Version detection patterns
    versionPatterns := map[string]string{
        "Windows": `windows nt (\d+\.\d+)`,
        "MacOS":   `mac os x (\d+[._]\d+[._]\d+)`,
        "Android": `android (\d+(\.\d+)?)`,
        "iOS":     `os (\d+[._]\d+)`,
    }

    // Browser detection patterns
    browserPatterns := map[string]string{
        "Chrome":    `chrome\/(\d+(\.\d+)?)`,
        "Firefox":   `firefox\/(\d+(\.\d+)?)`,
        "Safari":    `safari\/(\d+(\.\d+)?)`,
        "Edge":      `edg\/(\d+(\.\d+)?)`,
        "Opera":     `opr\/(\d+(\.\d+)?)`,
        "IE":        `msie (\d+(\.\d+)?)`,
    }

    // Engine detection patterns
    enginePatterns := map[string]string{
        "WebKit":    `webkit\/(\d+(\.\d+)?)`,
        "Gecko":     `gecko\/(\d+)`,
        "Blink":     `blink\/(\d+(\.\d+)?)`,
        "Trident":   `trident\/(\d+(\.\d+)?)`,
    }

    // Detect OS and version
    for os, patterns := range osPatterns {
        for _, pattern := range patterns {
            if regexp.MustCompile(pattern).MatchString(ua) {
                info.OS = os
                if vPattern, exists := versionPatterns[os]; exists {
                    if matches := regexp.MustCompile(vPattern).FindStringSubmatch(ua); len(matches) > 1 {
                        version := strings.Replace(matches[1], "_", ".", -1)
                        info.Version = &version
                    }
                }
                break
            }
        }
    }

    // Detect platform
    if strings.Contains(ua, "mobile") || strings.Contains(ua, "tablet") {
        if strings.Contains(ua, "tablet") {
            info.Platform = "Tablet"
        } else {
            info.Platform = "Mobile"
        }
    } else {
        info.Platform = "Desktop"
    }

    // Detect browser and version
    for browser, pattern := range browserPatterns {
        if matches := regexp.MustCompile(pattern).FindStringSubmatch(ua); len(matches) > 1 {
            browserName := browser
            browserVersion := matches[1]
            info.Browser = &browserName
            info.BrowserVersion = &browserVersion
            break
        }
    }

    // Detect engine and version
    for engine, pattern := range enginePatterns {
        if matches := regexp.MustCompile(pattern).FindStringSubmatch(ua); len(matches) > 1 {
            engineName := engine
            engineVersion := matches[1]
            info.Engine = &engineName
            info.EngineVersion = &engineVersion
            break
        }
    }

    // Detect architecture
    archPatterns := map[string]string{
        "x86_64": `(x86_64|x64|amd64)`,
        "x86":    `(x86|i386|i686)`,
        "ARM":    `(arm|aarch64)`,
    }

    for arch, pattern := range archPatterns {
        if regexp.MustCompile(pattern).MatchString(ua) {
            architecture := arch
            info.Architecture = &architecture
            break
        }
    }

    return info
}

