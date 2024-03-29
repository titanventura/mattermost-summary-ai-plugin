package main

import (
	"sync"

	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

type Plugin struct {
	plugin.MattermostPlugin
	client            *pluginapi.Client
	configurationLock sync.RWMutex
	configuration     *configuration
	botID             string
}
