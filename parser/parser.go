package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Contestant struct {
	Rank     int
	Name     string
	Country  string
	Solved   int
	Penalty  int
	Problems []ProblemResult
}

type ProblemResult struct {
	Status     string
	Attempts   int
	Time       string
	IsAccepted bool
}

type Service struct {
	doc         *goquery.Document
	contestants []Contestant
	stats       map[string]string
}

func NewSerive(doc *goquery.Document) *Service {
	return &Service{
		doc: doc,
	}
}

func (s *Service) ParseContestData() {
	var contestants []Contestant
	s.doc.Find(".standings tr[participantId]").Each(func(i int, s *goquery.Selection) {
		contestant := Contestant{}

		rankText := strings.TrimSpace(s.Find("td:first-child").Text())
		if rankText != "" && rankText != "&nbsp;" {
			contestant.Rank, _ = strconv.Atoi(rankText)
		}

		contestantCell := s.Find(".contestant-cell")

		flagImg := contestantCell.Find(".standings-flag")
		if flagImg.Length() > 0 {
			if alt, exists := flagImg.Attr("alt"); exists {
				contestant.Country = alt
			}
		}

		nameLink := contestantCell.Find("a.rated-user")
		if nameLink.Length() > 0 {
			contestant.Name = strings.TrimSpace(nameLink.Text())
		}

		solvedText := strings.TrimSpace(s.Find("td:nth-child(3)").Text())
		contestant.Solved, _ = strconv.Atoi(solvedText)

		penaltyText := strings.TrimSpace(s.Find("td:nth-child(4)").Text())
		if penaltyText != "" && penaltyText != "&nbsp;" {
			contestant.Penalty, _ = strconv.Atoi(penaltyText)
		}

		contestant.Problems = make([]ProblemResult, 5) // assuming 5 problems

		s.Find("td[problemId]").Each(func(j int, problemCell *goquery.Selection) {
			result := ProblemResult{}

			// Check if accepted
			acceptedSpan := problemCell.Find(".cell-accepted")
			if acceptedSpan.Length() > 0 {
				result.IsAccepted = true
				result.Status = strings.TrimSpace(acceptedSpan.Text())

				// Extract attempts from +1, +2, etc.
				if strings.Contains(result.Status, "+") {
					attemptsStr := strings.Trim(result.Status, "+")
					result.Attempts, _ = strconv.Atoi(attemptsStr)
					if result.Attempts == 0 {
						result.Attempts = 1 // + means 1 attempt
					}
				} else {
					result.Attempts = 1
				}

				// Extract time
				timeSpan := problemCell.Find(".cell-time")
				if timeSpan.Length() > 0 {
					result.Time = strings.TrimSpace(timeSpan.Text())
				}
			} else {
				// Check if rejected
				rejectedSpan := problemCell.Find(".cell-rejected")
				if rejectedSpan.Length() > 0 {
					result.Status = "-"
					result.IsAccepted = false
				}
			}

			contestant.Problems[j] = result
		})

		contestants = append(contestants, contestant)
	})

	s.contestants = contestants

	stats := make(map[string]string)
	s.doc.Find(".standingsStatisticsRow td.smaller").Each(func(i int, s *goquery.Selection) {
		if i > 0 {
			accepted := s.Find(".cell-passed-system-test").Text()
			tried := s.Find(".notice").Text()
			stats[fmt.Sprintf("Problem %c", 'A'+i-1)] = fmt.Sprintf("%s accepted / %s tried", accepted, tried)
		}

	})

	s.stats = stats
}

func (s *Service) GetContestantsData() []Contestant {
	return s.contestants
}

func (s *Service) GetContestStats() map[string]string {
	return s.stats
}
