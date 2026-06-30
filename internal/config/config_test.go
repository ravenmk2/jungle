package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadServer(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.toml")
	os.WriteFile(p, []byte(`[server]
addr = "127.0.0.1:7788"
[data]
dir = "./data"
[log]
level = "info"
`), 0o644)
	cfg, err := LoadServer(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Addr != "127.0.0.1:7788" {
		t.Fatalf("addr = %q", cfg.Server.Addr)
	}
	if cfg.Data.Dir != "./data" {
		t.Fatalf("data.dir = %q", cfg.Data.Dir)
	}
	if cfg.Log.Level != "info" {
		t.Fatalf("log.level = %q", cfg.Log.Level)
	}
}

func TestLoadWorkspace(t *testing.T) {
	p := filepath.Join(t.TempDir(), "demo.toml")
	os.WriteFile(p, []byte(`[java]
version = 8
home = "/jdk"
[maven]
home = "/mvn"
repo = "/repo"
[docs]
dirs = ["a", "b"]
[projects."demo"]
repo = "/proj"
[databases."db1"]
host = "127.0.0.1"
port = 3306
db = "demo"
user = "root"
password = "pw"
init-sql = "/init.sql"
[services."svc1"]
project = "demo"
module = "app"
work-dir = "/wd"
port = 8080
database = "db1"
[profiles]
items = ["dev", "staging"]
`), 0o644)
	cfg, err := LoadWorkspace(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Java.Version != 8 || cfg.Java.Home != "/jdk" {
		t.Fatalf("java = %+v", cfg.Java)
	}
	if cfg.Maven.Repo != "/repo" {
		t.Fatalf("maven.repo = %q", cfg.Maven.Repo)
	}
	if len(cfg.Docs.Dirs) != 2 || cfg.Docs.Dirs[1] != "b" {
		t.Fatalf("docs.dirs = %+v", cfg.Docs.Dirs)
	}
	if cfg.Projects["demo"].Repo != "/proj" {
		t.Fatalf("projects = %+v", cfg.Projects)
	}
	if cfg.Databases["db1"].InitSQL != "/init.sql" {
		t.Fatalf("db.init-sql = %q", cfg.Databases["db1"].InitSQL)
	}
	if cfg.Services["svc1"].Database != "db1" || cfg.Services["svc1"].WorkDir != "/wd" {
		t.Fatalf("service = %+v", cfg.Services["svc1"])
	}
	if len(cfg.Profiles.Items) != 2 {
		t.Fatalf("profiles = %+v", cfg.Profiles.Items)
	}
}
