package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/ini.v1"
)

// Config holds all configuration settings
type Config struct {
	Monitor MonitorConfig
}

// MonitorConfig holds monitoring settings
type MonitorConfig struct {
	WatchPath string
	Interval  float64
	Pattern   string
}

// LoadConfig loads configuration from config.ini
func LoadConfig() (*Config, error) {
	configPath := getConfigPath()
	if configPath == "" {
		return nil, fmt.Errorf("config.ini not found")
	}

	cfg, err := ini.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	config := &Config{}

	// Load Monitor section
	monitorSection := cfg.Section("monitor")
	config.Monitor.WatchPath = monitorSection.Key("watch_path").String()
	config.Monitor.Pattern = monitorSection.Key("pattern").String()

	if config.Monitor.WatchPath == "" {
		return nil, fmt.Errorf("monitor.watch_path is required")
	}

	if config.Monitor.Pattern == "" {
		config.Monitor.Pattern = "clipboard_*.tar.gz"
	}

	intervalStr := monitorSection.Key("interval").String()
	if intervalStr == "" {
		config.Monitor.Interval = 1.0
	} else {
		interval, err := strconv.ParseFloat(intervalStr, 64)
		if err != nil {
			config.Monitor.Interval = 1.0
		} else {
			config.Monitor.Interval = interval
		}
	}

	return config, nil
}

// getConfigPath returns the path to config.ini
func getConfigPath() string {
	// First, try executable directory
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		configPath := filepath.Join(execDir, "config.ini")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Then, try current directory
	if _, err := os.Stat("config.ini"); err == nil {
		return "config.ini"
	}

	return ""
}

// SaveConfig saves configuration to config.ini
func SaveConfig(config *Config) error {
	configPath := getConfigPath()
	if configPath == "" {
		// Create in executable directory
		execPath, err := os.Executable()
		if err != nil {
			configPath = "config.ini"
		} else {
			configPath = filepath.Join(filepath.Dir(execPath), "config.ini")
		}
	}

	cfg := ini.Empty()

	monitorSection, _ := cfg.NewSection("monitor")
	monitorSection.NewKey("watch_path", config.Monitor.WatchPath)
	monitorSection.NewKey("interval", fmt.Sprintf("%.1f", config.Monitor.Interval))
	monitorSection.NewKey("pattern", config.Monitor.Pattern)

	return cfg.SaveTo(configPath)
}
