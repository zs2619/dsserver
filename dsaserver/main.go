package main

import (
	"dsservices/config"
	"dsservices/dsaserver/dsa"
	"dsservices/dsaserver/proc"
	"github.com/zs2619/kissnet-go"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var g errgroup.Group

func main() {
	logrus.Info("dsa server start")
	_, err := config.LoadGameConfig("assets/config.json")
	if err != nil {
		logrus.WithError(err).Error("config.LoadGameConfig error")
		return
	}
	logrus.Info("config.LoadGameConfig OK")

	// 端口分配范围（container 模式由 env GAME_PORT_MIN/MAX 注入本 Deployment 段）。
	proc.GPortMgr.Init(config.GameConfig.Dsa.GamePortMin, config.GameConfig.Dsa.GamePortMax)

	// 玩家可达 IP：container 模式由 Downward API 注入 status.hostIP；process 模式回退 config dsIP。
	advertiseIP := os.Getenv("DS_ADVERTISE_IP")
	if advertiseIP == "" {
		advertiseIP = config.GameConfig.IP
	}
	if advertiseIP == "" {
		advertiseIP = "127.0.0.1"
	}
	dsa.GAdvertiseIP = advertiseIP
	logrus.WithField("advertiseIP", advertiseIP).Info("advertise IP set")

	grpcURL := os.Getenv("DS_DSC_GRPC_URI")
	if len(grpcURL) == 0 {
		logrus.Error("DS_DSC_GRPC_URI empty")
		return
	}

	port, err := strconv.Atoi(os.Getenv("DS_DSA_PORT"))
	if err != nil {
		logrus.WithError(err).Error("DS_DSA_PORT error")
		return
	}
	agentID := os.Getenv("DS_DSA_AgentID")
	if len(agentID) == 0 {
		logrus.Error("DS_DSA_AgentID empty")
		return
	}

	grpcPort, err := strconv.Atoi(os.Getenv("DS_DSA_GRPC_PORT"))
	if err != nil {
		logrus.WithError(err).Error("DS_DSA_GRPC_PORT error")
		return
	}
	advertise := os.Getenv("DS_DSA_GRPC_ADDR")
	if advertise == "" {
		advertise = "127.0.0.1:" + strconv.Itoa(grpcPort)
	}

	reportClient, err := dsa.NewReportClient(agentID, grpcURL, advertise)
	if err != nil {
		logrus.WithError(err).Error("NewReportClient error")
		return
	}

	agentServer := dsa.NewAgentServer(grpcPort)

	event := kissnet.NewNetEvent()

	var shutdownOnce sync.Once
	var gTemplateAcceptor kissnet.Acceptor
	var gAcceptor kissnet.Acceptor

	shutdown := func() {
		shutdownOnce.Do(func() {
			agentServer.Stop()
			dsa.GDSInfoMgr.ShutdownAll()
			if gAcceptor != nil {
				_ = gAcceptor.Close()
			}
			if gTemplateAcceptor != nil {
				_ = gTemplateAcceptor.Close()
			}
			if gAcceptor != nil && !gAcceptor.WaitConnectionsClosed(10*time.Second) {
				logrus.Warn("not all connections closed within grace period")
			}
			event.Close()
			reportClient.Close()
		})
	}

	logrus.Info("ds acceptor start")
	gAcceptor, err = kissnet.NewAcceptor(
		"tcp",
		port,
		&dsa.GDSAServerCB,
		event,
	)
	if err != nil {
		logrus.WithError(err).Fatal("NewAcceptor error")
		return
	}

	// 模板口独立 listener（模板进程连上后发 TemplateReady，DSA 据此 fork 实例）。
	if tp := config.GameConfig.Dsa.TemplatePort; tp > 0 {
		gTemplateAcceptor, err = kissnet.NewAcceptor("tcp", tp, &dsa.GTemplateCB, event)
		if err != nil {
			logrus.WithError(err).Fatal("template NewAcceptor error")
			return
		}
		logrus.WithField("templatePort", tp).Info("template acceptor start")
	}

	// 先起 EventLoop / Acceptor，再等模板就绪，最后 Register（避免被调度却无法 fork）。
	g.Go(func() error {
		err := gAcceptor.Run()
		if err != nil {
			logrus.WithError(err).Error("acceptor run failed, shutting down")
			shutdown()
		}
		return err
	})
	if gTemplateAcceptor != nil {
		g.Go(func() error {
			err := gTemplateAcceptor.Run()
			if err != nil {
				logrus.WithError(err).Error("template acceptor run failed, shutting down")
				shutdown()
			}
			return err
		})
	}
	g.Go(func() error {
		event.EventLoop()
		return nil
	})
	g.Go(func() error {
		return agentServer.Run()
	})

	if config.GameConfig.Dsa.TemplateReadyRequired && gTemplateAcceptor != nil {
		logrus.Info("waiting template ready before Register")
		waitTemplateReady(60 * time.Second)
		logrus.Info("template ready, registering")
	}

	if err := reportClient.Register(); err != nil {
		logrus.WithError(err).Error("RegisterAgent error")
		shutdown()
		return
	}

	// 本地 creating 超时回收：fork 后始终未 CreateOK 的 DS（实例死/连不上且未报 ForkFailed），
	// 在 EventLoop 上串行执行（与 dsCreateOK 无竞态），释放 PortMgr 口并上报 DEAD。
	createTimeout := config.GameConfig.Dsc.DsCreatingTimeoutSec
	if createTimeout > 0 {
		event.ScheduleTimer(createTimeout*1000, func() error {
			dsa.GDSInfoMgr.SweepStaleCreating(createTimeout)
			return nil
		})
	}

	g.Go(func() error {
		reportClient.RunHeartbeat()
		return nil
	})

	// 信号触发优雅退出
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		logrus.Info("shutdown signal received")
		shutdown()
	}()

	err = g.Wait()
	// 兜底清理（幂等）：信号/accept 错误路径可能已执行过
	shutdown()
	if err != nil {
		logrus.WithError(errors.WithStack(err)).Fatal("g.wait")
	}
	logrus.Info("dsa server end")
}

// waitTemplateReady 轮询模板就绪状态，最多等待 timeout。
func waitTemplateReady(timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if dsa.GTemplateClient.IsReady() {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	logrus.Warn("template not ready within timeout, registering anyway")
}
