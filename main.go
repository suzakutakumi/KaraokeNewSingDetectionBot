package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const (
	KEYWORD_SEARCH_URL = "https://www.clubdam.com/dkwebsys/search-api/SearchVariousByKeywordApi"
	MUSIC_SEARCH_URL   = "https://www.clubdam.com/dkwebsys/search-api/SearchMusicByKeywordApi"
)

type PostData struct {
	ModelTypeCode string `json:"modelTypeCode"`
	SerialNo      string `json:"serialNo"`
	Keyword       string `json:"keyword"`
	CompId        string `json:"compId"`
	AuthKey       string `json:"authKey"`
	Sort          string `json:"sort"`
	DispCount     string `json:"dispCount"`
	PageNo        string `json:"pageNo"`
}

func NewPostData(keyword string) PostData {
	return PostData{
		"1",
		"AT00001",
		keyword,
		"1",
		"2/Qb9R@8s*",
		"2",
		"100",
		"1",
	}
}

type Status struct {
	StatusCode string `json:"statusCode"`
	Message    string `json:"message"`
}

type MetaData struct {
	PageCount  int `json:"pageCount"`
	TotalCount int `json:"totalCount"`
}

type Music struct {
	RequestNumber string `json:"requestNo"`
	Title         string `json:"title"`
	Artist        string `json:"artist"`
}

type SearchResult struct {
	Result    Status    `json:"result"`
	Data      *MetaData `json:"data"`
	MusicList *[]Music  `json:"list"`
}

var s *discordgo.Session
var AppID string
var GuildID string

func init() {
	var err error

	if err = godotenv.Load(); err != nil {
		log.Fatalln(".envが読み込めませんでした")
	}

	if s, err = discordgo.New("Bot " + os.Getenv("TOKEN")); err != nil {
		log.Fatalln("discordBotの初期化に失敗", err.Error())
	}

	AppID = os.Getenv("APP_ID")
	GuildID = "795782540710510592"
}

func searchByKeyword(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) error {
	var value string = ""
	for _, opt := range options {
		if opt.Name == "keyword" {
			v, ok := opt.Value.(string)
			if ok || v != "" {
				value = v
			}
		}
	}
	if value == "" {
		return fmt.Errorf("値をうまく取得できませんでした")
	}

	// ここからHTTPリクエスト
	jsonString, err := json.Marshal(NewPostData(value))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", KEYWORD_SEARCH_URL, bytes.NewBuffer(jsonString))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		panic("Error")
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic("Error")
	}

	var data SearchResult
	if err = json.Unmarshal(byteArray, &data); err != nil {
		return err
	}

	// 返答処理
	if data.Result.StatusCode != "0000" {
		return fmt.Errorf("HTTPリクエストに失敗しました")
	}

	content := ""
	content += fmt.Sprintf("キーワード「%s」で検索結果: %d曲\n", value, data.Data.TotalCount)
	for _, m := range *data.MusicList {
		content += fmt.Sprintf("● %s / %s  %s\n", m.Title, m.Artist, m.RequestNumber)
	}

	if len(content) > 2000 {
		content = string([]rune(content)[:1996]) + "\n..."
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})

	return err
}

func searchByMusicName(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) error {
	var value string = ""
	for _, opt := range options {
		if opt.Name == "music_name" {
			v, ok := opt.Value.(string)
			if ok || v != "" {
				value = v
			}
		}
	}
	if value == "" {
		return fmt.Errorf("値をうまく取得できませんでした")
	}

	// ここからHTTPリクエスト
	jsonString, err := json.Marshal(NewPostData(value))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", MUSIC_SEARCH_URL, bytes.NewBuffer(jsonString))
	if err != nil {
		fmt.Println("fuck")
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		panic("Error")
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic("Error")
	}

	var data SearchResult
	if err = json.Unmarshal(byteArray, &data); err != nil {
		return err
	}

	// 返答処理
	if data.Result.StatusCode != "0000" {
		return fmt.Errorf("HTTPリクエストに失敗しました")
	}

	content := ""
	content += fmt.Sprintf("キーワード「%s」で検索結果: %d曲\n", value, data.Data.TotalCount)
	for _, m := range *data.MusicList {
		content += fmt.Sprintf("● %s / %s  %s\n", m.Title, m.Artist, m.RequestNumber)
	}

	if len(content) > 2000 {
		content = string([]rune(content)[:1996]) + "\n..."
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})

	return err
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	// Components are part of interactions, so we register InteractionCreate handler
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if i.ApplicationCommandData().Name == "search" {
				options := make(map[string][]*discordgo.ApplicationCommandInteractionDataOption)
				for _, opt := range i.ApplicationCommandData().Options {
					options[opt.Name] = opt.Options
				}
				v1, ok1 := options["keyword"]
				v2, ok2 := options["music"]
				if ok1 {
					err := searchByKeyword(s, i, v1)
					if err != nil {
						s.ChannelMessageSend(i.ChannelID, err.Error())
						fmt.Println("知らんが、落ちた", err.Error())
					}
				} else if ok2 {
					err := searchByMusicName(s, i, v2)
					if err != nil {
						s.ChannelMessageSend(i.ChannelID, err.Error())
						fmt.Println("知らんが、落ちた", err.Error())
					}
				} else {
					log.Fatalln("コマンドに対応する処理が見つかりませんでした")
				}
			}
		case discordgo.InteractionMessageComponent:

		}
	})

	_, err := s.ApplicationCommandCreate(AppID, "", &discordgo.ApplicationCommand{
		Name: "search",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "keyword",
				Description: "キーワード検索します",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "keyword",
						Required:    true,
						Description: "曲のキーワード",
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "music",
				Description: "曲名で検索します",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "music_name",
						Required:    true,
						Description: "曲名",
					},
				},
			},
		},
		Description: "Lo and behold: dropdowns are coming",
	})

	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}
