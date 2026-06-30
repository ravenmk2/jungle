// Package config loads TOML configuration (server-level and per-workspace).
package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

// ServerConfig is the jungle binary's own config (config.toml).
type ServerConfig struct {
	Server ServerSection `toml:"server"`
	Data   DataSection   `toml:"data"`
	Log    LogSection    `toml:"log"`
}

type ServerSection struct {
	Addr string `toml:"addr"`
}

type DataSection struct {
	Dir string `toml:"dir"`
}

type LogSection struct {
	Level string `toml:"level"`
}

// WorkspaceConfig is a per-workspace config (config/workspaces/<name>.toml).
type WorkspaceConfig struct {
	Java      JavaSection         `toml:"java"`
	Maven     MavenSection        `toml:"maven"`
	Docs      DocsSection         `toml:"docs"`
	Projects  map[string]Project  `toml:"projects"`
	Databases map[string]Database `toml:"databases"`
	Services  map[string]Service  `toml:"services"`
	Profiles  ProfilesSection     `toml:"profiles"`
}

type JavaSection struct {
	Version int    `toml:"version"`
	Home    string `toml:"home"`
}

type MavenSection struct {
	Home string `toml:"home"`
	Repo string `toml:"repo"`
}

type DocsSection struct {
	Dirs []string `toml:"dirs"`
}

type Project struct {
	Repo string `toml:"repo"`
}

type Database struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	DB       string `toml:"db"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	InitSQL  string `toml:"init-sql"`
}

type Service struct {
	Project  string `toml:"project"`
	Module   string `toml:"module"`
	WorkDir  string `toml:"work-dir"`
	Port     int    `toml:"port"`
	Database string `toml:"database"`
}

type ProfilesSection struct {
	Items []string `toml:"items"`
}

// LoadServer reads and parses config.toml.
func LoadServer(path string) (*ServerConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ServerConfig
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadWorkspace reads and parses a workspace toml.
func LoadWorkspace(path string) (*WorkspaceConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg WorkspaceConfig
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
