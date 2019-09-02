package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type gameCatalogue map[string]game

func (gameCatalogue gameCatalogue) Add(game game) bool {
	gameCatalogue[game.URL] = game
	return gameCatalogue.Has(game.URL)
}

func (gameCatalogue gameCatalogue) Get(key string) (game, bool) {
	game, ok := gameCatalogue[key]
	return game, ok
}

func (gameCatalogue gameCatalogue) Has(key string) bool {
	_, ok := gameCatalogue.Get(key)
	return ok
}

type queryMap map[string]string

func (queryMap queryMap) Add(key, value string) bool {
	queryMap[key] = value
	return queryMap.Has(key)
}

func (queryMap queryMap) Get(key string) (string, bool) {
	value, ok := queryMap[key]
	return value, ok
}

func (queryMap queryMap) Has(key string) bool {
	_, ok := queryMap.Get(key)
	return ok
}

type queryCategories map[string]queryMap

func (queryCategories queryCategories) Add(tag, key, value string) bool {
	if ok := queryCategories.Has(tag); ok != true {
		queryCategories.Set(tag)
	}
	queryMap, _ := queryCategories.Get(tag)
	queryMap.Add(key, value)
	queryCategories[tag] = queryMap
	return queryCategories.Has(tag)
}

func (queryCategories queryCategories) Get(tag string) (queryMap, bool) {
	queryMap, ok := queryCategories[tag]
	return queryMap, ok
}

func (queryCategories queryCategories) Has(tag string) bool {
	_, ok := queryCategories[tag]
	return ok
}

func (queryCategories queryCategories) Set(tag string) bool {
	_, ok := queryCategories[tag]
	if ok != true {
		queryCategories[tag] = queryMap{}
	}
	return (ok == false)
}

type game struct {
	AppID              string            `json:"appid"`
	BundleID           string            `json:"bundleid"`
	Categories         []gameCategory    `json:"categories"`
	CrtrID             string            `json:"crtrid"`
	DescriptionID      string            `json:"descriptionid"`
	Description        string            `json:"description"`
	DescriptionVerbose string            `json:"descriptionverbose"`
	Developer          []gameDeveloper   `json:"developer"`
	Genre              []gameGenre       `json:"genre"`
	Languages          []gameLanguage    `json:"languages"`
	Meta               []gameMeta        `json:"meta"`
	Name               string            `json:"name"`
	PackageID          string            `json:"packageid"`
	Publisher          []gamePublisher   `json:"publisher"`
	ReleaseDate        string            `json:"releasedate"`
	Requirements       []gameRequirement `json:"requirements"`
	TagID              string            `json:"tagid"`
	Tags               []string          `json:"tags"`
	Title              string            `json:"title"`
	URL                string            `json:"url"`
}

type gameCategory struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (gameCategory gameCategory) String() string {
	return fmt.Sprintf("%s", gameCategory.Name)
}

type gameDeveloper struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (gameDeveloper gameDeveloper) String() string {
	return fmt.Sprintf("%s", gameDeveloper.Name)
}

type gameGenre struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (gameGenre gameGenre) String() string {
	return fmt.Sprintf("%s", gameGenre.Name)
}

type gameLanguage struct {
	Audio     bool   `json:"audio"`
	Interface bool   `json:"interface"`
	Name      string `json:"name"`
	Subtitles bool   `json:"subtitles"`
}

type gameMeta struct {
	Content  string `json:"content"`
	Name     string `json:"name"`
	Property string `json:"property"`
}

type gamePublisher struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (gamePublisher gamePublisher) String() string {
	return fmt.Sprintf("%s", gamePublisher.Name)
}

type gameRequirement struct {
	DirectX   string `json:"directx"`
	Graphics  string `json:"graphics"`
	Memory    string `json:"memory"`
	Name      string `json:"name"`
	Network   string `json:"network"`
	OS        string `json:"os"`
	Processor string `json:"processor"`
	SoundCard string `json:"soundcard"`
	Storage   string `json:"storage"`
}

const steamSearchURL string = "https://store.steampowered.com/search/"

var hrefGroup []string

var wg sync.WaitGroup

var filterMap queryCategories = queryCategories{}

var queryMapReverse queryMap = queryMap{}

var gameMap gameCatalogue = gameCatalogue{}

var client *http.Client = (&http.Client{Timeout: (time.Second * 1)})

var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)

var regexpFilterNonAlphaNumeric *regexp.Regexp = regexp.MustCompile(`[^a-zA-Z0-9]+`)

var regexpFilterWhitespace *regexp.Regexp = regexp.MustCompile(`\s{2,}`)

//var flagVerboseBool *bool = flag.Bool("verbose", false, "-v")

