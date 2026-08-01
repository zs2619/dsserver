package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	HeartbeatIntervalSec int64 `json:"heartbeatIntervalSec"`
}

type GameConfigType struct {
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
	return &GameConfig, nil
}
