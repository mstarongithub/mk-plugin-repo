package log

type LogrusGormWrapper struct {
}

var log *LogrusGormWrapper

func init() {
	log = &LogrusGormWrapper{}
}