func scrapeGameCategory(d *goquery.Document) []gameCategory {
	a := d.Find("div.game_area_details_specs a.name")
	gameCategories := make([]gameCategory, a.Length())
	a.Each(func(i int, s *goquery.Selection) {
		gameCategories[i] = gameCategory{
			Name: strings.TrimSpace(s.Text()),
			URL:  strings.TrimSpace(s.AttrOr("href", "NIL"))}
	})
	return gameCategories
}

func scrapeGameDate(d *goquery.Document) string {
	date := strings.TrimSpace(d.Find("div.release_date div.date").First().Text())
	return date
}

func scrapeGameDescription(d *goquery.Document) string {
	description := strings.TrimSpace(d.Find("div.game_description_snippet").First().Text())
	return description
}

func scrapeGameDescriptionVerbose(d *goquery.Document) string {
	descriptionVerbose := strings.TrimSpace(d.Find("#game_area_description").First().Text())
	return descriptionVerbose
}

func scrapeGameDevelopers(d *goquery.Document) []gameDeveloper {
	a := d.Find("#developers_list a")
	gameDevelopers := make([]gameDeveloper, a.Length())
	a.Each(func(i int, s *goquery.Selection) {
		gameDevelopers[i] = gameDeveloper{
			Name: strings.TrimSpace(s.Text()),
			URL:  strings.TrimSpace(s.AttrOr("href", "NIL"))}
	})
	return gameDevelopers
}

func scrapeGameGenre(d *goquery.Document) []gameGenre {
	a := d.Find("div.game_details div.details_block:first-child > a")
	gameGenres := make([]gameGenre, a.Length())
	a.Each(func(i int, s *goquery.Selection) {
		gameGenres[i] = gameGenre{
			Name: strings.TrimSpace(s.Text()),
			URL:  strings.TrimSpace(s.AttrOr("href", "NIL"))}
	})
	return gameGenres
}

func scrapeGameLanguages(d *goquery.Document) []gameLanguage {
	tr := d.Find("table.game_language_options tr[class='']")
	gameLanguages := make([]gameLanguage, tr.Length())
	tr.Each(func(i int, s *goquery.Selection) {
		var (
			lang      = strings.TrimSpace(s.Find("td:nth-child(1)").Text())
			inter     = strings.TrimSpace(s.Find("td:nth-child(2)").Text())
			audio     = strings.TrimSpace(s.Find("td:nth-child(3)").Text())
			subtitles = strings.TrimSpace(s.Find("td:nth-child(4)").Text())
		)
		gameLanguage := gameLanguage{
			Audio:     (len(audio) != 0),
			Interface: (len(inter) != 0),
			Name:      lang,
			Subtitles: (len(subtitles) != 0)}
		gameLanguages[i] = gameLanguage
	})
	return gameLanguages
}

func scrapeGameMeta(d *goquery.Document) []gameMeta {
	m := d.Find("meta")
	gameMetaTags := make([]gameMeta, m.Length())
	m.Each(func(i int, s *goquery.Selection) {
		var (
			content  = s.AttrOr("content", "NIL")
			name     = s.AttrOr("name", "NIL")
			property = s.AttrOr("property", "NIL")
		)
		gameMeta := gameMeta{
			Content:  content,
			Name:     name,
			Property: property}
		gameMetaTags[i] = gameMeta
	})
	return gameMetaTags
}

func scrapeGamePublisher(d *goquery.Document) []gamePublisher {
	a := d.Find("div.dev_row > b:first-child + a")
	gamePublishers := make([]gamePublisher, a.Length())
	a.Each(func(i int, s *goquery.Selection) {
		gamePublishers[i] = gamePublisher{
			Name: strings.TrimSpace(s.Text()),
			URL:  s.AttrOr("href", "NIL")}
	})
	return gamePublishers
}

func scrapeGameTags(d *goquery.Document) []string {
	a := d.Find("a.app_tag")
	gameTags := make([]string, a.Length())
	a.Each(func(i int, s *goquery.Selection) {
		gameTags[i] = strings.TrimSpace(s.Text())
	})
	return gameTags
}

func scrapeGameTitle(d *goquery.Document) string {
	title := strings.TrimSpace(d.Find("div.apphub_AppName").First().Text())
	return title
}

func scrapeGameRequirements(d *goquery.Document) []gameRequirement {
	gameRequirements := []gameRequirement{}
	d.Find("div.game_area_sys_req[data-os]").Each(func(_ int, s *goquery.Selection) {
		reg := regexp.MustCompile(`[^a-zA-Z]+`)
		gameRequirement := gameRequirement{
			Name: strings.TrimSpace(s.AttrOr("data-os", "NIL"))}
		s.Find("ul.bb_ul").First().Each(func(i int, s *goquery.Selection) {
			m := map[string]string{}
			s.Find("li").Each(func(j int, s *goquery.Selection) {
				key := s.Find("strong").First().Text()
				key = reg.ReplaceAllString(key, "")
				key = strings.ToLower(key)
				m[key] = strings.TrimSpace(s.Text())
			})
			b, err := json.Marshal(m)
			if err != nil {
				panic(err)
			}
			if err := json.Unmarshal(b, &gameRequirement); err != nil {
				panic(err)
			}
			gameRequirements = append(gameRequirements, gameRequirement)
		})
	})
	return gameRequirements
}

