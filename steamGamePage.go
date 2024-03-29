package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type SteamGamePage struct {
	AppID                   int                          `json:"app_ID"`
	Available               bool                         `json:"available"`
	Categories              []SteamPageGameCategory      `json:"categories"`
	ComingSoon              bool                         `json:"coming_soon"`
	Description             string                       `json:"description"`
	Developers              []SteamPageGameDeveloper     `json:"developers"`
	EarlyAccess             bool                         `json:"early_access"`
	Genres                  []SteamPageGameGenre         `json:"genres"`
	Languages               []SteamPageGameLanguage      `json:"languages"`
	Name                    string                       `json:"name"`
	Publishers              []SteamPageGamePublisher     `json:"publishers"`
	ReleaseDate             time.Time                    `json:"release_date"`
	RequirementsMinimum     []SteamPageGameRequirement   `json:"requirements_minimum"`
	RequirementsRecommended []SteamPageGameRequirement   `json:"requirements_recommended"`
	ReviewsAll              SteamPageGameAggregateReview `json:"reviews_all"`
	ReviewsRecent           SteamPageGameAggregateReview `json:"reviews_recent"`
	SocialMedia             []SteamGameSocialMedia       `json:"social_media"`
	Tags                    []SteamPageGameTag           `json:"tags"`
	Title                   string                       `json:"title"`
	Timestamp               time.Time                    `json:"timestamp"`
	URL                     string                       `json:"URL"`
	Verbose                 string                       `json:"verbose"`
	Website                 string                       `json:"website"`
}

func NewSteamGamePage(s *goquery.Selection) *SteamGamePage {
	return &SteamGamePage{
		AppID:                   scrapeSteamGameAppID(s),
		Available:               scrapeSteamGameAvailable(s),
		Categories:              scrapeSteamGameCategories(s),
		ComingSoon:              scrapeSteamGameComingSoon(s),
		Description:             scrapeSteamGameDescription(s),
		Developers:              scrapeSteamGameDevelopers(s),
		EarlyAccess:             scrapeSteamGameEarlyAccess(s),
		Genres:                  scrapeSteamGameGenres(s),
		Languages:               scrapeSteamGameLanguages(s),
		Name:                    scrapeSteamGameName(s),
		Publishers:              scrapeSteamGamePublishers(s),
		ReleaseDate:             scrapeSteamGameReleaseDate(s),
		RequirementsMinimum:     scrapeSteamGameRequirementsMinimum(s),
		RequirementsRecommended: scrapeSteamGameRequirementsRecommended(s),
		ReviewsAll:              scrapeSteamGameReviewsAll(s),
		ReviewsRecent:           scrapeSteamGameReviewsRecent(s),
		SocialMedia:             scrapeSteamGameSocialMedia(s),
		Tags:                    scrapeSteamGameGameTags(s),
		Title:                   scrapeSteamGameTitle(s),
		Timestamp:               time.Now(),
		URL:                     scrapeSteamGameURL(s),
		Verbose:                 scrapeSteamGameVerbose(s),
		Website:                 scrapeSteamGameWebsite(s)}
}

func onGetSteamGamePage(c *http.Client, URL string, revisit bool, snap func(s *Snapshot), success func(s *SteamGamePage), err func(e error)) {
	if revisit == false {
		if u, err := url.Parse(URL); err == nil {
			if ok, _ := hasVisitedURLDefault(u); ok {
				return
			}
		}
	}
	lastAgeCheckCookie := &http.Cookie{
		Domain:   "store.steampowered.com",
		HttpOnly: false,
		Name:     "lastagecheckage",
		Path:     "/",
		Value:    "1-0-1940"}
	birthtimeCookie := &http.Cookie{
		Domain:   "store.steampowered.com",
		HttpOnly: false,
		Name:     "birthtime",
		Path:     "/",
		Value:    "-949485599"}
	snapshot := NewSnapshot(c, http.MethodGet, URL, &[]*http.Cookie{birthtimeCookie, lastAgeCheckCookie})
	snap(snapshot)
	if ok := (snapshot.StatusCode == http.StatusOK); ok != true {
		err(errors.New(snapshot.Status))
		return
	}
	if ok := (snapshot.document != nil); ok != true {
		err(snapshot.ErrDoc)
		return
	}
	CSSSelector := "html"
	goQuerySelection := snapshot.document.Find(CSSSelector)
	goQuerySelectionLength := goQuerySelection.Length()
	if ok := (goQuerySelectionLength > 0); ok != true {
		err(errors.New("goquery.Selection empty"))
		return
	}
	steamGamePage := NewSteamGamePage(goQuerySelection)
	if ok := steamGamePage.AppID > -1; ok != true {
		err(errors.New("SteamGamePage.AppID negative"))
		return
	}
	success(steamGamePage)
}

func scrapeSteamGameAppID(s *goquery.Selection) int {
	ID, _ := strconv.Atoi(s.Find("div[data-appid]").AttrOr("data-appid", "-1"))
	return ID
}

func scrapeSteamGameAvailable(s *goquery.Selection) bool {
	return (s.Find("div.game_area_comingsoon span.not_yet").First().Length() == 0)
}

