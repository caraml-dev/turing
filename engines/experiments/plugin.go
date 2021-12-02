package experiments

import "github.com/hashicorp/go-plugin"

const (
	ManagerPluginIdentifier = "manager"
	RunnerPluginIdentifier  = "runner"
)

var (
	HandshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "experiments",
		MagicCookieValue: "experiments",
	}
)
