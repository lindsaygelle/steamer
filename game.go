package main

// Game is a struct that expresses a video game that has been scraped from the Steam store. It
// contains properties and attributes that aim to provide a snapshot of the game's current
// state on the date it was collected. A Game is expected to be built by querying and
// reducing HTML elements found across the Steam game page. As this data is generally
// provided from a raw HTML document, and each field may have varying levels of artifacting.
// Most artifacting has been handled by various string mutation operations but inconsistencies
// may persist.
type game struct {
	AppID              string            `json:"appid"`              // {AppID: "306130"} OR {AppID: "NIL"} (NIL when BundleID)
	BundleID           string            `json:"bundleid"`           // {BundleID: "NIL"} OR {BundleID: "11802"} (NIL when AppID)
	Categories         []gameCategory    `json:"categories"`         // {Categories: [MMO Steam Trading Cards Partial Controller Support]}
	CrtrID             []string          `json:"crtrid"`             // {CrtrID: "[33028765,35501445]"}
	DescriptionID      []string          `json:"descriptionid"`      // {DescriptionID: "[2,5]"}
	Description        string            `json:"description"`        // {Description: "...."}
	DescriptionVerbose string            `json:"descriptionverbose"` // {DescriptionVerbose: "...."}
	Developer          []gameDeveloper   `json:"developer"`          // {Developer: [Zenimax Online Studios]}
	Genre              []gameGenre       `json:"genre"`              // {Genre: [Massively Multiplayer RPG]}
	Languages          []gameLanguage    `json:"languages"`          // {Languages: [{English Audio true Interface true Subtitles true}]}
	Meta               []gameMeta        `json:"meta"`               // {Meta: [...]}
	Name               string            `json:"name"`               // {Name: "THE-ELDER-SCROLLS-ONLINE"}
	PackageID          string            `json:"packageid"`          // {PackageID: "124923"}
	Publisher          []gamePublisher   `json:"publisher"`          // {Publisher: [ZENIMAX-ONLINE-STUDIOS ...]}
	ReleaseDate        string            `json:"releasedate"`        // {ReleaseDate: "4 Apr, 2014"}
	Requirements       []gameRequirement `json:"requirements"`       // {Requirements: {DirectX: Version 11 Graphics: AMD Radeon RX 470 Memory: ...}]
	TagID              []string          `json:"tagid"`              // {TagID: "[19,1685,3814,29482,122,3859,21]"}
	Tags               []gameTag         `json:"tags"`               // {Tags: [Post-apocalyptic Difficult Survival Lovecraft ...]}
	Title              string            `json:"title"`              // {Title: "The Elder Scrolls® Online"}
	URL                string            `json:"url"`                // {URL: "https://store.steampowered.com/app/306130/The_Elder_Scrolls_Online/?snr=1_7_7_230_150_1"}
}