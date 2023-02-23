package resultlog

import (
	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

type UPILogger struct {
	*KafkaLogger
}

func newUPILogger(cfg *config.KafkaConfig) (*UPILogger, error) {
	kafkaLogger, err := newKafkaLogger(cfg)
	if err != nil {
		return nil, err
	}
	return &UPILogger{
		kafkaLogger,
	}, nil
}

// MarshalJSON implement custom Marshaling for TuringResultLogEntry, using the underlying proto def
//func (l *UPILogger) writeUPILog(log *upiv1.RouterLog) ([]byte, error) {
//	return protojson.MarshalOptions{
//		UseProtoNames: true, // Use the json field name instead of the camel case struct field name
//	}.Marshal(log)
//}
