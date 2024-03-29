package main

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

type configuration struct {
	GeminiAPIKey string
}

func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}

	return p.configuration
}

func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

func (p *Plugin) OnConfigurationChange() error {
	if p.client == nil {
		p.client = pluginapi.NewClient(p.API, p.Driver)
	}

	configuration := p.getConfiguration()

	// Load the public configuration fields from the Mattermost server configuration.
	if loadConfigErr := p.API.LoadPluginConfiguration(configuration); loadConfigErr != nil {
		return errors.Wrap(loadConfigErr, "failed to load plugin configuration")
	}

	botID, ensureBotError := p.client.Bot.EnsureBot(&model.Bot{
		Username:    "summary-ai",
		DisplayName: "Summary AI",
		Description: "Summon to summary AI when you want to tap into an unknown conversation.",
	}, pluginapi.ProfileImagePath("/assets/icon.png"))
	if ensureBotError != nil {
		return errors.Wrap(ensureBotError, "failed to ensure Summary AI bot")
	}

	p.botID = botID

	p.setConfiguration(configuration)

	return nil
}
