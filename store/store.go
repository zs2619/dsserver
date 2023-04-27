package store

func InitStore() error {
	err := MongoInit()
	if err != nil {
		return err
	}
	err = RedisInit()
	if err != nil {
		return err
	}
	return nil
}

func FinishStore() error {
	MongoFinish()
	RedisFinish()
	return nil
}
