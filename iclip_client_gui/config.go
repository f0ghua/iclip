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
	SMB     SMBConfig
	Monitor MonitorConfig
}

// SMBConfig holds SMB connection settings
type SMBConfig struct {
	Server    string
	Share     string
	Path      string
	Username  string
	Password  string
	Workgroup string
}

// MonitorConfig holds monitoring settings
type MonitorConfig struct {
	Interval float64
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

	// Load SMB section
	smbSection := cfg.Section("smb")
	config.SMB.Server = smbSection.Key("server").String()
	config.SMB.Share = smbSection.Key("share").String()
	config.SMB.Path = smbSection.Key("path").String()
	config.SMB.Username = smbSection.Key("username").String()
	config.SMB.Password = smbSection.Key("password").String()
	config.SMB.Workgroup = smbSection.Key("workgroup").String()

	// Validate required fields
	if config.SMB.Server == "" {
		return nil, fmt.Errorf("smb.server is required")
	}
	if config.SMB.Share == "" {
		return nil, fmt.Errorf("smb.share is required")
	}
	if config.SMB.Username == "" {
		return nil, fmt.Errorf("smb.username is required")
	}
	if config.SMB.Password == "" {
		return nil, fmt.Errorf("smb.password is required")
	}

	// Load Monitor section
	monitorSection := cfg.Section("monitor")
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

	smbSection, _ := cfg.NewSection("smb")
	smbSection.NewKey("server", config.SMB.Server)
	smbSection.NewKey("share", config.SMB.Share)
	smbSection.NewKey("path", config.SMB.Path)
	smbSection.NewKey("username", config.SMB.Username)
	smbSection.NewKey("password", config.SMB.Password)
	if config.SMB.Workgroup != "" {
		smbSection.NewKey("workgroup", config.SMB.Workgroup)
	}

	monitorSection, _ := cfg.NewSection("monitor")
	monitorSection.NewKey("interval", fmt.Sprintf("%.1f", config.Monitor.Interval))

	return cfg.SaveTo(configPath)
}
