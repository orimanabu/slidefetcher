package main

import (
	"errors"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/gocolly/colly"
	"github.com/urfave/cli/v2"
	// "github.com/gocolly/colly/debug"
)

const (
	targetQuery = "/?iframe=no"
	// targetQuery = "/?searchstring=CRI-O&iframe=no"
	selectorSession = "#sched-content-inner .list-simple .sched-container-inner .event a"
	selectorPDF = ".file-uploaded-pdf"
	selectorFile = ".file-uploaded"
	selectorVideo1 = ".sched-event-details iframe" // 2022-
	selectorVideo2 = ".sched-button" // 2021
	selectorVideo3 = ".sched-button" // 2021
	selectorHTMLTitle = "html title"
	selectorTitle = ".list-single__event a"
	selectorDescription = ".tip-description"
)

type KnownConf struct {
	Name string `json:"name"`
	Description string `json:"description"`
	URL string `json:"url"`
}

var knownConfs = []KnownConf{
	{
		"kccnceu2021",
		"KubeCon + CloudNativeCon Europe 2021",
		"https://kccnceu2021.sched.com",
	},
	{
		"kccnceu2022",
		"KubeCon + CloudNativeCon Europe 2022",
		"https://kccnceu2022.sched.com",
	},
	{
		"kccnceu2023",
		"KubeCon + CloudNativeCon Europe 2023",
		"https://kccnceu2023.sched.com",
	},
	{
		"kccnceu2024",
		"KubeCon + CloudNativeCon Europe 2024",
		"https://kccnceu2024.sched.com",
	},
	{
		"kccncna2017",
		"KubeCon + CloudNative North America 2017",
		"https://kccncna17.sched.com",
	},
	{
		"kccncna2018",
		"KubeCon + CloudNative North America 2018",
		"https://kccna18.sched.com",
	},
	{
		"kccncna2019",
		"KubeCon + CloudNative North America 2019",
		"https://kccncna19.sched.com",
	},
	{
		"kccncna2020",
		"KubeCon + CloudNative North America 2020",
		"https://kccncna20.sched.com",
	},
	{
		"kccncna2021",
		"KubeCon + CloudNative North America 2021",
		"https://kccncna2021.sched.com",
	},
	{
		"kccncna2022",
		"KubeCon + CloudNative North America 2022",
		"https://kccncna2022.sched.com",
	},
	{
		"kccncna2023",
		"KubeCon + CloudNative North America 2023",
		"https://kccncna2023.sched.com",
	},
	{
		"kccncna2024",
		"KubeCon + CloudNative North America 2024",
		"https://kccncna2024.sched.com",
	},
}

type Conference struct {
	URL string `json:"url"`
	Name string `json:"name"`
	Sessions []Session `json:"sessions"`
}

type Session struct {
    Name   string `json:"name"`
    URL     string `json:"url"`
    Description string `json:"description"`
    Attachments []Attachment `json:"attachment"`
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

func prepare(targetURL string) {
	c := colly.NewCollector(
		// colly.AllowedDomains(targetDomain),
		colly.CacheDir("./.cache"),
		// colly.Debugger(&debug.LogDebugger{}),
	)
	c.Limit(&colly.LimitRule{
		// DomainGlob: targetDomain,
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

	c.OnHTML(selectorHTMLTitle, func(e *colly.HTMLElement) {
		// fmt.Printf("<OnHTML-selectorHTMLTitle> [%s]\n", e.Text)
		if (conf.Name == "") {
			conf.Name = e.Text
		}
	})

	c.OnHTML(selectorTitle, func(e *colly.HTMLElement) {
		// fmt.Printf("<OnHTML-selectorTitle> [%s]\n", e.Text)
		s.Name = strings.TrimSpace(e.Text)
		// fmt.Println("**", s.Video)
		// fmt.Println("")
	})

	c.OnHTML(selectorDescription, func(e *colly.HTMLElement) {
		// fmt.Printf("<OnHTML-selectorDescription> [%s]\n", e.Text)
		desc := e.Text
		// if strings.HasPrefix("\n  ", desc) {
		// 	desc = desc[3:]
		// }
		s.Description = strings.TrimSpace(desc)
	})

	c.OnHTML(selectorPDF, func(e *colly.HTMLElement) {
		// fmt.Printf("<OnHTML-selectorPDF> [%s]\n", e.Text)
		at := setupAttachment(e, "slides")
		if !existsAttachment(s, at) {
			s.Attachments = append(s.Attachments, at)
		}
	})

	c.OnHTML(selectorFile, func(e *colly.HTMLElement) {
		// fmt.Printf("<OnHTML-selectorFile> [%s]\n", e.Text)
		at := setupAttachment(e, "other")
		if !existsAttachment(s, at) {
			s.Attachments = append(s.Attachments, at)
		}
	})

	c.OnHTML(selectorVideo1, func(e *colly.HTMLElement) {
		// fmt.Printf("<OnHTML-selectorVideo1> [%s]\n", e.Text)
		s.Video = strings.Replace(e.Attr("src"), "embed/", "watch?v=", 1)
	})

	c.OnHTML(selectorVideo2, func(e *colly.HTMLElement) {
		// fmt.Printf("<OnHTML-selectorVideo2> [%s]\n", e.Text)
		if e.Text == "LINK TO VIDEO RECORDING" {
			e.ForEach("a", func(_ int, el *colly.HTMLElement) {
				s.Video = el.Attr("href")
			})
		}
	})

	c.OnHTML(selectorSession, func(e *colly.HTMLElement) {
		s = Session{}
		href := e.Attr("href")
		// fmt.Printf("<OnHTML-selectorSession> [%s][%s]\n", e.Text, href)
		s.URL = targetURL + "/" + href + "?iframe=no"
		e.Request.Visit(s.URL)
	})

	c.OnScraped(func(r *colly.Response) {
		if r.Request.URL.String() != targetURL + targetQuery {
			conf.Sessions = append(conf.Sessions, s)
		}
		// fmt.Printf("<OnScraped> [%s][%s] (%d)\n", r.Request.URL, targetURL + targetQuery, len(conf.Sessions))
	})

	c.Visit(targetURL + targetQuery)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(conf)
}

func list(verbose bool) {
	if verbose {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(knownConfs)
	} else {
		for _, v := range knownConfs {
			fmt.Println(v.Name)
		}
	}
}

func name2url(name string) (string, error) {
	for _, v := range knownConfs {
		if v.Name == name {
			return v.URL, nil
		}
	}
	return "", fmt.Errorf("no such name: %s", name)
}

func downloadFile(url, filepath string) error {
	// file already exists
	if _, err := os.Stat(filepath); err == nil {
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name: "list",
				Aliases: []string{"l"},
				Usage: "list the pre-defined conferences",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}},
				},
				Action: func(cCtx *cli.Context) error {
					verbose := false
					if cCtx.Bool("verbose") {
						verbose = true
					}
					list(verbose)
					return nil
				},
			},
			{
				Name: "prepare",
				Aliases: []string{"p"},
				Usage: "Prepare for download",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Args().Len() != 1 {
						return errors.New("prepare gets 1 arg")
					}
					targetURL, err := name2url(cCtx.Args().Get(0))
					if err != nil {
						return err
					}
					prepare(targetURL)
					return nil
				},
			},
			{
				Name: "download",
				Aliases: []string{"d"},
				Usage: "Do download",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("XXX not implemented")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
