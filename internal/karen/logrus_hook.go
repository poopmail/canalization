package karen

import (
	"github.com/go-redis/redis/v8"
	"github.com/poopmail/canalization/internal/static"
	"github.com/sirupsen/logrus"
	"github.com/ztrue/tracerr"
)

// levelMap maps logrus log levels to karen message types
var levelMap = map[logrus.Level]MessageType{
	logrus.WarnLevel:  MessageTypeWarning,
	logrus.ErrorLevel: MessageTypeError,
	logrus.FatalLevel: MessageTypePanic,
	logrus.PanicLevel: MessageTypePanic,
}

// LogrusHook represents the logrus hook that notifies karen about any important log entry
type LogrusHook struct {
	Redis *redis.Client
}

// Levels returns the levels the karen logrus hook should apply to
func (hook *LogrusHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

// Fire sends a message to karen with the data of the log entry set
func (hook *LogrusHook) Fire(entry *logrus.Entry) error {
	msg := entry.Message
	if err, ok := entry.Data[logrus.ErrorKey].(error); ok {
		msg = tracerr.Sprint(tracerr.Wrap(err))
	}

	return Send(hook.Redis, Message{
		Type:        levelMap[entry.Level],
		Service:     static.KarenServiceName,
		Topic:       "",
		Description: msg,
	})
}
