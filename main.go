package main

import (
	"fmt"
	"github.com/devasherr/contest-bot/bot"
	"github.com/devasherr/contest-bot/parser"
	"github.com/devasherr/contest-bot/sheet"
	"github.com/joho/godotenv"
	"log"
	"os"
)

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

	parseSvc := parser.NewSerive()
	sheetSvc, err := sheet.NewService(spreadSheetID)
	if err != nil {
		log.Fatal("failed to start sheet serivice: ", err)
	}
	botSvc := bot.NewService(parseSvc, sheetSvc, telegramApiKey)
	botSvc.Start()
}
