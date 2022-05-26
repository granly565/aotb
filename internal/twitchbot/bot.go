package twitchbot

import (
	"os"

	"github.com/gempir/go-twitch-irc/v3"
)

type Bot struct {
	bot *twitch.Client
}

func NewBot(bot *twitch.Client) *Bot {
	return &Bot{bot: bot}
}

func (b *Bot) Start() error {

	b.AddHandlersToBot()

	b.bot.Join(os.Getenv("BOT_CHANNEL_NAME"))
	return b.bot.Connect()
}