func scrapeGamePage(d *goquery.Document) game {
	ID := d.Url.String()
	game, ok := gameMap.Get(ID)
	if ok != true {
		panic(fmt.Sprintf("game not found! %s", ID))
	}
	game.Categories = scrapeGameCategory(d)
	game.Description = scrapeGameDescription(d)
	game.DescriptionVerbose = scrapeGameDescriptionVerbose(d)
	game.Developer = scrapeGameDevelopers(d)
	game.Genre = scrapeGameGenre(d)
	game.Languages = scrapeGameLanguages(d)
	game.Meta = scrapeGameMeta(d)
	game.Publisher = scrapeGamePublisher(d)
	game.ReleaseDate = scrapeGameDate(d)
	game.Requirements = scrapeGameRequirements(d)
	game.Title = scrapeGameTitle(d)
	game.Tags = scrapeGameTags(d)
	if ok := gameMap.Add(game); ok != true {
		panic(fmt.Sprintf("game not added to map! %s", ID))
	}
	return game
}

func scrapePageItemHrefAttribute(s *goquery.Selection) string {
	href, exists := s.Attr("href")
	if exists == true {
		hrefGroup = append(hrefGroup, href)
	}
	return href
}

func scrapePageItemAppIDAttribute(s *goquery.Selection) string {
	ID := strings.TrimSpace(s.AttrOr("data-ds-appid", "NIL"))
	return ID
}

func scrapePageItemBundleIDAttribute(s *goquery.Selection) string {
	ID := strings.TrimSpace(s.AttrOr("data-ds-bundleid", "NIL"))
	return ID
}

func scrapePageItemCrtrIDAttribute(s *goquery.Selection) string {
	ID := strings.TrimSpace(s.AttrOr("data-ds-crtrids", "NIL"))
	return ID
}

func scrapePageItemDescIDAttribute(s *goquery.Selection) string {
	ID := strings.TrimSpace(s.AttrOr("data-ds-descids", "NIL"))
	return ID
}

func scrapePageItemPackageIDAttribute(s *goquery.Selection) string {
	ID := strings.TrimSpace(s.AttrOr("data-ds-packageid", "NIL"))
	return ID
}

func scrapePageItemTagIDAttribute(s *goquery.Selection) string {
	ID := strings.TrimSpace(s.AttrOr("data-ds-tagids", "NIL"))
	return ID
}

func scrapePageItemTitle(s *goquery.Selection) string {
	title := strings.TrimSpace(s.Find("div.search_name span.title").Text())
	return title
}

func scrapePageItem(s *goquery.Selection) game {
	game := game{
		AppID:         scrapePageItemAppIDAttribute(s),
		BundleID:      scrapePageItemBundleIDAttribute(s),
		CrtrID:        scrapePageItemCrtrIDAttribute(s),
		DescriptionID: scrapePageItemDescIDAttribute(s),
		Name:          scrapePageItemTitle(s),
		PackageID:     scrapePageItemPackageIDAttribute(s),
		TagID:         scrapePageItemTagIDAttribute(s),
		URL:           scrapePageItemHrefAttribute(s)}
	gameMap[game.URL] = game
	return game
}

func scrapeStoreCategories(s *goquery.Selection) {
	tag, ok := s.Attr("data-param")
	if ok != true {
		return
	}
	if ok := strings.ToUpper(tag) == "HIDE"; ok {
		return
	}
	value, ok := s.Attr("data-value")
	if ok != true {
		return
	}
	key, ok := s.Attr("data-loc")
	if ok != true {
		return
	}
	if ok := filterMap.Add(tag, key, value); ok != true {
		panic(fmt.Sprintf("filter map did not receive lookup keyset! %s", tag))
	}
	if ok := queryMapReverse.Add(key, tag); ok != true {
		panic(fmt.Sprintf("option map did not receive reverse lookup key! %s->%s", key, tag))
	}
}

func netrunnerGamePages(c chan string) {
	defer wg.Done()
	req, err := http.NewRequest(http.MethodGet, <-c, nil)
	if err != nil {
		return
	}
	req.Header.Set("Cookie", "birthtime=-949485599; lastagecheckage=1-0-1900; wants_mature_content=1")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	if res.StatusCode != http.StatusOK {
		return
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return
	}
	scrapeGamePage(doc)
}

