package resultlog

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
)

type UPILogger struct {
	*KafkaLogger
	// This corresponds to the name and version of the router deployed from the Turing app
	// Format: {router_name}-{router_version}.{project_name}
	routerName    string
	routerVersion string
	projectName   string
}

var routerRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+-\d+.[a-zA-Z0-9\-]+$`)

func newUPILogger(cfg *config.KafkaConfig, appName string) (*UPILogger, error) {

	if !routerRegex.MatchString(appName) {
		return nil, fmt.Errorf("invalid router name ")
	}
	s := strings.Split(appName, ".")
	routerNameWithVersion := s[0]
	projectName := s[1]
	i := strings.LastIndex(routerNameWithVersion, "-")
	routerName := routerNameWithVersion[:i]
	// do not include '-'
	routerVersion := routerNameWithVersion[i+1:]

	kafkaLogger, err := newKafkaLogger(cfg)
	if err != nil {
		return nil, err
	}
	return &UPILogger{
		KafkaLogger:   kafkaLogger,
		routerName:    routerName,
		routerVersion: routerVersion,
		projectName:   projectName,
	}, nil
}

//writeUPILog implement custom Marshaling for TuringResultLogEntry, using the underlying proto def
func (l *UPILogger) writeUPILog(routerLog *upiv1.RouterLog) error {

	routerLog.RouterName = l.routerName
	routerLog.RouterVersion = l.routerVersion
	routerLog.ProjectName = l.projectName

	return l.writeToKafka(
		routerLog,
		routerLog.PredictionId,
		routerLog.RequestTimestamp,
	)
}
