package main

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// Config - Base Workable Bot Configuration
type Config struct {
	ConfigPath    string                  `yaml:"-"`
	GroupChannels map[string]GroupChannel `yaml:"group_channels"`
	UserStatuses  map[string]int          `yaml:"user_statuses"`
}

type GroupChannel struct {
	SlackChannel   string `yaml:"slack_channel"`
	SlackChannelID string `yaml:"slack_channel_id"`
}

func (c *Config) LoadConfig(fs fileSystem) error {
	err := c.ValidateConfigPath(fs)
	if err != nil {
		return err
	}

	file, err := fs.Open(c.ConfigPath)
	if err != nil {
		return err
	}
	defer file.Close()

	decodedFile := yaml.NewDecoder(file)
	if err := decodedFile.Decode(c); err != nil {
		return err
	}

	return nil
}

func (c *Config) ValidateConfigPath(fs fileSystem) error {
	s, err := fs.Stat(c.ConfigPath)

	if err != nil {
		return err
	}
	if s.Size() == 0 {
		return fmt.Errorf("'%s' configuration file is empty.", c.ConfigPath)
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file.", c.ConfigPath)
	}
	return nil
}
