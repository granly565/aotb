package tokens

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Tokens struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

func TwitchAuthentication(wg *sync.WaitGroup) {
	state := GenerateURLWithState()
	ServerForGenTokens(state, wg)
}

func GenerateURLWithState() string {
	scope := []string{
		"channel%3Amanage%3Abroadcast",
		"channel%3Amanage%3Apolls",
		"channel%3Aread%3Apolls",
		"channel%3Amanage%3Aredemptions",
		"channel%3Aread%3Aredemptions",
		"user%3Aread%3Afollows",
		"chat%3Aedit",
		"chat%3Aread",
		"whispers%3Aread",
		"whispers%3Aedit",
		"channel%3Amoderate",
	}
	state := randStringBytes(32)

	uri := "https://id.twitch.tv/oauth2/authorize" +
		"?response_type=code" +
		"&client_id=" + os.Getenv("BOT_CLIENT_ID") +
		"&scope=" + strings.Join(scope, "+") +
		"&state=" + state +
		"&redirect_uri=" + os.Getenv("BOT_REDIRECT_URI")
	fmt.Printf("Enter this URL in the browser address bar:\n%s", uri)
	return state
}

func ServerForGenTokens(state string, wg *sync.WaitGroup) {
	wg.Add(1)

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth_code", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		if params.Get("state") != state {
			log.Println("State is not equal in request. Request is ignored.")
			return
		}

		if params.Get("error") == "access_denied" {
			log.Printf("Authentication is failed, skip this process. Check Application in Twitch console or .env file on correct bot data. Description: %s", params.Get("error_description"))
			wg.Done()
			return
		}

		code := params.Get("code")
		if code != "" {
			uri := "https://id.twitch.tv/oauth2/token"
			uri_params := "client_id=" + os.Getenv("BOT_CLIENT_ID") +
				"&client_secret=" + os.Getenv("BOT_SECRET") +
				"&code=" + code +
				"&grant_type=authorization_code" +
				"&redirect_uri=" + os.Getenv("BOT_REDIRECT_URI")
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

			tokens := Tokens{}
			err = json.Unmarshal(bodyBytes, &tokens)
			if err != nil {
				log.Println(err)
				return
			}

			err = UpdateRefreshTokenInEnvFile(tokens)
			if err != nil {
				log.Println(err)
				return
			}

			os.Setenv("BOT_ACCESS_TOKEN", tokens.AccessToken)
			log.Println("Authentication is done.")
			wg.Done()
		}
	})

	srv := &http.Server{
		Addr:    ":443",
		Handler: mux,
	}

	go func() {
		err := srv.ListenAndServeTLS("keys/auth.crt", "keys/auth.key")
		if err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
		wg.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	wg.Wait()
}

func randStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := make([]byte, n)

	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func UpdateRefreshTokenInEnvFile(tokens Tokens) error {
	envMap, err := godotenv.Read()
	if err != nil {
		return errors.New("error reading .env file")
	}

	envMap["BOT_REFRESH_TOKEN"] = tokens.RefreshToken

	if err = godotenv.Write(envMap, ".env"); err != nil {
		return errors.New("error writing to .env file")
	}
	err = os.Setenv("BOT_REFRESH_TOKEN", tokens.RefreshToken)
	if err != nil {
		return err
	}
	return nil
}
