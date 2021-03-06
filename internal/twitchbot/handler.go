package twitchbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v3"
	"github.com/icza/gox/timex"
	"golang.org/x/exp/slices"
)

type UserFollows struct {
	Total int `json:"total"`
	Data  []struct {
		FromID     string    `json:"from_id"`
		FromLogin  string    `json:"from_login"`
		FromName   string    `json:"from_name"`
		ToID       string    `json:"to_id"`
		ToLogin    string    `json:"to_login"`
		ToName     string    `json:"to_name"`
		FollowedAt time.Time `json:"followed_at"`
	} `json:"data"`
}

type UserStream struct {
	Data []struct {
		ID           string    `json:"id"`
		UserID       string    `json:"user_id"`
		UserLogin    string    `json:"user_login"`
		UserName     string    `json:"user_name"`
		GameID       string    `json:"game_id"`
		GameName     string    `json:"game_name"`
		Type         string    `json:"type"`
		Title        string    `json:"title"`
		ViewerCount  int       `json:"viewer_count"`
		StartedAt    time.Time `json:"started_at"`
		Language     string    `json:"language"`
		ThumbnailURL string    `json:"thumbnail_url"`
		TagIds       []string  `json:"tag_ids"`
		IsMature     bool      `json:"is_mature"`
	} `json:"data"`
}

type UserChatters struct {
	ChatterCount int `json:"chatter_count"`
	Chatters     struct {
		Broadcaster []string `json:"broadcaster"`
		Vips        []string `json:"vips"`
		Moderators  []string `json:"moderators"`
		Staff       []string `json:"staff"`
		Admins      []string `json:"admins"`
		GlobalMods  []string `json:"global_mods"`
		Viewers     []string `json:"viewers"`
	} `json:"chatters"`
}

func (b *Bot) AddHandlersToBot() {
	b.bot.OnConnect(func() {
		log.Printf("Bot has been started.")
	})

	b.bot.OnPrivateMessage(func(message twitch.PrivateMessage) {
		if strings.HasPrefix(message.Message, "!") {
			b.HandleCommand(message)
		}
	})
}

