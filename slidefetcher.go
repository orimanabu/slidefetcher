package main

import (
	// "fmt"
	"os"
	"strings"
	"time"
	"encoding/json"
	"github.com/gocolly/colly"
	// "github.com/gocolly/colly/debug"
)

const (
	// targetDomain = "kccnceu2021.sched.com"
	// targetDomain = "kccnceu2023.sched.com"
	targetDomain = "kccnceu2024.sched.com"
	// targetURL = "https://" + targetDomain + "/?iframe=no"
	targetURL = "https://" + targetDomain + "/?searchstring=CRI-O&iframe=no"
	selectorSession = "#sched-content-inner .list-simple .sched-container-inner .event a"
	selectorPDF = ".file-uploaded-pdf"
	selectorFile = ".file-uploaded"
	selectorVideo1 = ".sched-event-details iframe" // 2022-
	selectorVideo2 = ".sched-button" // 2021
	selectorTitle = ".list-single__event a"
)

type Conference struct {
	URL string `json:"url"`
	Name string `json:"name"`
	Sessions []Session `json:"sessions"`
}

type Session struct {
    Name   string `json:"name"`
    URL     string `json:"url"`
    Attachments []Attachment
    Video   string `json:"video"`
}

type Attachment struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL string `json:"url"`
}

var pageCount int = 0

func existsAttachment(s Session, at Attachment) bool {
	for _, v := range s.Attachments {
		if v.URL == at.URL {
			// fmt.Printf("<existsAttachment> %s (%s, %s)\n", v.Name, v.Type, at.Type)
			return true
		}
	}
	return false
}

func setupAttachment(e *colly.HTMLElement, typeStr string) Attachment {
	at := Attachment{}
	at.Name = strings.TrimSpace(e.Text)
	at.URL = e.Attr("href")
	at.Type = typeStr
	return at
}

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains(targetDomain),
		colly.CacheDir("./.cache"),
//		colly.Debugger(&debug.LogDebugger{}),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob: targetDomain,
		Delay: time.Second,
		RandomDelay: time.Second,
	})

	conf := Conference{}
	s := Session{}

	c.OnRequest(func(r *colly.Request) {
		if (conf.URL == "") {
			conf.URL = r.URL.String()
		}
		// fmt.Println("Visiting: ", r.URL)
		// fmt.Println("")
	})

	// c.OnError(func(_ *colly.Response, err error) {
	// 	fmt.Println("Something went wrong: ", err)
	// 	fmt.Println("")
	// })

	// c.OnResponse(func(r *colly.Response) {
	// 	pageCount++
	// 	// fmt.Println("Page visited: ", r.Request.URL)
	// 	// fmt.Println("")
	// })

	c.OnHTML("html title", func(e *colly.HTMLElement) {
		if (conf.Name == "") {
			conf.Name = e.Text
		}
	})

	c.OnHTML(selectorTitle, func(e *colly.HTMLElement) {
		s.Name = strings.TrimSpace(e.Text)
		// fmt.Println("**", s.Video)
		// fmt.Println("")
	})

	c.OnHTML(selectorPDF, func(e *colly.HTMLElement) {
		at := setupAttachment(e, "slides")
		if !existsAttachment(s, at) {
			s.Attachments = append(s.Attachments, at)
		}
	})

	c.OnHTML(selectorFile, func(e *colly.HTMLElement) {
		at := setupAttachment(e, "other")
		if !existsAttachment(s, at) {
			s.Attachments = append(s.Attachments, at)
		}
	})

	c.OnHTML(selectorVideo1, func(e *colly.HTMLElement) {
		// fmt.Println("**", s.Video)
		// fmt.Println("")
		s.Video = strings.Replace(e.Attr("src"), "embed/", "watch?v=", 1)
	})

	c.OnHTML(selectorVideo2, func(e *colly.HTMLElement) {
		if e.Text == "LINK TO VIDEO RECORDING" {
			e.ForEach("a", func(_ int, el *colly.HTMLElement) {
				s.Video = el.Attr("href")
			})
		}
	})

	c.OnHTML(selectorSession, func(e *colly.HTMLElement) {
		s = Session{}
		href := e.Attr("href")
		// fmt.Println("*", e.Text)
		// fmt.Println("*", href)
		s.URL = "https://" + targetDomain + "/" + href + "?iframe=no"
		e.Request.Visit(s.URL)
	})

	c.OnScraped(func(r *colly.Response) {
		// fmt.Println(r.Request.URL, " scraped!")
		if r.Request.URL.String() != targetURL {
			conf.Sessions = append(conf.Sessions, s)
		}
	})

	c.Visit(targetURL)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(conf)
}
