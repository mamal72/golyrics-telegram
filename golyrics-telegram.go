package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mamal72/golyrics"
	"github.com/tucnak/telebot"
)

func getTrackTitle(track *golyrics.Track) string {
	return fmt.Sprintf("%s:%s", track.Artist, track.Name)
}

func onStart(bot *telebot.Bot, message *telebot.Message) {
	bot.SendMessage(
		message.Chat,
		fmt.Sprintf(`Hello, %s!
Send me a track name to get lyrics of it! 🎵`, message.Sender.FirstName),
		nil,
	)
}

func onHelp(bot *telebot.Bot, message *telebot.Message) {
	bot.SendMessage(
		message.Chat,
		"Just send me a track name to get lyrics of it! 🎵",
		nil,
	)
}

func onError(bot *telebot.Bot, message *telebot.Message, err error) {
	fmt.Println(err)
	bot.SendMessage(
		message.Chat,
		"An error happened! ☹️",
		nil,
	)
}

func onNotFound(bot *telebot.Bot, message *telebot.Message, query string) {
	bot.SendMessage(
		message.Chat,
		fmt.Sprintf("No lyrics found for \"%s\"! ☹️", query),
		nil,
	)
}

func onNoTextMessage(bot *telebot.Bot, message *telebot.Message) {
	bot.SendMessage(
		message.Chat,
		"I only understand text messages. ☹️",
		nil,
	)
}

func sendLyrics(bot *telebot.Bot, message *telebot.Message, track *golyrics.Track) {
	if len(track.Lyrics) == 0 {
		track.Lyrics = fmt.Sprintf("No lyrics found for %s:%s! ☹️\nIt may be an album, not a track!", track.Artist, track.Name)
	}
	bot.SendMessage(
		message.Chat,
		fmt.Sprintf("🎵 *%s* by *%s*:\n\n%s", track.Name, track.Artist, track.Lyrics),
		&telebot.SendOptions{
			ParseMode: telebot.ModeMarkdown,
		},
	)
}

func sendSuggestions(bot *telebot.Bot, message *telebot.Message, query string, suggestions *[]golyrics.Track) {
	replyKeyboard := [][]string{}
	for _, track := range *suggestions {
		trackTitle := getTrackTitle(&track)
		replyKeyboard = append(replyKeyboard, []string{trackTitle})
	}
	bot.SendMessage(
		message.Chat,
		fmt.Sprintf("Found *%d* items for \"*%s*\". Which one is it?", len(*suggestions), query),
		&telebot.SendOptions{
			ParseMode: telebot.ModeMarkdown,
			ReplyMarkup: telebot.ReplyMarkup{
				ForceReply:      true,
				OneTimeKeyboard: true,
				CustomKeyboard:  replyKeyboard,
			},
		},
	)
}

func main() {
	// Load ENV vars from .env file
	godotenv.Load()

	botToken := os.Getenv("BOT_TOKEN")
	if len(botToken) == 0 {
		log.Fatalln("No Telegram bot token (BOT_TOKEN) provided in .env file or ENV variables")
	}

	bot, err := telebot.NewBot(botToken)
	if err != nil {
		log.Fatalln(err)
	}

	messages := make(chan telebot.Message, 100)
	bot.Listen(messages, 30*time.Second)

	for message := range messages {
		messageText := strings.TrimSpace(message.Text)

		// Only text messages
		if len(messageText) == 0 {
			onNoTextMessage(bot, &message)
			continue
		}

		// Bot started
		if message.Text == "/start" {
			onStart(bot, &message)
			continue
		}

		// Bot help
		if message.Text == "/help" {
			onHelp(bot, &message)
			continue
		}

		suggestions, err := golyrics.SearchTrack(message.Text)
		// On error
		if err != nil {
			onError(bot, &message, err)
			continue
		}
		// On not found
		if len(suggestions) == 0 {
			onNotFound(bot, &message, messageText)
			continue
		}

		trackTitle := getTrackTitle(&suggestions[0])

		if len(suggestions) > 1 && trackTitle != messageText {
			sendSuggestions(bot, &message, messageText, &suggestions)
			continue
		}

		track := suggestions[0]
		track.FetchLyrics()
		sendLyrics(bot, &message, &track)
		continue
	}
}
