package alertschema

// Alert is a generic representation of an alert that can be used to configure alert managers
// such as Prometheus Alertmanager (not implemented currently).
//
// Currently this schema is used to create a yaml file that will be committed to a Git repo
// for GitOps based alert management. The responsibility to register and activate these alerts
// is left to the continuous integration assumed to have been setup in the Git repo.
type Alert struct {
	AlertName   string            `yaml:"alert"          validate:"required"`
	Expression  string            `yaml:"expr"           validate:"required"             hash:"ignore"`
	For         string            `yaml:"for,omitempty"                                  hash:"ignore"`
	Labels      map[string]string `yaml:"labels"         validate:"required,labels"`
	Annotations map[string]string `yaml:"annotations"    validate:"required,annotations" hash:"ignore"`
}

// AlertGroup contains alerts grouped by name.
type AlertGroup struct {
	Name   string  `yaml:"name"`
	Alerts []Alert `yaml:"rules"`
}
