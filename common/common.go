package common

import "time"

// 偏移时间用来gm调整时间
var OffSetTime = int64(0)

func GetNowTime() time.Time {
	return time.Now().Add(time.Duration(OffSetTime * int64(time.Second)))
}
