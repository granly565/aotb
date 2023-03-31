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
			log.Printf("Failed send request in !followage command: %s", err)
			return
		}

		var result UserFollows
		err = json.Unmarshal(respBytes, &result)
		if err != nil {
			log.Printf("Failed parsing data to UserFollows struct in !followage command: %s", err)
			return
		}

		if result.Total == 0 {
			b.bot.Say(message.Message, "Ты всё ещё не подписан :(")
			return
		}

		year, month, day, _, _, _ := timex.Diff(result.Data[0].FollowedAt, time.Now())
		answerMessage := fmt.Sprintf("%s подписан уже", message.User.DisplayName)
		if year != 0 {
			if year > 4 {
				answerMessage += fmt.Sprintf(" %d лет", year)
			} else {
				postfix := ""
				if year > 1 {
					postfix = "а"
				}
				answerMessage += fmt.Sprintf(" %d год%s", year, postfix)
			}
		}
		if month != 0 {
			if year != 0 {
				if day == 0 {
					answerMessage += " и"
				} else {
					answerMessage += ","
				}
			}
			answerMessage += fmt.Sprintf(" %d месяц", month)
			if month > 4 {
				answerMessage += "ев"
			} else if month > 1 {
				answerMessage += "а"
			}
		}
		if day != 0 {
			if year != 0 || month != 0 {
				answerMessage += " и"
			}
			answerMessage += fmt.Sprintf(" %d д", day)
			if day%10 == 1 && day != 11 {
				answerMessage += "ень"
			} else {
				answerMessage += "н"
				if day%10 >= 2 && day%10 <= 4 && (day < 10 || day > 20) {
					answerMessage += "я"
				} else {
					answerMessage += "ей"
				}
			}
		}
		answerMessage += "!"
		if year == 0 && month == 0 && day == 0 {
			answerMessage = fmt.Sprintf("%s подписан с сегодняшнего дня!", message.User.DisplayName)
		}

		b.bot.Say(message.Channel, answerMessage)

	case "uptime":
		url := fmt.Sprintf("https://api.twitch.tv/helix/streams?user_login=%s", message.Channel)
		respBytes, err := getRequest(url)
		if err != nil {
			log.Printf("Failed send request in !uptime command: %s", err)
			return
		}

		var result UserStream
		err = json.Unmarshal(respBytes, &result)
		if err != nil {
			log.Printf("Failed parsing data to UserStream struct in !uptime command: %s", err)
			return
		}

		if len(result.Data) == 0 {
			b.bot.Say(message.Channel, "Стрим сейчас оффлайн")
			return
		}

		_, _, day, hour, minute, second := timex.Diff(result.Data[0].StartedAt, time.Now())

		answerMessage := "Стрим длится уже"
		if day != 0 {
			answerMessage += fmt.Sprintf(" %d д", day)
			if day%10 == 1 && day != 11 {
				answerMessage += "ень"
			} else {
				answerMessage += "н"
				if day%10 >= 2 && day%10 <= 4 && (day < 10 || day > 20) {
					answerMessage += "я"
				} else {
					answerMessage += "ей"
				}
			}
		}
		if hour != 0 {
			if day != 0 {
				if minute == 0 && second == 0 {
					answerMessage += " и"
				} else {
					answerMessage += ","
				}
			}
			answerMessage += fmt.Sprintf(" %d час", hour)
			if hour >= 5 && hour <= 20 {
				answerMessage += "ов"
			} else if hour != 1 && hour != 21 {
				answerMessage += "а"
			}
		}
		if minute != 0 {
			if day != 0 || hour != 0 {
				if second == 0 {
					answerMessage += " и"
				} else {
					answerMessage += ","
				}
			}
			answerMessage += fmt.Sprintf(" %d минут", minute)
			if minute%10 == 1 && minute != 11 {
				answerMessage += "у"
			} else if minute%10 >= 2 && minute%10 <= 4 && (minute < 10 || minute > 20) {
				answerMessage += "ы"
			}
		}
		if second != 0 {
			if day != 0 || hour != 0 || minute != 0 {
				answerMessage += " и"
			}
			answerMessage += fmt.Sprintf(" %d секунд", second)
			if second%10 == 1 && second != 11 {
				answerMessage += "у"
			} else if second%10 >= 2 && second%10 <= 4 && (second < 10 || second > 20) {
				answerMessage += "ы"
			}
		}
		answerMessage += "!"
		if day == 0 && hour == 0 && minute == 0 && second == 0 {
			answerMessage = "Стрим только начался!"
		}
		b.bot.Say(message.Channel, answerMessage)
	case "info":
		b.bot.Say(message.Channel, "Steam: https://steamcommunity.com/id/granly565; Канал в Discord: https://discord.com/invite/NjDzNM3;")
	case "archive":
		b.bot.Say(message.Channel, "Архивы со стримами: YouTube https://www.youtube.com/channel/UCiHb-yU7r4u59YNvjANcasQ; Телеграм https://t.me/archive_granly565;")
	case "меч":
		username := message.User.DisplayName
		if len(args) != 0 {
			username = args[0]
		}
		if username == "fiberka" {
			b.bot.Say(message.Channel, fmt.Sprintf("Меч %s длиной аж в %d см!", username, RandFromRange(-10, 0)))
		} else {
			b.bot.Say(message.Channel, fmt.Sprintf("Меч %s длиной аж в %d см!", username, RandFromRange(0, 30)))
		}
	case "обнимашки":
		chatters := GetFilteredChatters(message)

		if len(chatters) != 0 {
			b.bot.Say(message.Channel, fmt.Sprintf("%s крепко обнимает %s SirUwU", message.User.DisplayName, chatters[RandFromRange(0, len(chatters)-1)]))
		} else {
			b.bot.Say(message.Channel, "В чате некого обнимать :(")
		}
	case "снежок":
		chatters := GetFilteredChatters(message)

		if len(chatters) != 0 {
			b.bot.Say(message.Channel, fmt.Sprintf("%s попадает снежком в %s", message.User.DisplayName, chatters[RandFromRange(0, len(chatters)-1)]))
		} else {
			b.bot.Say(message.Channel, "Нет никого, в кого бросить снежком, поэтому вы бросаете в себя :(")
		}
	case "games", "игры", "геймс", "игоры":
		b.bot.Say(message.Channel, "Список запланированных игр: 1. В тылу врага 2 (Men of War); 2. Planescape: Torment")
	case "music", "музыка", "vepsrf", "ьгышс":
		b.bot.Say(message.Channel, "Заказ музыки: https://streamdj.ru/c/Granly")
		// case "auction", "аукцион":
		// 	b.bot.Say(message.Channel, "Сейчас проводится аукцион+розыгрыш. В чём суть: вначале протекает стандартный аукцион, в котором побеждает лот с наибольшим кол-вом баллов; затем все остальные лоты участвуют в розыгрыше, в котором кол-во баллов влияет на шанс выигрыша, и побеждает рандомная игра. Так мы выбираем две игры для последующих прохождений. Уточняйте, если что-то не понятно.")
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
