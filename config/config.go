package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

type TeamConfig struct {
	MatchTimeoutMs int64 `json:"matchTimeoutMs"`
}

type RealmConfig struct {
	DisconnectGraceSec int64 `json:"disconnectGraceSec"`
}

type DscConfig struct {
	OrphanScanIntervalSec    int64 `json:"orphanScanIntervalSec"`
	DsCreatingTimeoutSec     int64 `json:"dsCreatingTimeoutSec"`
	DsReadyIdleTimeoutSec    int64 `json:"dsReadyIdleTimeoutSec"`
	DsPlayingIdleTimeoutSec  int64 `json:"dsPlayingIdleTimeoutSec"`
	AgentHeartbeatTimeoutSec int64 `json:"agentHeartbeatTimeoutSec"`
}

type DsaConfig struct {
	HeartbeatIntervalSec  int64  `json:"heartbeatIntervalSec"`
	Capacity              int32  `json:"capacity"`
	TemplatePort          int    `json:"templatePort"` // DSA 监听模板连接的口（模板 fork 口）
	TemplateReadyRequired bool   `json:"templateReadyRequired"`
	GamePortMin           int    `json:"gamePortMin"`
	GamePortMax           int    `json:"gamePortMax"`
}

type GameConfigType struct {
	DSRuntime string      `json:"dsRuntime"` // process|container
	DSPath    string      `json:"dsPath"`
	DSArgv    string      `json:"dsArgv"`
	DsWorkDir string      `json:"dsWorkDir"`
	IP        string      `json:"dsIP"`
	Team      TeamConfig  `json:"team"`
	Realm     RealmConfig `json:"realm"`
	Dsc       DscConfig   `json:"dsc"`
	Dsa       DsaConfig   `json:"dsa"`
}

var GameConfig GameConfigType

func applyDefaults(cfg *GameConfigType) {
	if cfg.DSRuntime == "" {
		cfg.DSRuntime = "process"
	}
	if cfg.Team.MatchTimeoutMs <= 0 {
		cfg.Team.MatchTimeoutMs = 10000
	}
	if cfg.Realm.DisconnectGraceSec <= 0 {
		cfg.Realm.DisconnectGraceSec = 60
	}
	if cfg.Dsc.OrphanScanIntervalSec <= 0 {
		cfg.Dsc.OrphanScanIntervalSec = 30
	}
	if cfg.Dsc.DsCreatingTimeoutSec <= 0 {
		cfg.Dsc.DsCreatingTimeoutSec = 60
	}
	if cfg.Dsc.DsReadyIdleTimeoutSec <= 0 {
		cfg.Dsc.DsReadyIdleTimeoutSec = 120
	}
	if cfg.Dsc.DsPlayingIdleTimeoutSec <= 0 {
		cfg.Dsc.DsPlayingIdleTimeoutSec = 3600
	}
	if cfg.Dsc.AgentHeartbeatTimeoutSec <= 0 {
		cfg.Dsc.AgentHeartbeatTimeoutSec = 15
	}
	if cfg.Dsa.HeartbeatIntervalSec <= 0 {
		cfg.Dsa.HeartbeatIntervalSec = 3
	}
	if cfg.Dsa.Capacity <= 0 {
		cfg.Dsa.Capacity = 100
	}
	if cfg.Dsa.GamePortMin <= 0 {
		cfg.Dsa.GamePortMin = 30000
	}
	if cfg.Dsa.GamePortMax <= 0 {
		cfg.Dsa.GamePortMax = 60000
	}
}

func getConfigRaw(config string) ([]byte, error) {
	configPath := os.Getenv("DS_CONFIGPATH")
	filePath := filepath.Join(configPath, config)
	raw, err := os.ReadFile(filePath)
	return raw, err
}

func LoadGameConfig(config string) (*GameConfigType, error) {
	rawFile, err := getConfigRaw(config)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(rawFile, &GameConfig)
	if err != nil {
		return nil, err
	}
	applyDefaults(&GameConfig)
	applyEnvOverrides(&GameConfig)
	return &GameConfig, nil
}

// applyEnvOverrides: env 覆盖配置（container 模式下由 Downward API 注入）。
func applyEnvOverrides(cfg *GameConfigType) {
	if v := os.Getenv("DS_DSA_RUNTIME"); v != "" {
		cfg.DSRuntime = v
	}
	if v := os.Getenv("DSA_TEMPLATE_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Dsa.TemplatePort = n
		}
	}
	if v := os.Getenv("TEMPLATE_READY_REQUIRED"); v == "true" || v == "1" {
		cfg.Dsa.TemplateReadyRequired = true
	}
	if v := os.Getenv("GAME_PORT_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Dsa.GamePortMin = n
		}
	}
	if v := os.Getenv("GAME_PORT_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Dsa.GamePortMax = n
		}
	}
}
