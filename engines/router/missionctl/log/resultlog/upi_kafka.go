package resultlog

import (
	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
)

type UPIKafkaLogger struct {
	*KafkaLogger
}

func NewUPIKafkaLogger(cfg *config.KafkaConfig) (*UPIKafkaLogger, error) {
	kafkaLogger, err := NewKafkaLogger(cfg)
	if err != nil {
		return nil, err
	}
	return &UPIKafkaLogger{kafkaLogger}, nil
}

func (l *UPIKafkaLogger) write(routerLog *upiv1.RouterLog) error {
	return l.writeToKafka(
		routerLog,
		routerLog.PredictionId,
		routerLog.RequestTimestamp,
	)
}
