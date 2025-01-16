package dsc

import (
	"dsservices/pb"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type RouterMapNatsHandle func(*DSCNatsConn, proto.Message) error

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		logrus.Printf("Disconnected due to: %s, will attempt reconnects for %.0fm", err, totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		logrus.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		logrus.Fatalf("Exiting: %v", nc.LastError())
	}))
	return opts
}

type DSCNatsConn struct {
	NatsConn   *nats.Conn
	dscName    string
	agentName  string
	dscSubject string
	run        atomic.Bool
	cb         RouterMapNatsHandle
}

func NewDSCNatsClient(natsURl, dscName, agentName string) (err error, nc *DSCNatsConn) {
	opts := []nats.Option{nats.Name("NewDSCNatsClient")}
	opts = setupConnOptions(opts)
	nc = &DSCNatsConn{
		dscName:   dscName,
		agentName: agentName,
	}
	nc.NatsConn, err = nats.Connect(natsURl, opts...)
	if err != nil {
		return
	}
	nc.dscSubject = "dsc"
	go nc.natDSAWoker()
	return
}

func (nc *DSCNatsConn) natDSAWoker() error {
	for nc.run.Load() {
		nc.NatsConn.QueueSubscribe(nc.dscSubject, "dsa", func(msg *nats.Msg) {
			var natsEvnet pb.NatsEvent
			err := proto.Unmarshal(msg.Data, &natsEvnet)
			if err != nil {
				return
			}
			if nc.cb != nil {
				nc.cb(nc, natsEvnet.NatsEventMsg)
			}
		})
	}
	return nil
}

func (nc *DSCNatsConn) PubDSANatsEvent(agentName string, msg []byte) error {
	return nil
}
