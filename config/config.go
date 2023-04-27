package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type GameConfigType struct {
	DSPath    string `json:"dsPath"`
	DSArgv    string `json:"dsArgv"`
	DsWorkDir string `json:"dsWorkDir"`
	IP        string `json:"dsIP"`
}

var GameConfig GameConfigType

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
	return &GameConfig, nil
}
