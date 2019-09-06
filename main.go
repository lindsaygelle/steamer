package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	i := 1
	n := 1
	wg := &sync.WaitGroup{}
	client := &http.Client{Timeout: time.Second * 10}
	steamerLog := &SteamerLog{
		PagesFrom: i,
		PagesTo:   n,
		PagesOK:   &SteamerLogPageOK{},
		TimeStart: time.Now()}
	for i := 1; i <= n; i++ {
		URL := fmt.Sprintf("store.steampowered.com/search/?page=%d", i)
		wg.Add(1)
		go func(client *http.Client, URL string) {
			defer wg.Done()
			onGetSteamGameAbbreviation(client, URL,
				func(s *Snapshot) {
					writeSnapshotDefault(s)
				},
				func(s *SteamGameAbbreviation) {
					onGetSteamGamePage(client, s.URL,
						func(s *Snapshot) {
							writeSnapshotDefault(s)
						},
						func(s *SteamGamePage) {
							fmt.Println("page\t", s.Name)
							onGetSteamChartPage(client, fmt.Sprintf("https://steamcharts.com/app/%d", s.AppID),
								func(s *Snapshot) {
									writeSnapshotDefault(s)
								},
								func(s *SteamChartPage) {
									fmt.Println("chart\t", s.Name)
								},
								func(e error) {

								})
						},
						func(e error) {
						})
				},
				func(e error) {
				})
		}(client, URL)
	}
	wg.Wait()
	steamerLog.TimeEnd = time.Now()
	steamerLog.TimeDuration = steamerLog.TimeEnd.Sub(steamerLog.TimeStart)
	writeSteamerLogDefault(steamerLog)
	fmt.Println("done!")
}
