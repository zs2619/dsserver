package dsa

import (
	"dsservices/pb"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type RouterMapNatsHandle func(*DSANatsConn, proto.Message) error

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

type DSANatsConn struct {
	NatsConn   *nats.Conn
	dscName    string
	agentName  string
	dsaSubject string
	dscSubject string
	run        atomic.Bool
	cb         RouterMapNatsHandle
}

func NewDSANatsClient(natsURl, dscName, agentName string) (err error, nc *DSANatsConn) {
	opts := []nats.Option{nats.Name("NewDSANatsClient")}
	opts = setupConnOptions(opts)
	nc = &DSANatsConn{
		dscName:   dscName,
		agentName: agentName,
	}
	nc.NatsConn, err = nats.Connect(natsURl, opts...)
	if err != nil {
		return
	}
	err = nc.initDscNats()
	if err != nil {
		return
	}
	nc.dsaSubject = "dsa." + agentName
	nc.dscSubject = "dsc"
	go nc.natWoker()
	return
}
func (nc *DSANatsConn) initDscNats() error {
	hello := &pb.DSADSC_Hello{AgentName: nc.agentName}
	requestMsg, _ := proto.Marshal(hello)
	_, err := nc.NatsConn.Request(nc.dscName, requestMsg, 5*time.Second)
	if err != nil {
		return err
	}
	return nil
}

func (nc *DSANatsConn) natWoker() error {
	for nc.run.Load() {
		nc.NatsConn.Subscribe(nc.dsaSubject, func(msg *nats.Msg) {
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

func (nc *DSANatsConn) PubDscNatsEvent(dscMsg []byte) error {
	return nc.NatsConn.Publish(nc.dscSubject, dscMsg)
}
