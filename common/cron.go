package common

import "github.com/robfig/cron/v3"

var GCron *cron.Cron

func InitCron() {
	GCron = cron.New()
	GCron.Start()

}
func FinishCron() {
	GCron.Stop()
}
