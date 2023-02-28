package resultlog

import (
	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
)

//UPIKafkaLogger extends the KafkaLogger to UPILogger interface
type UPIKafkaLogger struct {
	KafkaLogger
}

//WriteUPIRouterLog satisfy the UPILogger interface
func (l *KafkaLogger) WriteUPIRouterLog(routerLog *upiv1.RouterLog) error {
	return l.writeToKafka(
		routerLog,
		routerLog.PredictionId,
		routerLog.RequestTimestamp,
	)
}

func NewUPIKafkaLogger(cfg *config.KafkaConfig) (*UPIKafkaLogger, error) {
	kafkaLogger, err := NewKafkaLogger(cfg)
	if err != nil {
		return nil, err
	}
	return &UPIKafkaLogger{
		KafkaLogger: *kafkaLogger,
	}, nil
}
