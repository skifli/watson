package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"
)

var args struct {
	Username string `arg:"positional" help:"The username to check for."`

	Sites         string `help:"The file containing the sites to search." default:"./sites.json"`
	Colourless    bool   `help:"Disables coloured output." default:"false"`
	PrintAll      bool   `help:"Print all sites, even ones which matches are not found for." default:"false"`
	ReadTimeout   int    `help:"Timeout for reading request response (in milliseconds)." default:"500"`
	WriteTimeout  int    `help:"Timeout for writing request (in milliseconds)." default:"500"`
	ReqsPerThread int    `help:"The amount of requests per thread. Can significantly increase or decrease speed." default:"3"`
}

var colours struct {
	Red    string
	Yellow string
	Green  string
	Reset  string
}

var channel = make(chan bool)

var client = fasthttp.Client{
	ReadBufferSize:                8192,
	ReadTimeout:                   time.Duration(args.ReadTimeout) * time.Millisecond,
	WriteTimeout:                  time.Duration(args.WriteTimeout) * time.Millisecond,
	NoDefaultUserAgentHeader:      true,
	DisableHeaderNamesNormalizing: true,
	DisablePathNormalizing:        true,
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func checkSites(username string, sites []map[string]string) {
	for _, site := range sites {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		url := site["url"] + username
		req.SetRequestURI(url)
		req.Header.SetMethod(fasthttp.MethodGet)

		check(client.Do(req, resp))

		successful := false

		if site["type"] == "statusCode" {
			statusCode := resp.StatusCode()

			if 199 < statusCode && 300 > statusCode {
				successful = true
			}
		} else if site["type"] == "text" {
			if !strings.Contains(string(resp.Body()), site["text"]) {
				successful = true
			}
		}

		if successful {
			fmt.Printf("[%s+%s] %s%s%s: %s\n", colours.Green, colours.Reset, colours.Green, site["name"], colours.Reset, url)
			channel <- true
		} else {
			if args.PrintAll {
				fmt.Printf("[%s-%s] %s%s%s\n", colours.Yellow, colours.Reset, colours.Yellow, site["name"], colours.Reset)
			}

			channel <- false
		}
	}
}

func main() {
	arg.MustParse(&args)

	if !args.Colourless {
		colours.Red = "\u001b[31m"
		colours.Yellow = "\u001b[33m"
		colours.Green = "\u001b[32m"
		colours.Reset = "\u001b[0m"
	}

	file, err := os.ReadFile(args.Sites)

	if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("%serror: %ssites file not found", colours.Red, colours.Reset)

		os.Exit(1)
	} else {
		check(err)
	}

	var sitesJSON map[string][]map[string]string

	json.Unmarshal(file, &sitesJSON)
	sites := sitesJSON["sites"]

	sitesLen := len(sites)
	overflow := sitesLen % args.ReqsPerThread
	threadsNum := ((sitesLen - overflow) / args.ReqsPerThread)

	if overflow != 0 {
		threadsNum++
	}

	start := time.Now()

	for i := 0; i < sitesLen; i += args.ReqsPerThread {
		if i+args.ReqsPerThread > sitesLen {
			go checkSites(args.Username, sites[i:sitesLen])
		} else {
			go checkSites(args.Username, sites[i:i+args.ReqsPerThread])
		}
	}

	successful := 0

	for i := 0; i < sitesLen; i++ {
		if <-channel {
			successful++
		}
	}

	fmt.Printf("\nFound %d / %d matches (%.2f%%) in %.2f seconds.\n", successful, sitesLen, float64(successful)/float64(sitesLen)*100, time.Since(start).Seconds())
}