func netrunnerStorePages(c chan string) string {
	defer wg.Done()
	req, err := http.NewRequest(http.MethodGet, <-c, nil)
	if err != nil {
		return "ERR"
	}
	res, err := client.Do(req)
	if err != nil {
		return "ERR"
	}
	if res.StatusCode != http.StatusOK {
		return "ERR"
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return "ERR"
	}
	a := doc.Find("a.search_result_row")
	a.Each(func(i int, s *goquery.Selection) {
		scrapePageItem(s)
	})
	if ok := a.Length() == 0; ok {
		return "EMPTY"
	}
	return "OK"
}

func netrunnerStoreCategories(URL string) {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	if res.StatusCode != http.StatusOK {
		return
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return
	}
	s := doc.Find("#additional_search_options div.tab_filter_control")
	s.Each(func(i int, s *goquery.Selection) {
		scrapeStoreCategories(s)
	})
}

func fPrintlnGame(w *tabwriter.Writer, game game) {
	s := reflect.ValueOf(&game).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Fprintln(w, fmt.Sprintf("%s\t|%v", typeOfT.Field(i).Name, f.Interface()))
	}
	//"%d: %s %s = %v\n", i, typeOfT.Field(i).Name, f.Type(), f.Interface()
}

func fPrintlnStoreFilter(w *tabwriter.Writer, queryCategories queryCategories) {
	var i int
	for tag := range queryCategories {
		i = (i + 1)
		fmt.Fprintln(w, fmt.Sprintf("%v %s", i, normalizeMapKey(tag)))
		for key := range queryCategories[tag] {

			fmt.Println(fmt.Sprintf("\t%s", normalizeMapKey(key)))
		}
		fmt.Println("")
	}
}

func normalizeMapKey(key string) string {
	key = strings.TrimSpace(key)
	key = regexpFilterNonAlphaNumeric.ReplaceAllString(key, " ")
	key = regexpFilterWhitespace.ReplaceAllString(key, " ")
	key = strings.TrimSpace(key)
	key = strings.ReplaceAll(strings.ToUpper(key), " ", "-")
	return key
}

func parseUserSearchQueryInput(input string) {
	for _, s := range regexp.MustCompile(`(\s|\,|\|)`).Split(input, -1) {
		key := strings.TrimSpace(s)
		key = strings.ToUpper(key)
		_, ok := queryMapReverse.Get(key)
		if ok != true {
			fmt.Println(fmt.Sprintf("tag not found using hash %s", key))
		}
	}
}

func main() {
	flag.Parse()
	netrunnerStoreCategories(steamSearchURL)
	w := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 0, '\t', 0)
	//if *flagVerboseBool != false {
	fPrintlnStoreFilter(w, filterMap)
	//}
	fmt.Println("Steamer.exe\t>\tinput N filters for search")
	if ok := scanner.Scan(); !ok {
		return
	}
	categoryOptions := strings.TrimSpace(scanner.Text())

	parseUserSearchQueryInput(categoryOptions)

	fmt.Println("Steamer.exe\t>\tinput N pages to search")
	if ok := scanner.Scan(); !ok {
		return
	}
	n, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return
	}
	fmt.Println(fmt.Sprintf("Steamer.exe\t>\tcollecting %d pages", n))
	if err := w.Flush(); err != nil {
		panic(err)
	}
	c := make(chan string, n)
	hrefGroup = []string{}
	for i := 1; i < n+1; i++ {
		wg.Add(1)
		c <- fmt.Sprintf("%s?page=%d", steamSearchURL, i)
		switch netrunnerStorePages(c) {
		case "EMPTY":
			fmt.Println(fmt.Sprintf("Steamer.exe\t>\tnothing more to process"))
			break
		case "ERR":
			fmt.Println(fmt.Sprintf("Steamer.exe\t>\terr for %d", i))
		case "OK":
			fmt.Println(fmt.Sprintf("Steam.exe\t>\tpage %d is OK", i))
		}
	}
	wg.Wait()
	close(c)
	fmt.Println(fmt.Sprintf("Steamer.exe\t>\tfound %d games", len(hrefGroup)))
	c = make(chan string, len(hrefGroup))
	for _, href := range hrefGroup {
		wg.Add(1)
		c <- href
		netrunnerGamePages(c)
	}
	wg.Wait()
	close(c)
	fmt.Println(fmt.Sprintf("Steamer.exe\t>\tbuilt %d games", len(gameMap)))
	w = new(tabwriter.Writer).Init(os.Stdout, 0, 8, 0, '\t', 0)
	for _, game := range gameMap {
		fPrintlnGame(w, game)
		fmt.Fprintln(w, "")
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
}