func (b *Bot) HandleCommand(message twitch.PrivateMessage) {
	command := strings.Fields(message.Message)
	commandname, args := command[0][1:], command[1:]

	switch commandname {
	case "followage", "followtime":
		if message.User.Name == message.Channel {
			b.bot.Say(message.Channel, "Always has been B)")
			break
		}
		url := fmt.Sprintf("https://api.twitch.tv/helix/users/follows?from_id=%s&to_id=%s", message.User.ID, message.RoomID)
		respBytes, err := getRequest(url)
		if err != nil {
			log.Fatal(err)
		}

		var result UserFollows
		err = json.Unmarshal(respBytes, &result)
		if err != nil {
			log.Fatal(err)
		}

		if result.Total == 0 {
			b.bot.Say(message.Message, "???? ?????? ?????? ???? ???????????????? :(")
			return
		}

		year, month, day, _, _, _ := timex.Diff(result.Data[0].FollowedAt, time.Now())
		answerMessage := fmt.Sprintf("%s ???????????????? ??????", message.User.DisplayName)
		if year != 0 {
			if year > 4 {
				answerMessage += fmt.Sprintf(" %d ??????", year)
			} else {
				postfix := ""
				if year > 1 {
					postfix = "??"
				}
				answerMessage += fmt.Sprintf(" %d ??????%s", year, postfix)
			}
		}
		if month != 0 {
			if year != 0 {
				if day == 0 {
					answerMessage += " ??"
				} else {
					answerMessage += ","
				}
			}
			answerMessage += fmt.Sprintf(" %d ??????????", month)
			if month > 4 {
				answerMessage += "????"
			} else if month > 1 {
				answerMessage += "??"
			}
		}
		if day != 0 {
			if year != 0 || month != 0 {
				answerMessage += " ??"
			}
			answerMessage += fmt.Sprintf(" %d ??", day)
			if day%10 == 1 && day != 11 {
				answerMessage += "??????"
			} else {
				answerMessage += "??"
				if day%10 >= 2 && day%10 <= 4 && (day < 10 || day > 20) {
					answerMessage += "??"
				} else {
					answerMessage += "????"
				}
			}
		}
		answerMessage += "!"
		if year == 0 && month == 0 && day == 0 {
			answerMessage = fmt.Sprintf("%s ???????????????? ?? ???????????????????????? ??????!", message.User.DisplayName)
		}

		b.bot.Say(message.Channel, answerMessage)

	case "uptime":
		url := fmt.Sprintf("https://api.twitch.tv/helix/streams?user_login=%s", message.Channel)
		respBytes, err := getRequest(url)
		if err != nil {
			log.Fatal(err)
		}

		var result UserStream
		err = json.Unmarshal(respBytes, &result)
		if err != nil {
			log.Fatal(err)
		}

		_, _, day, hour, minute, second := timex.Diff(result.Data[0].StartedAt, time.Now())

		answerMessage := "?????????? ???????????? ??????"
		if day != 0 {
			answerMessage += fmt.Sprintf(" %d ??", day)
			if day%10 == 1 && day != 11 {
				answerMessage += "??????"
			} else {
				answerMessage += "??"
				if day%10 >= 2 && day%10 <= 4 && (day < 10 || day > 20) {
					answerMessage += "??"
				} else {
					answerMessage += "????"
				}
			}
		}
		if hour != 0 {
			if day != 0 {
				if minute == 0 && second == 0 {
					answerMessage += " ??"
				} else {
					answerMessage += ","
				}
			}
			answerMessage += fmt.Sprintf(" %d ??????", hour)
			if hour >= 5 && hour <= 20 {
				answerMessage += "????"
			} else if hour != 1 && hour != 21 {
				answerMessage += "??"
			}
		}
		if minute != 0 {
			if day != 0 || hour != 0 {
				if second == 0 {
					answerMessage += " ??"
				} else {
					answerMessage += ","
				}
			}
			answerMessage += fmt.Sprintf(" %d ??????????", minute)
			if minute%10 == 1 && minute != 11 {
				answerMessage += "??"
			} else if minute%10 >= 2 && minute%10 <= 4 && (minute < 10 || minute > 20) {
				answerMessage += "??"
			}
		}
		if second != 0 {
			if day != 0 || hour != 0 || minute != 0 {
				answerMessage += " ??"
			}
			answerMessage += fmt.Sprintf(" %d ????????????", second)
			if second%10 == 1 && second != 11 {
				answerMessage += "??"
			} else if second%10 >= 2 && second%10 <= 4 && (second < 10 || second > 20) {
				answerMessage += "??"
			}
		}
		answerMessage += "!"
		if day == 0 && hour == 0 && minute == 0 && second == 0 {
			answerMessage = "?????????? ???????????? ??????????????!"
		}
		b.bot.Say(message.Channel, answerMessage)
	case "info":
		b.bot.Say(message.Channel, "Steam: ... ?????????? ?? Discord: ...")
	case "??????":
		username := message.User.DisplayName
		if len(args) != 0 {
			username = args[0]
		}
		b.bot.Say(message.Channel, fmt.Sprintf("?????? %s ???????????? ???? ?? %d ????!", username, RandFromRange(0, 30)))
	case "??????????????????":
		chatters := GetFilteredChatters(message)

		if len(chatters) != 0 {
			b.bot.Say(message.Channel, fmt.Sprintf("%s ???????????? ???????????????? %s SirUwU", message.User.DisplayName, chatters[RandFromRange(0, len(chatters)-1)]))
		} else {
			b.bot.Say(message.Channel, "?? ???????? ???????????? ???????????????? :(")
		}
	case "????????????":
		chatters := GetFilteredChatters(message)

		if len(chatters) != 0 {
			b.bot.Say(message.Channel, fmt.Sprintf("%s ???????????????? ?????????????? ?? %s", message.User.DisplayName, chatters[RandFromRange(0, len(chatters)-1)]))
		} else {
			b.bot.Say(message.Channel, "?? ???????????? ?????????????? ??????????????, ?????????????? ???? ???????????????? ?? ???????? :(")
		}
	case "games", "????????", "??????????", "??????????":
		b.bot.Say(message.Channel, "???????????? ?????????????????????????????? ??????: ...")
	case "music", "????????????", "vepsrf", "??????????":
		b.bot.Say(message.Channel, "?????????? ????????????: ...")
	}

}

func RandFromRange(min, max int) int {
	if min > max {
		min, max = max, min
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func getRequest(url string) ([]byte, error) {
	client := http.Client{Timeout: time.Second * 10}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte(""), err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("BOT_ACCESS_TOKEN")))
	req.Header.Add("Client-Id", os.Getenv("BOT_CLIENT_ID"))

	resp, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}

	return data, err
}

func AppendAllSlices(slices ...[]string) []string {
	lenslices := 0
	for _, slice := range slices {
		lenslices += len(slice)
	}
	tmp := make([]string, lenslices)
	i := 0
	for _, slice := range slices {
		i += copy(tmp[i:], slice)
	}
	return tmp
}

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func GetFilteredChatters(message twitch.PrivateMessage) []string {
	url := fmt.Sprintf("https://tmi.twitch.tv/group/user/%s/chatters", message.Channel)
	respBytes, err := getRequest(url)
	if err != nil {
		log.Fatal(err)
	}

	var result UserChatters
	err = json.Unmarshal(respBytes, &result)
	if err != nil {
		log.Fatal(err)
	}
	chatters := AppendAllSlices(
		result.Chatters.Broadcaster,
		result.Chatters.Vips,
		result.Chatters.Moderators,
		result.Chatters.Viewers,
	)
	filterChatters := func(s string) bool {
		ignoreList := []string{message.User.Name, "soundalerts", "commanderroot", "nightbot", "streamelements", "anotherttvviewer", "moobot", "gametrendanalytics", "jointeffortt"}
		return !slices.Contains(ignoreList, s)
	}
	return Filter(chatters, filterChatters)
}
