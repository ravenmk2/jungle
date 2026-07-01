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
	Java      JavaSection         `toml:"java" json:"java"`
	Maven     MavenSection        `toml:"maven" json:"maven"`
	Docs      DocsSection         `toml:"docs" json:"docs"`
	Projects  map[string]Project  `toml:"projects" json:"projects"`
	Databases map[string]Database `toml:"databases" json:"databases"`
	Services  map[string]Service  `toml:"services" json:"services"`
	Profiles  ProfilesSection     `toml:"profiles" json:"profiles"`
}

type JavaSection struct {
	Version int    `toml:"version" json:"version"`
	Home    string `toml:"home" json:"home"`
}

type MavenSection struct {
	Home string `toml:"home" json:"home"`
	Repo string `toml:"repo" json:"repo"`
}

type DocsSection struct {
	Dirs []string `toml:"dirs" json:"dirs"`
}

type Project struct {
	Repo string `toml:"repo" json:"repo"`
}

type Database struct {
	Host     string `toml:"host" json:"host"`
	Port     int    `toml:"port" json:"port"`
	DB       string `toml:"db" json:"db"`
	User     string `toml:"user" json:"user"`
	Password string `toml:"password" json:"-"`
	InitSQL  string `toml:"init-sql" json:"initSql,omitempty"`
}

type Service struct {
	Project  string `toml:"project" json:"project"`
	Module   string `toml:"module" json:"module"`
	WorkDir  string `toml:"work-dir" json:"workDir"`
	Port     int    `toml:"port" json:"port,omitempty"`
	Database string `toml:"database" json:"database,omitempty"`
}

type ProfilesSection struct {
	Items []string `toml:"items" json:"items"`
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
