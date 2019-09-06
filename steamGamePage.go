package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamGamePage struct {
	AppID                   int                          `json:"app_ID"`
	Categories              []SteamPageGameCategory      `json:"categories"`
	Description             string                       `json:"description"`
	Developers              []SteamPageGameDeveloper     `json:"developers"`
	Genres                  []SteamPageGameGenre         `json:"genres"`
	Languages               []SteamPageGameLanguage      `json:"languages"`
	Name                    string                       `json:"name"`
	RequirementsMinimum     []SteamPageGameRequirement   `json:"requirements_minimum"`
	RequirementsRecommended []SteamPageGameRequirement   `json:"requirements_recommended"`
	ReviewsAll              SteamPageGameAggregateReview `json:"reviews_all"`
	ReviewsRecent           SteamPageGameAggregateReview `json:"reviews_recent"`
	Tags                    []SteamPageGameTag           `json:"tags"`
	URL                     string                       `json:"URL"`
	Verbose                 string                       `json:"verbose"`
}

func NewSteamGamePage(s *goquery.Selection) *SteamGamePage {
	return &SteamGamePage{
		AppID:                   scrapeSteamGameAppID(s),
		Categories:              scrapeSteamGameCategories(s),
		Description:             scrapeSteamGameDescription(s),
		Developers:              scrapeSteamGameDevelopers(s),
		Genres:                  scrapeSteamGameGenres(s),
		Languages:               scrapeSteamGameLanguages(s),
		Name:                    scrapeSteamGameName(s),
		RequirementsMinimum:     scrapeSteamGameRequirementsMinimum(s),
		RequirementsRecommended: scrapeSteamGameRequirementsRecommended(s),
		ReviewsAll:              scrapeSteamGameReviewsAll(s),
		ReviewsRecent:           scrapeSteamGameReviewsRecent(s),
		Tags:                    scrapeSteamGameGameTags(s),
		URL:                     scrapeSteamGameURL(s),
		Verbose:                 scrapeSteamGameVerbose(s)}
}

func scrapeSteamGameAppID(s *goquery.Selection) int {
	ID, _ := strconv.Atoi(s.Find("div[data-appid]").AttrOr("data-appid", "-1"))
	return ID
}

func scrapeSteamGameDescription(s *goquery.Selection) string {
	return strings.TrimSpace(s.Find("div.game_description_snippet").Text())
}

func scrapeSteamGameCategories(s *goquery.Selection) []SteamPageGameCategory {
	var steamPageGameCategories []SteamPageGameCategory
	s.Find("div.game_area_details_specs a.name").Each(func(i int, s *goquery.Selection) {
		steamPageGameCategories = append(steamPageGameCategories, NewSteamPageGameCategory(s))
	})
	return steamPageGameCategories
}

func scrapeSteamGameDevelopers(s *goquery.Selection) []SteamPageGameDeveloper {
	var steamPageGameDevelopers []SteamPageGameDeveloper
	s.Find("#developers_list a").Each(func(i int, s *goquery.Selection) {
		steamPageGameDeveloper := NewSteamPageGameDeveloper(s)
		steamPageGameDevelopers = append(steamPageGameDevelopers, steamPageGameDeveloper)
	})
	return steamPageGameDevelopers
}

func scrapeSteamGameGenres(s *goquery.Selection) []SteamPageGameGenre {
	var steamPageGameGenres []SteamPageGameGenre
	s.Find("div.game_details div.details_block:first-child > a").Each(func(i int, s *goquery.Selection) {
		steamPageGameGenres = append(steamPageGameGenres, NewSteamPageGameGenre(s))
	})
	return steamPageGameGenres
}

func scrapeSteamGameLanguages(s *goquery.Selection) []SteamPageGameLanguage {
	var steamPageGameLanguages []SteamPageGameLanguage
	s.Find("table.game_language_options tr[class='']").Each(func(i int, s *goquery.Selection) {
		steamPageGameLanguages = append(steamPageGameLanguages, NewSteamPageGameLanguage(s))
	})
	return steamPageGameLanguages
}

