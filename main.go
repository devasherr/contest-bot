package main

import (
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/devasherr/contest-bot/parser"
	"github.com/devasherr/contest-bot/sheet"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env")
	}

	spreadSheetID := os.Getenv("SPREADSHEET_ID")

	file, err := os.Open("page-content.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		log.Fatal(err)
	}

	contestName := strings.TrimSpace(doc.Find(".contest-name a").Text())
	parseSvc := parser.NewSerive(doc)
	parseSvc.ParseContestData()

	sheetSvc, err := sheet.NewService()
	if err != nil {
		log.Fatal("failed to start sheet serivice: ", err)
	}

	if err := sheetSvc.CreateSheetIfNotExists(contestName, spreadSheetID); err != nil {
		log.Fatal(err)
	}
	err = sheetSvc.WriteContestData(contestName, spreadSheetID, parseSvc.GetContestantsData(), parseSvc.GetContestStats())
	if err != nil {
		log.Fatal("failed to save contest to spread sheet")
	}
}
