package dsa

import (
	"context"
	"dsservices/common"
	"dsservices/pb"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var GDSAClient *DSAClient

type RouterMapHandle func(*DSAClient, proto.Message) error

type DSAClient struct {
	agentID               string
	ctx                   context.Context
	ctxCancel             context.CancelFunc // Cancelled on close.
	client                pb.StreamDscDSAClient
	streamClientEventChan chan *pb.StreamClientEvent
	stream                pb.StreamDscDSA_StreamServiceClient
	grpcConn              *grpc.ClientConn
	quit                  atomic.Bool
	cb                    RouterMapHandle
}

func NewDSAClient(agentID, addr string) (agentClient *DSAClient, err error) {
	var conn *grpc.ClientConn
	var opts []grpc.DialOption
	for {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		conn, err = grpc.DialContext(ctx, addr, opts...)
		if err != nil {
			logrus.WithError(errors.WithStack(err)).WithFields(logrus.Fields{
				"addr": addr,
			}).Error("grpc.Dial error")
		} else {
			logrus.Info("grpc.Dial ok")
			break
		}
	}
	agentClient = &DSAClient{
		streamClientEventChan: make(chan *pb.StreamClientEvent, 1024),
		agentID:               agentID,
		grpcConn:              conn,
	}
	agentClient.client = pb.NewStreamDscDSAClient(conn)
	agentClient.quit.Store(false)
	return
}

func (agent *DSAClient) Send2DSC(resp proto.Message) (err error) {
	if !agent.quit.Load() {
		clientEvent := &pb.StreamClientEvent{}
		clientEvent.CEvent, err = anypb.New(resp)
		if err != nil {
			return
		}
		logrus.WithFields(logrus.Fields{"cevent": clientEvent.CEvent, "resp": resp}).Info("Send2DSC")
		agent.streamClientEventChan <- clientEvent
	}
	return
}
func (agent *DSAClient) Close() error {
	logrus.WithFields(logrus.Fields{}).Info("AgentClient::Close")
	agent.quit.Store(true)
	agent.ctxCancel()
	agent.grpcConn.Close()
	return nil
}

func (agent *DSAClient) RunStreamService() error {
	logrus.WithFields(logrus.Fields{}).Info("RunStreamService start")
	var err error
	md := metadata.Pairs(common.MD_KEY_AGENTID, agent.agentID)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	retry := false
	for {
		if agent.quit.Load() {
			break
		}
		for {
			if agent.quit.Load() {
				return nil
			}
			if retry {
				time.Sleep(3 * time.Second)
				retry = false
			}
			agent.ctx, agent.ctxCancel = context.WithCancel(ctx)
			agent.stream, err = agent.client.StreamService(agent.ctx)
			if err != nil {
				logrus.WithError(errors.WithStack(err)).Error("runStreamService:client.StreamService error")
				retry = true
			} else {
				logrus.WithFields(logrus.Fields{}).Info("runStreamService:client.StreamService ok")
				break
			}
		}
		recvQuit := make(chan struct{})
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer func() {
				wg.Done()
			}()
			go func() {
				for req := range agent.streamClientEventChan {
					if err := agent.stream.Send(req); err != nil {
						logrus.WithError(errors.WithStack(err)).Error("runStreamService:client.StreamServiceReqChan send error")
						agent.stream.CloseSend()
						return
					}
				}
			}()

			go func() {
				select {
				case <-recvQuit:
					agent.stream.CloseSend()
					return
				}
			}()
		}()

		go func() {
			defer func() {
				wg.Done()
			}()
			for {
				recvEvent, err := agent.stream.Recv()
				if err != nil {
					logrus.WithError(errors.WithStack(err)).Error("runStreamService:client recv error")
					recvQuit <- struct{}{}
					return
				}

				pmsg, err := recvEvent.SEvent.UnmarshalNew()
				if err != nil {
					logrus.WithError(errors.WithStack(err)).Error("recvEvent.PEvent.UnmarshalNew error")
					continue
				}
				if agent.cb != nil {
					err = agent.cb(agent, pmsg)
					if err != nil {
						logrus.WithError(errors.WithStack(err)).Error("agent RouterMap error")
					}
				}
			}
		}()
		wg.Wait()
		retry = true
	}
	close(agent.streamClientEventChan)
	logrus.WithFields(logrus.Fields{}).Info("runStreamService quit")
	return nil
}
