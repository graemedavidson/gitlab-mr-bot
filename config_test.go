package main

import (
	// "fmt"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// Setup

var (
	mockFS afero.Fs = afero.NewMemMapFs()
)

func init() {
	emptyConfig := ""
	noChannelsConfig := "---\ngroup_channels:"
	passConfig := `---
group_channels:
  chan1:
    slack_channel: "#chan1"
    slack_channel_id: "AAAAAAAA"
  chan2:
    slack_channel: "#chan2"
    slack_channel_id: "BBBBBBBB"
`

	err := afero.WriteFile(mockFS, "empty.yaml", []byte(emptyConfig), os.ModePerm)
	if err != nil {
		fmt.Println("error setting up mock filesystem.")
	}
	err = afero.WriteFile(mockFS, "no-channels.yaml", []byte(noChannelsConfig), os.ModePerm)
	if err != nil {
		fmt.Println("error setting up mock filesystem.")
	}
	err = afero.WriteFile(mockFS, "pass.yaml", []byte(passConfig), os.ModePerm)
	if err != nil {
		fmt.Println("error setting up mock filesystem.")
	}
}

type MockFS struct {
}

func (o *MockFS) Open(name string) (file, error) {
	file, err := mockFS.Open(name)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (o *MockFS) Stat(name string) (os.FileInfo, error) {
	stat, err := mockFS.Stat(name)
	if err != nil {
		return nil, err
	}
	return stat, nil
}

// Tests

func TestLoadConfig(t *testing.T) {
	mockFS := &MockFS{}

	type test struct {
		path              string
		wantNumOfChannels int
		wantChannels      map[string]GroupChannel
		err               error
	}

	tests := []test{
		{
			path:              "pass.yaml",
			wantNumOfChannels: 2,
			wantChannels: map[string]GroupChannel{
				"chan1": {SlackChannel: "#chan1", SlackChannelID: "AAAAAAAA"},
				"chan2": {SlackChannel: "#chan2", SlackChannelID: "BBBBBBBB"},
			},
			err: nil,
		},
		{
			path:              "empty.yaml",
			wantNumOfChannels: 0,
			wantChannels:      nil,
			err:               errors.New("'empty.yaml' configuration file is empty."),
		},
		{
			path:              "no-channels.yaml",
			wantNumOfChannels: 0,
			wantChannels:      nil,
			err:               nil,
		},
		{
			path:              "fail.yaml",
			wantNumOfChannels: 0,
			wantChannels:      nil,
			err:               errors.New("open fail.yaml: file does not exist"),
		},
	}

	for _, tc := range tests {
		c := &Config{ConfigPath: tc.path}
		err := c.LoadConfig(mockFS)
		if err != nil {
			assert.Equal(t, err.Error(), tc.err.Error())
		} else {
			assert.Equal(t, len(c.GroupChannels), tc.wantNumOfChannels)
			assert.Equal(t, c.GroupChannels, tc.wantChannels)
		}
	}
}

func TestValidateConfigPath(t *testing.T) {
	mockFS := &MockFS{}

	type test struct {
		path string
		err  string
	}

	tests := []test{
		{path: "pass.yaml", err: ""},
		{path: "empty.yaml", err: "'empty.yaml' configuration file is empty."},
		{path: "no-channels.yaml", err: ""},
		{path: "fail.yaml", err: "open fail.yaml: file does not exist"},
	}

	for _, tc := range tests {
		c := &Config{ConfigPath: tc.path}
		err := c.LoadConfig(mockFS)
		if err != nil {
			assert.Equal(t, err.Error(), tc.err)
		} else {
			assert.Nil(t, err)
		}
	}
}
