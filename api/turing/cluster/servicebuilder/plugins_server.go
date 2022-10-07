package servicebuilder

import (
	v1 "k8s.io/api/core/v1"
)

const (
	envPluginName = "PLUGIN_NAME"
	envPluginsDir = "PLUGINS_DIR"
)

var (
	pluginsMountPath = "/app/plugins"
	pluginsVolume    = v1.Volume{
		Name: "plugins-volume",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
)