func scrapeSteamGameCategories(s *goquery.Selection) []SteamPageGameCategory {
	var steamPageGameCategories []SteamPageGameCategory
	s.Find("div.game_area_details_specs a.name").Each(func(i int, s *goquery.Selection) {
		steamPageGameCategories = append(steamPageGameCategories, NewSteamPageGameCategory(s))
	})
	return steamPageGameCategories
}

func scrapeSteamGameComingSoon(s *goquery.Selection) bool {
	substring := strings.ReplaceAll(strings.TrimSpace(s.Find("div.coming_soon_date h1").First().Text()), " ", "-")
	ok := strings.ToUpper(substring) == "COMING-SOON"
	return ok
}

func scrapeSteamGameDescription(s *goquery.Selection) string {
	return strings.TrimSpace(s.Find("div.game_description_snippet").Text())
}

func scrapeSteamGameDevelopers(s *goquery.Selection) []SteamPageGameDeveloper {
	var steamPageGameDevelopers []SteamPageGameDeveloper
	s.Find("#developers_list a").Each(func(i int, s *goquery.Selection) {
		steamPageGameDeveloper := NewSteamPageGameDeveloper(s)
		steamPageGameDevelopers = append(steamPageGameDevelopers, steamPageGameDeveloper)
	})
	return steamPageGameDevelopers
}

func scrapeSteamGameEarlyAccess(s *goquery.Selection) bool {
	substring := strings.ReplaceAll(strings.TrimSpace(s.Find("div.early_access_header h1").First().Text()), " ", "-")
	ok := strings.ToUpper(substring) == "EARLY-ACCESS-GAME"
	return ok
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

func scrapeSteamGamePublishers(s *goquery.Selection) []SteamPageGamePublisher {
	var steamPageGamePublishers []SteamPageGamePublisher
	s.Find("div.user_reviews div.dev_row .summary:not([id])").Each(func(i int, s *goquery.Selection) {
		s.Find("a").Each(func(i int, s *goquery.Selection) {
			steamPageGamePublishers = append(steamPageGamePublishers, NewSteamPageGamePublisher(s))
		})
	})
	return steamPageGamePublishers
}

func scrapeSteamGameReleaseDate(s *goquery.Selection) time.Time {
	substring := strings.TrimSpace(s.Find("div.release_date div.date").First().Text())
	substring = strings.ReplaceAll(substring, ",", "")
	p := strings.Split(substring, " ")
	if ok := len(p) == 3; ok != true {
		return time.Time{}
	}
	day := p[0]
	if len(day) == 1 {
		day = fmt.Sprintf("0%s", day)
	}
	month := p[1]
	year := string(p[2][2:])
	timestamp := fmt.Sprintf("%s %s %s 00:00 UTC", day, month, year)
	t, err := time.Parse(time.RFC822, timestamp)
	if err != nil {
		return time.Time{}
	}
	return t
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

func scrapeSteamGameSocialMedia(s *goquery.Selection) []SteamGameSocialMedia {
	var steamGameSocialMedia []SteamGameSocialMedia
	s.Find("a[href][rel='noopener']").Each(func(i int, s *goquery.Selection) {
		if s.Children().Length() == 0 {
			return
		}
		href, exists := s.Attr("href")
		if exists != true {
			return
		}
		ok := strings.HasPrefix(href, "https://steamcommunity.com/linkfilter/?url")
		if ok != true {
			return
		}
		ok = strings.Contains(href, "facebook") || strings.Contains(href, "twitch") || strings.Contains(href, "twitter")
		if ok != true {
			return
		}
		/*for _, x := range s.Children().First().Nodes {
			fmt.Println(x)
		}*/
		steamGameSocialMedia = append(steamGameSocialMedia, NewSteamGameSocialMedia(s))
	})
	return steamGameSocialMedia
}

func scrapeSteamGameGameTags(s *goquery.Selection) []SteamPageGameTag {
	var steamPageGameTags []SteamPageGameTag
	s.Find("a.app_tag").Each(func(i int, s *goquery.Selection) {
		steamPageGameTags = append(steamPageGameTags, NewSteamPageGameTag(s))
	})
	return steamPageGameTags
}

func scrapeSteamGameTitle(s *goquery.Selection) string {
	return strings.TrimSpace(s.Find("div.apphub_AppName").First().Text())
}

func scrapeSteamGameURL(s *goquery.Selection) string {
	return strings.TrimSpace(s.Find("link[rel='canonical'][href]").First().AttrOr("href", ""))
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

func scrapeSteamGameWebsite(s *goquery.Selection) string {
	x := s.Find("a[href][rel].linkbar").FilterFunction(func(_ int, s *goquery.Selection) bool {
		substring := strings.ReplaceAll(strings.TrimSpace(s.Text()), " ", "-")
		ok := strings.ToUpper(substring) == "VISIT-THE-WEBSITE"
		return ok
	})
	return strings.TrimPrefix(strings.TrimSpace(x.First().AttrOr("href", "NIL")), "https://steamcommunity.com/linkfilter/?url=")
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
	filename := fmt.Sprintf("page-result-%s.json", strings.ToLower(s.Name))
	fullname := filepath.Join(fullpath, filename)
	err = ioutil.WriteFile(fullname, b, os.ModePerm)
	return err
}

func writeSteamGamePageDefault(s *SteamGamePage) error {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fullpath := filepath.Join(user.HomeDir, "Desktop", "steambot", "games", s.Name)
	err = writeSteamGamePage(fullpath, s)
	return err
}