func scrapeSteamGameName(s *goquery.Selection) string {
	return regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(strings.TrimSpace(s.Find("div.apphub_AppName").Text()), "")
}

func scrapeSteamGameRequirementsMinimum(s *goquery.Selection) []SteamPageGameRequirement {
	var steamPageGameRequirements []SteamPageGameRequirement
	s.Find("div.game_area_sys_req").Each(func(i int, s *goquery.Selection) {
		OS, ok := s.Attr("data-os")
		if ok != true {
			return
		}
		s.Find(".game_area_sys_req_full").Each(func(_ int, s *goquery.Selection) {
			s.RemoveClass(".game_area_sys_req_full").AddClass("game_area_sys_req_leftCol")
		})
		s.Find(".game_area_sys_req_leftCol").First().Each(func(i int, s *goquery.Selection) {
			steamPageGameRequirement := NewSteamPageGameRequirement(s)
			steamPageGameRequirement.OS = OS
			steamPageGameRequirements = append(steamPageGameRequirements, steamPageGameRequirement)
		})
	})
	return steamPageGameRequirements
}

func scrapeSteamGameRequirementsRecommended(s *goquery.Selection) []SteamPageGameRequirement {
	var steamPageGameRequirements []SteamPageGameRequirement
	s.Find("div.game_area_sys_req").Each(func(i int, s *goquery.Selection) {
		OS, ok := s.Attr("data-os")
		if ok != true {
			return
		}
		s.Find(".game_area_sys_req_full").Each(func(_ int, s *goquery.Selection) {
			s.RemoveClass(".game_area_sys_req_full").AddClass("game_area_sys_req_leftCol")
		})
		s.Find(".game_area_sys_req_rightCol").First().Each(func(i int, s *goquery.Selection) {
			steamPageGameRequirement := NewSteamPageGameRequirement(s)
			steamPageGameRequirement.OS = OS
			steamPageGameRequirements = append(steamPageGameRequirements, steamPageGameRequirement)
		})
	})
	return steamPageGameRequirements
}

func scrapeSteamGameReviewsAll(s *goquery.Selection) SteamPageGameAggregateReview {
	return NewSteamPageGameAggregateReview(s.Find(".user_reviews_summary_row[itemprop]"))
}

func scrapeSteamGameReviewsRecent(s *goquery.Selection) SteamPageGameAggregateReview {
	return NewSteamPageGameAggregateReview(s.Find(".user_reviews_summary_row:not([itemprop])"))
}

func scrapeSteamGameGameTags(s *goquery.Selection) []SteamPageGameTag {
	var steamPageGameTags []SteamPageGameTag
	s.Find("a.app_tag").Each(func(i int, s *goquery.Selection) {
		steamPageGameTags = append(steamPageGameTags, NewSteamPageGameTag(s))
	})
	return steamPageGameTags
}

func scrapeSteamGameURL(s *goquery.Selection) string {
	return strings.TrimSpace(s.Find("link[rel='canonical'][href]").First().Text())
}

func scrapeSteamGameVerbose(s *goquery.Selection) string {
	var verbose []string
	s.Find("#game_area_description").Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#text" {
			verbose = append(verbose, strings.TrimSpace(s.Text()))
		}
	})
	return strings.Join(verbose, "")
}

func writeSteamGamePage(fullpath string, s *SteamGamePage) error {
	err := os.MkdirAll(fullpath, os.ModePerm)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("page-result-%d.json", s.AppID)
	fullname := filepath.Join(fullpath, filename)
	err = ioutil.WriteFile(fullname, b, os.ModePerm)
	return err
}

func writeSteamGamePageDefault(s *SteamGamePage) error {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fullpath := filepath.Join(user.HomeDir, "Desktop", "steambot", s.Name)
	err = writeSteamGamePage(fullpath, s)
	return err
}