package main

import (
	"log"
	"os"
	"path"
	"sync"

	"github.com/gempir/go-twitch-irc/v3"
	"github.com/granly565/aotb/internal/twitchbot"
	tauth "github.com/granly565/aotb/internal/twitchbot/auth"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(path.Join(os.Getenv("HOME"), "/aotb/.env")); err != nil {
		log.Fatal("Error loading .env file")
	}

	if os.Getenv("BOT_REFRESH_TOKEN") == "" {
		log.Println("Refresh token is missing. Starting authentication...")
		var wg sync.WaitGroup
		tauth.TwitchAuthentication(&wg)
		wg.Wait()
	} else {
		tauth.RefreshTokens(os.Getenv("BOT_REFRESH_TOKEN"))
		go tauth.StartTaskRefreshTokens(os.Getenv("BOT_REFRESH_TOKEN"))
	}

	client := twitch.NewClient(os.Getenv("BOT_CHANNEL_NAME"), "oauth:"+os.Getenv("BOT_ACCESS_TOKEN"))

	bot := twitchbot.NewBot(client)
	if err := bot.Start(); err != nil {
		log.Fatalf("Failed starting a bot: %s", err)
	}
}
