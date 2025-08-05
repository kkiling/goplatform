package config

type Arguments struct {
	ConfigFilename string `short:"c" long:"cfg" description:"A path to cfg file" required:"true" default:"configs/config.yml"`
}
