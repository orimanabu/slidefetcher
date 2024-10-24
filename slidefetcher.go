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
	targetQuery2 = "?iframe=no&fixedHeight=no&w=100%&sidebar=yes&bg=false&mobileoff=Y&ssl=yes"
	// targetQuery = "/?searchstring=CRI-O&iframe=no"
	// targetQuery = "/?searchstring=runtime&iframe=no"
	selectorSession = "#sched-content-inner .list-simple .sched-container-inner .event a"
	selectorPDF = ".file-uploaded-pdf"
	selectorFile = ".file-uploaded"
	selectorVideo1 = ".sched-event-details iframe" // 2022-
	selectorVideo2 = ".sched-button" // 2021
	selectorVideo3 = ".sched-event-type .tip-custom-fields a" // 2020 EU
	selectorHTMLTitle = "html title"
	selectorTitle = ".list-single__event a"
	selectorDescription = ".tip-description"
)

type KnownConf struct {
	Name string `json:"name"`
	Description string `json:"description"`
	URL string `json:"url"`
	QueryURL string `json:"queryurl"`
}

var knownConfs = []KnownConf{
	// {
	// 	"kubecon-eu-2016",
	// 	"KubeCon Europe 2016",
	// 	"https://kubeconeurope2016.sched.com/",
	// 	"https://kubeconeurope2016.sched.com/" + targetQuery2,
	// },
	// {
	// 	"kubecon-eu-2017",
	// 	"KubeCon + CloudNativeCon Europe 2017",
	// 	"https://cloudnativeeu2017.sched.com/",
	// 	"https://cloudnativeeu2017.sched.com/",
	// },
	{
		"kubecon-eu-2018",
		"KubeCon + CloudNativeCon Europe 2018",
		"https://kccnceu18.sched.com",
		"https://kccnceu18.sched.com" + targetQuery2,
	},
	{
		"kubecon-eu-2019",
		"KubeCon + CloudNativeCon Europe 2019",
		"https://kccnceu19.sched.com",
		"https://kccnceu19.sched.com" + targetQuery2,
	},
	{
		"kubecon-eu-2020",
		"KubeCon + CloudNativeCon Europe 2020",
		"https://kccnceu20.sched.com",
		"https://kccnceu20.sched.com" + targetQuery2,
	},
	{
		"kubecon-eu-2021",
		"KubeCon + CloudNativeCon Europe 2021",
		"https://kccnceu2021.sched.com",
		"https://kccnceu2021.sched.com" + targetQuery,
	},
	{
		"kubecon-eu-2022",
		"KubeCon + CloudNativeCon Europe 2022",
		"https://kccnceu2022.sched.com",
		"https://kccnceu2022.sched.com" + targetQuery,
	},
	{
		"kubecon-eu-2023",
		"KubeCon + CloudNativeCon Europe 2023",
		"https://kccnceu2023.sched.com",
		"https://kccnceu2023.sched.com" + targetQuery,
	},
	{
		"kubecon-eu-2024",
		"KubeCon + CloudNativeCon Europe 2024",
		"https://kccnceu2024.sched.com",
		"https://kccnceu2024.sched.com" + targetQuery,
	},
	// {
	// 	"kubecon-na-2015",
	// 	"KubeCon 2015",
	// 	"https://kubecon2015.sched.com",
	// 	"https://kubecon2015.sched.com" + targetQuery2,
	// },
	// {
	// 	"kubecon-na-2016",
	// 	"KubeCon + CloudNative North America 2016",
	// 	"https://cnkc16.sched.com/",
	// 	"https://cnkc16.sched.com/" + targetQuery2,
	// },
	// {
	// 	"kubecon-na-2017",
	// 	"KubeCon + CloudNative North America 2017",
	// 	"https://kccncna17.sched.com",
	// 	"https://kccncna17.sched.com" + targetQuery2,
	// },
	{
		"kubecon-na-2018",
		"KubeCon + CloudNative North America 2018",
		"https://kccna18.sched.com",
		"https://kccna18.sched.com" + targetQuery2,
	},
	{
		"kubecon-na-2019",
		"KubeCon + CloudNative North America 2019",
		"https://kccncna19.sched.com",
		"https://kccncna19.sched.com" + targetQuery2,
	},
	{
		"kubecon-na-2020",
		"KubeCon + CloudNative North America 2020",
		"https://kccncna20.sched.com",
		"https://kccncna20.sched.com" + targetQuery2,
	},
	{
		"kubecon-na-2021",
		"KubeCon + CloudNative North America 2021",
		"https://kccncna2021.sched.com",
		"https://kccncna2021.sched.com" + targetQuery,
	},
	{
		"kubecon-na-2022",
		"KubeCon + CloudNative North America 2022",
		"https://kccncna2022.sched.com",
		"https://kccncna2022.sched.com" + targetQuery,
	},
	{
		"kubecon-na-2023",
		"KubeCon + CloudNative North America 2023",
		"https://kccncna2023.sched.com",
		"https://kccncna2023.sched.com" + targetQuery,
	},
	{
		"kubecon-na-2024",
		"KubeCon + CloudNative North America 2024",
		"https://kccncna2024.sched.com",
		"https://kccncna2024.sched.com" + targetQuery,
	},
	// {
	// 	"test",
	// 	"test test",
	// 	"http://ringeye.jawfish.org",
	// 	"http://ringeye.jawfish.org/~ori/misc/leapsecond-20120701.html",
	// },
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
			// fmt.Fprintf(os.Stderr, "<existsAttachment> %s (%s, %s)\n", v.Name, v.Type, at.Type)
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

func prepare(targetURL, queryURL string) {
	var targetDomain string
	if strings.HasPrefix(targetURL, "http://") {
		targetDomain = targetURL[7:]
	} else if strings.HasPrefix(targetURL, "https://") {
		targetDomain = targetURL[8:]
	}
	// fmt.Fprintf(os.Stderr, "XXX %s %s %s\n", targetURL, queryURL, targetDomain)
	c := colly.NewCollector(
		colly.AllowedDomains(targetDomain),
		colly.CacheDir("./.cache"),
		// colly.Debugger(&debug.LogDebugger{}),
		// colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 14.7; rv:131.0) Gecko/20100101 Firefox/131.0"),
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
		colly.MaxBodySize(100 * 1024 * 1024),
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
		// fmt.Fprintf(os.Stderr, "<OnRequest> [%s]\n", conf.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Fprintln(os.Stderr, "<OnError> ", err)
	})

	c.OnResponse(func(r *colly.Response) {
		pageCount++
		// fmt.Fprintln(os.Stderr, "Page visited: ", r.Request.URL)
		// fmt.Fprintln("")
	})

	c.OnHTML(selectorHTMLTitle, func(e *colly.HTMLElement) {
		// fmt.Fprintf(os.Stderr, "<OnHTML-selectorHTMLTitle> [%s]\n", e.Text)
		if (conf.Name == "") {
			conf.Name = e.Text
		}
	})

	c.OnHTML(selectorTitle, func(e *colly.HTMLElement) {
		// fmt.Fprintf(os.Stderr, "<OnHTML-selectorTitle> [%s]\n", e.Text)
		s.Name = strings.TrimSpace(e.Text)
	})

	c.OnHTML(selectorDescription, func(e *colly.HTMLElement) {
		s.Description = e.Text
	})

	c.OnHTML(selectorPDF, func(e *colly.HTMLElement) {
		// fmt.Fprintf(os.Stderr"<OnHTML-selectorPDF> [%s]\n", e.Text)
		at := setupAttachment(e, "slides")
		if !existsAttachment(s, at) {
			s.Attachments = append(s.Attachments, at)
		}
	})

	c.OnHTML(selectorFile, func(e *colly.HTMLElement) {
		// fmt.Fprintf(os.Stderr, "<OnHTML-selectorFile> [%s]\n", e.Text)
		at := setupAttachment(e, "other")
		if !existsAttachment(s, at) {
			s.Attachments = append(s.Attachments, at)
		}
	})

	c.OnHTML(selectorVideo1, func(e *colly.HTMLElement) {
		// fmt.Fprintf(os.Stderr, "<OnHTML-selectorVideo1> [%s][%s]\n", s.Video, e.Attr("src"))
		videoURL := strings.Replace(e.Attr("src"), "embed/", "watch?v=", 1) 
		if strings.Contains(videoURL, "youtube") {
			s.Video = videoURL
		}
	})

	c.OnHTML(selectorVideo2, func(e *colly.HTMLElement) {
		// fmt.Fprintf(os.Stderr, "<OnHTML-selectorVideo2> [%s][%s]\n", s.Video, e.Text)
		if e.Text == "LINK TO VIDEO RECORDING" {
			e.ForEach("a", func(_ int, el *colly.HTMLElement) {
				videoURL := el.Attr("href")
				if strings.Contains(videoURL, "youtube") {
					s.Video = videoURL
				}
			})
		}
	})

	c.OnHTML(selectorVideo3, func(e *colly.HTMLElement) {
		// fmt.Fprintf(os.Stderr, "<OnHTML-selectorVideo3> [%s][%s]\n", s.Video, e.Text)
		videoURL := strings.Replace(e.Attr("href"), "youtu.be/", "www.youtube.com/watch?v=", 1)
		if strings.Contains(videoURL, "youtube") {
			s.Video = videoURL
		}
	})

	c.OnHTML(selectorSession, func(e *colly.HTMLElement) {
		s = Session{}
		href := e.Attr("href")
		// fmt.Fprintf(os.Stderr, "<OnHTML-selectorSession> [%s][%s]\n", e.Text, href)
		s.URL = targetURL + "/" + href + "?iframe=no"
		e.Request.Visit(s.URL)
	})

	c.OnScraped(func(r *colly.Response) {
		if r.Request.URL.String() != queryURL {
			conf.Sessions = append(conf.Sessions, s)
		}
		// fmt.Fprintf(os.Stderr, "<OnScraped> [%s][%s] (%d)\n", r.Request.URL, queryURL, len(conf.Sessions))
	})

	c.Visit(queryURL)
	c.Wait()
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

func name2url(name string) (string, string, error) {
	for _, v := range knownConfs {
		if v.Name == name {
			return v.URL, v.QueryURL, nil
		}
	}
	return "", "", fmt.Errorf("no such name: %s", name)
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
					targetURL, queryURL, err := name2url(cCtx.Args().Get(0))
					if err != nil {
						return err
					}
					prepare(targetURL, queryURL)
					return nil
				},
			},
			{
				Name: "download",
				Aliases: []string{"d"},
				Usage: "Do download",
				Action: func(cCtx *cli.Context) error {
					fmt.Fprintln(os.Stderr, "XXX not implemented")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
