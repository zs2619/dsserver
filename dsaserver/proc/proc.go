package proc

import (
	"dsservices/config"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type ProcInfo struct {
	RealmCfgID string
	OSProcess  *os.Process
	Ip         string
	Port       int
}

func formatDSArgv(cmdName, dsID, realmCfgID string, port int) []string {
	argv := []string{cmdName}
	gameModeName := "ServerGameMode"
	serverMapLayerName := "FuBen"
	playerStartName := "PlayerStart_Server_MLFB"
	portArgv := fmt.Sprintf(config.GameConfig.DSArgv, dsID, port, gameModeName, serverMapLayerName, playerStartName)
	logrus.Info(argv)
	argvList := strings.Fields(strings.TrimSpace(portArgv))
	argv = append(argv, argvList...)
	return argv
}

func StartProc(dsID, realmCfgID string) (*ProcInfo, error) {
	port := GPortMgr.GetValidPort()
	cmdName := config.GameConfig.DSPath
	argv := formatDSArgv(cmdName, dsID, realmCfgID, port)

	procAttr := &os.ProcAttr{
		Dir:   config.GameConfig.DsWorkDir,
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}
	osProcess, err := os.StartProcess(cmdName, argv, procAttr)
	if err != nil {
		return nil, err
	}

	procInfo := &ProcInfo{
		Port:       port,
		Ip:         config.GameConfig.IP,
		RealmCfgID: realmCfgID,
		OSProcess:  osProcess}

	return procInfo, nil
}
