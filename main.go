package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	port = ":8080"
)

var (
	bot *tgbotapi.BotAPI
	ctx context.Context
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	var err error
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	ctx = context.Background()

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	updates := bot.ListenForWebhook("/bot")
	go http.ListenAndServe(port, nil)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "**Hava durumunu öğrenmek istediğin ili gir.‼️**")
				msg.ParseMode = "Markdown"
				bot.Send(msg)
			}
		} else {
			handleMessage(update.Message)
		}
	}
}

func handleMessage(message *tgbotapi.Message) {
	text := strings.ToLower(message.Text)

	if isCity(text) {
		weather, err := getWeather(text)
		if err != nil {
			log.Printf("Hava durumu alınırken hata: %s", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "**Hava durumu alınamadı. Lütfen tekrar deneyin.**")
			msg.ParseMode = "Markdown"
			bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, weather)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return
	}

	closestCity := getClosestCity(text)
	if closestCity == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "**Gönderdiğin ili bulamadım.**")
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return
	}

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Evet ✅", "evet"),
			tgbotapi.NewInlineKeyboardButtonData("Hayır ❌", "hayir"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Sahip 👍", "sahip"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("**%s** mı demek istediniz❓", strings.Title(closestCity)))
	msg.ReplyMarkup = markup
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	switch callbackQuery.Data {
	case "evet":
		weather, err := getWeather(callbackQuery.Message.Text)
		if err != nil {
			log.Printf("Hava durumu alınırken hata: %s", err)
			callbackQuery.Answer(fmt.Sprintf("Hava durumu alınamadı. Lütfen tekrar deneyin."), true)
			return
		}
		callbackQuery.Message.Reply(weather, nil)
	case "hayir":
		callbackQuery.Message.Reply("Lütfen ilinizi tekrar girin.", nil)
	case "sahip":
		callbackQuery.Answer("Bu bot, hava durumunu doğrudan almanız için @erd3mbey tarafından yazılmıştır.", true)
	default:
		callbackQuery.Answer("Bir şeyler ters gitti. Tekrar dene.", true)
	}
}

func isCity(text string) bool {
	for _, city := range cities {
		if strings.EqualFold(text, city) {
			return true
		}
	}
	return false
}

func getClosestCity(text string) string {
	closestCity := ""
	highestSimilarity := 0.0
	for _, city := range cities {
		similarity := difflib.GetCloseMatches(text, []string{city}, 1, 0.5)
		if len(similarity) > 0 && similarity[0] == city {
			return city
		}
		if similarityRatio(text, city) > highestSimilarity {
			highestSimilarity = similarityRatio(text, city)
			closestCity = city
		}
	}
	return closestCity
}

func similarityRatio(s1, s2 string) float64 {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if len(s1) < 3 || len(s2) < 3 {
		return 0.0
	}

	var (
		commonChars  int
		totalChars   int
		normalizedS1 []rune
		normalizedS2 []rune
	)

	for _, r1 := range s1 {
		if unicode.IsLetter(r1) {
			normalizedS1 = append(normalizedS1, r1)
		}
	}
	for _, r2 := range s2 {
		if unicode.IsLetter(r2) {
			normalizedS2 = append(normalizedS2, r2)
		}
	}

	totalChars = len(normalizedS1) + len(normalizedS2)

	for _, r1 := range normalizedS1 {
		for _, r2 := range normalizedS2 {
			if r1 == r2 {
				commonChars++
			}
		}
	}

	return 2.0 * float64(commonChars) / float64(totalChars)
}

func getWeather(city string) (string, error) {
	cmd := exec.Command("curl", "https://wttr.in/"+city+"?qmT0", "-H", "'Accept-Language: tr'")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

var cities = []string{
	"adana", "adıyaman", "afyonkarahisar", "ağrı", "amasya", "ankara", "antalya", "artvin", "aydın", "balıkesir",
	"bilecik", "bingöl", "bitlis", "bolu", "burdur", "bursa", "çanakkale", "çankırı", "çorum", "denizli", "diyarbakır",
	"edirne", "elazığ", "erzincan", "erzurum", "eskişehir", "gaziantep", "giresun", "gümüşhane", "hakkari", "hatay",
	"ısparta", "mersin", "istanbul", "izmir", "kars", "kastamonu", "kayseri", "kırklareli", "kırşehir", "kocaeli",
	"konya", "kütahya", "malatya", "manisa", "kahramanmaraş", "mardin", "muğla", "muş", "nevşehir", "niğde", "ordu",
	"rize", "sakarya", "samsun", "siirt", "sinop", "sivas", "tekirdağ", "tokat", "trabzon", "tunceli", "şanlıurfa",
	"uşak", "van", "yozgat", "zonguldak", "aksaray", "bayburt", "karaman", "kırıkkale", "batman", "şırnak", "bartın",
	"ardahan", "ığdır", "yalova", "karabük", "kilis", "osmaniye", "düzce",
}
