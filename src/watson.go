package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"
)

const VERSION = "v2.2"

var args struct {
	Colourless    bool     `arg:"-c" help:"Disables coloured output." default:"false"`
	OutputFolder  string   `arg:"-o" help:"Folder name in which to dump outputs to. Files will be named according to the account's username. Set to an empty string to disable." default:"results"`
	ShowAll       bool     `arg:"-a" help:"Show all sites, even ones which matches are not found for. Also applies to the output file." default:"false"`
	ReadTimeout   int      `help:"The amount of requests per thread. Can significantly increase or decrease speed." default:"3"`
	ReqsPerThread int      `help:"The amount of requests per thread. Can significantly increase or decrease speed." default:"3"`
	Sites         string   `arg:"-s" help:"JSON file to load data from a. Can be local or a URL." default:"https://raw.githubusercontent.com/skifli/watson/main/src/sites.json"`
	Username      string   `arg:"positional" help:"The username to check for."`
	Usernames     []string `arg:"-u" help:"Used to check for multiple usernames. Has a high precedence than 'username'."`
	Version       bool     `arg:"-v" help:"Display the current version of the program and exit."`
	WriteTimeout  int      `help:"Timeout for writing request (in milliseconds)." default:"1000"`
}

var colours struct {
	Red    string
	Yellow string
	Green  string
	Reset  string
}

var client = fasthttp.Client{
	ReadBufferSize:                8192,
	ReadTimeout:                   time.Duration(args.ReadTimeout) * time.Millisecond,
	WriteTimeout:                  time.Duration(args.WriteTimeout) * time.Millisecond,
	NoDefaultUserAgentHeader:      true,
	DisableHeaderNamesNormalizing: true,
	DisablePathNormalizing:        true,
}

var results = make(chan string)

var sites []map[string]string
var sitesLen int

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func isUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func checkSites(results map[string][]map[string]string, channel chan bool, username string, replacer *strings.Replacer, sites []map[string]string) {
	for _, site := range sites {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		url := replacer.Replace(site["url"])

		req.SetRequestURI(url)
		req.Header.SetMethod(fasthttp.MethodGet)

		check(client.Do(req, resp))

		statusCode := resp.StatusCode()

		if 199 < statusCode && 300 > statusCode {
			fmt.Printf("[%s+%s] %s%s%s: %s\n", colours.Green, colours.Reset, colours.Green, site["name"], colours.Reset, url)
			channel <- true
			results["results"] = append(results["results"], map[string]string{"found": "true", "name": site["name"], "url": url})
		} else {
			if args.ShowAll {
				fmt.Printf("[%s-%s] %s%s%s\n", colours.Yellow, colours.Reset, colours.Yellow, site["name"], colours.Reset)
				results["results"] = append(results["results"], map[string]string{"found": "false", "name": site["name"], "url": url})
			}

			channel <- false
		}
	}
}

func checkUsername(username string) {
	start := time.Now()
	channel := make(chan bool)
	resultsMap := map[string][]map[string]string{"results": {}}
	replacer := strings.NewReplacer("{}", username)

	for i := 0; i < sitesLen; i += args.ReqsPerThread {
		if i+args.ReqsPerThread > sitesLen {
			go checkSites(resultsMap, channel, username, replacer, sites[i:sitesLen])
		} else {
			go checkSites(resultsMap, channel, username, replacer, sites[i:i+args.ReqsPerThread])
		}
	}

	successful := 0

	for i := 0; i < sitesLen; i++ {
		if <-channel {
			successful++
		}
	}

	if args.OutputFolder != "" {
		resultsJSON, err := json.Marshal(resultsMap)
		check(err)

		check(os.MkdirAll(args.OutputFolder, 0777))
		check(os.WriteFile(path.Join(args.OutputFolder, username+".json"), resultsJSON, 0777))
	}

	results <- fmt.Sprintf("\nFound %d / %d matches (%.2f%%) for '%s' in %.2f seconds.", successful, sitesLen, float64(successful)/float64(sitesLen)*100, username, time.Since(start).Seconds())
}

func main() {
	parser := arg.MustParse(&args)

	if args.Version {
		fmt.Printf("watson %s\n", VERSION)
		os.Exit(0)
	}

	if !args.Colourless {
		colours.Red = "\u001b[31m"
		colours.Yellow = "\u001b[33m"
		colours.Green = "\u001b[32m"
		colours.Reset = "\u001b[0m"
	}

	if len(args.Usernames) == 0 {
		if args.Username == "" {
			parser.Fail(fmt.Sprintf("%serror:%s empty username.\n", colours.Red, colours.Reset))
		}

		args.Usernames = append(args.Usernames, args.Username)
	} else {
		if args.Username != "" {
			parser.Fail(fmt.Sprintf("%serror:%s cannot combine 'username' and 'usernames' flag.\n", colours.Red, colours.Reset))
		}
	}

	var err error
	var sitesRaw []byte

	if isUrl(args.Sites) {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		req.SetRequestURI(args.Sites)
		req.Header.SetMethod(fasthttp.MethodGet)

		err = client.Do(req, resp)

		if err == nil {
			sitesRaw = resp.Body()
		}
	} else {
		sitesRaw, err = os.ReadFile(args.Sites)
	}

	if err != nil {
		fmt.Printf("%serror:%s failed to get sites: '%s'\n", colours.Red, colours.Reset, err.Error())
		os.Exit(1)
	}

	var sitesJSON map[string][]map[string]string

	json.Unmarshal(sitesRaw, &sitesJSON)
	sites = sitesJSON["sites"]
	sitesLen = len(sites)

	for _, username := range args.Usernames {
		go checkUsername(username)
	}

	var resultsSlice []string

	for range args.Usernames {
		resultsSlice = append(resultsSlice, <-results)
	}

	for _, output := range resultsSlice {
		fmt.Print(output)
	}
}
