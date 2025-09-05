package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/devasherr/contest-bot/parser"
	"github.com/devasherr/contest-bot/sheet"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

func handleFileUpload(bot *tgbot.BotAPI, update tgbot.Update) (io.ReadCloser, error) {
	fileID := update.Message.Document.FileID
	fileConfig := tgbot.FileConfig{FileID: fileID}
	file, err := bot.GetFile(fileConfig)
	if err != nil {
		return nil, err
	}

	fileUrl := file.Link(bot.Token)
	resp, err := bot.Client.Get(fileUrl)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env")
	}

	fmt.Println("server started!!")

	spreadSheetID := os.Getenv("SPREADSHEET_ID")
	if spreadSheetID == "" {
		log.Fatal("failed to load spreadsheet id")
	}

	telegramApiKey := os.Getenv("TELEGRAM_API_KEY")
	if telegramApiKey == "" {
		log.Fatal("failed to load telgram api key")
	}

	sheetSvc, err := sheet.NewService()
	if err != nil {
		log.Fatal("failed to start sheet serivice: ", err)
	}

	bot, err := tgbot.NewBotAPI(telegramApiKey)
	if err != nil {
		log.Fatal(err)
	}

	u := tgbot.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("failed to get updates channel")
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Text == "addContest" {
			msg := tgbot.NewMessage(update.Message.Chat.ID, "upload the contest page source")
			bot.Send(msg)
			continue
		}

		if update.Message.Document != nil {
			if filepath.Ext(update.Message.Document.FileName) != ".txt" {
				msg := tgbot.NewMessage(update.Message.Chat.ID, "please upload a .txt file")
				bot.Send(msg)
				continue
			}

			file, err := handleFileUpload(bot, update)
			doc, err := goquery.NewDocumentFromReader(file)
			if err != nil {
				log.Fatal(err)
			}

			contestName := strings.TrimSpace(doc.Find(".contest-name a").Text())
			parseSvc := parser.NewSerive(doc)
			parseSvc.ParseContestData()

			if err := sheetSvc.CreateSheetIfNotExists(contestName, spreadSheetID); err != nil {
				log.Fatal(err)
			}
			err = sheetSvc.WriteContestData(contestName, spreadSheetID, parseSvc.GetContestantsData(), parseSvc.GetContestStats())
			if err != nil {
				log.Fatal("failed to save contest to spread sheet")
			}
		}
	}
}
