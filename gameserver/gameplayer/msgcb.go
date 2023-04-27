package gameplayer

import (
	"encoding/binary"
	"fmt"

	"dsservices/kissnet"
	"dsservices/pb"
)

func GameServerCB(conn kissnet.IConnection, msg []byte) error {
	if msg == nil {
		GPlayerMgr.DelPlayer(conn)
		return nil
	}
	if len(msg) < 2 {
		GPlayerMgr.DelPlayer(conn)
		return fmt.Errorf("game msg len error")
	}
	msgID := pb.C2S_MsgID_MsgID(binary.LittleEndian.Uint16(msg))
	err := ProcGameMsg(conn, msgID, msg[2:])
	if err != nil {
		return nil
	}
	return nil
}

func ProcGameMsg(conn kissnet.IConnection, msgID pb.C2S_MsgID_MsgID, msg []byte) error {
	if msgID == pb.C2S_MsgID_LocalLogin {
		return LocalLogin(conn, msgID, msg)
	} else if msgID == pb.C2S_MsgID_AuthLogin {
		return AuthLogin(conn, msgID, msg)
	} else {
		f, ok := msgMap[msgID]
		if !ok {
			return fmt.Errorf("(%d) msgID nil", msgID)
		}
		u := GPlayerMgr.GetPlayerByClientID(conn)
		if u == nil {
			return fmt.Errorf("(%d) ProcMsg player nil", msgID)
		}
		return f(u, msg)
	}

}
