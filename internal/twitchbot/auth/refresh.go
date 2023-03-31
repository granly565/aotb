package tokens

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

func StartTaskRefreshTokens(refresh string) {
	for {
		<-time.After(2 * time.Hour)
		go RefreshTokens(refresh)
	}
}

func RefreshTokens(refresh string) {
	uri := "https://id.twitch.tv/oauth2/token"
	uri_params := "client_id=" + os.Getenv("BOT_CLIENT_ID") +
		"&client_secret=" + os.Getenv("BOT_SECRET") +
		"&grant_type=refresh_token" +
		"&refresh_token=" + url.QueryEscape(os.Getenv("BOT_REFRESH_TOKEN"))
	resp, err := http.Post(uri, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(uri_params)))
	if err != nil {
		log.Println(err)
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var resp_data map[string]interface{}
	err = json.Unmarshal(bodyBytes, &resp_data)
	if err != nil {
		log.Println(err)
		return
	}

	if resp_data["status"] == float64(400) {
		log.Printf("Failed updating tokens: %s", resp_data["message"])
		return
	}

	tokens := Tokens{}
	err = json.Unmarshal(bodyBytes, &tokens)
	if err != nil {
		log.Println(err)
		return
	}

	UpdateRefreshTokenInEnvFile(tokens)
	os.Setenv("BOT_ACCESS_TOKEN", tokens.AccessToken)
	log.Println("Tokens has been updated.")
}
