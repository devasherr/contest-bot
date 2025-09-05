package bot

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/devasherr/contest-bot/parser"
	"github.com/devasherr/contest-bot/sheet"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Service struct {
	parseSvc *parser.Service
	sheetSvc *sheet.Service
	bot      *tgbot.BotAPI
}

func NewService(parseSvc *parser.Service, sheetSvc *sheet.Service, telegramApiKey string) *Service {
	bot, err := tgbot.NewBotAPI(telegramApiKey)
	if err != nil {
		log.Fatal(err)
	}

	return &Service{
		parseSvc: parseSvc,
		sheetSvc: sheetSvc,
		bot:      bot,
	}
}

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

func (s *Service) Start() {
	u := tgbot.NewUpdate(0)
	u.Timeout = 60

	updates, err := s.bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("failed to get updates channel")
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Text == "addContest" {
			msg := tgbot.NewMessage(update.Message.Chat.ID, "upload the contest page source")
			s.bot.Send(msg)
			continue
		}

		if update.Message.Document != nil {
			if filepath.Ext(update.Message.Document.FileName) != ".txt" {
				msg := tgbot.NewMessage(update.Message.Chat.ID, "please upload a .txt file")
				s.bot.Send(msg)
				continue
			}

			file, err := handleFileUpload(s.bot, update)
			doc, err := goquery.NewDocumentFromReader(file)
			if err != nil {
				log.Fatal(err)
			}

			contestName := strings.TrimSpace(doc.Find(".contest-name a").Text())
			s.parseSvc.ParseContestData(doc)

			if err := s.sheetSvc.CreateSheetIfNotExists(contestName); err != nil {
				log.Fatal(err)
			}
			err = s.sheetSvc.WriteContestData(contestName, s.parseSvc.GetContestantsData(), s.parseSvc.GetContestStats())
			if err != nil {
				log.Fatal("failed to save contest to spread sheet")
			}
		}
	}
}
