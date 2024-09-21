package proc

import (
	"dsservices/config"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
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
	portArgv := fmt.Sprintf(config.GameConfig.DSArgv, dsID, port, realmCfgID)
	logrus.WithFields(logrus.Fields{"argv": argv}).Info("formatDSArgv")
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

func (proc *ProcInfo) KillProc() error {
	defer func() {
		processState, err := proc.OSProcess.Wait()
		if err != nil {
			logrus.WithError(errors.WithStack(err)).Error("KillProc error")
		} else {
			logrus.WithFields(logrus.Fields{
				"pid":          proc.OSProcess.Pid,
				"processState": processState,
			}).Info("KillProc ok")
		}
	}()
	return proc.OSProcess.Kill()
}
