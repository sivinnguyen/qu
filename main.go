package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

var httpClient *http.Client = &http.Client{
	Timeout: 10 * time.Second,
}

type Story struct {
	Title string `json:"title"`
	By    string `json:"by"`
	URL   string `json:"url"`
}

type Word struct {
	Word    string `json:"word"`
	Meaning string `json:"meaning"`
	Example string `json:"example"`
}

type Words struct {
	Data []Word `json:"data"`
}

type Quote struct {
	Text   string
	Author string
}

func getCacheDir() string {
	currentDir := "./cache"
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return currentDir
	}
	quCache := filepath.Join(cacheDir, "qu")
	err = os.MkdirAll(quCache, 0755)
	if err != nil {
		return currentDir
	}
	return quCache
}

func fetchData(url string, data any) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("Fetch %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Fetch %q: status %d", url, resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(data)
}

func rndHN() {
	url := "https://hacker-news.firebaseio.com/v0/topstories.json"

	var storyIDs []int
	if err := fetchData(url, &storyIDs); err != nil {
		log.Fatal(err)
	}

	storyID := storyIDs[rand.IntN(5)]
	storyUrl := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", storyID)

	var story Story
	if err := fetchData(storyUrl, &story); err != nil {
		log.Fatal(err)
	}

	title := fmt.Sprintf("[Hacker News] %s -- %s", story.Title, story.By)
	fmt.Println(title)
}

func rndUD() {
	url := "https://unofficialurbandictionaryapi.com/api/random?limit=1&"

	var words Words
	if err := fetchData(url, &words); err != nil {
		log.Fatal(err)
	}

	word := fmt.Sprintf("[Urban Dictionary] %s: %s", strings.TrimSpace(words.Data[0].Word), words.Data[0].Meaning)
	fmt.Println(word)
}

func rndQU() {
	var quotes []Quote
	cacheDir := getCacheDir()

	c := colly.NewCollector(
		colly.AllowedDomains("www.goodreads.com", "goodreads.com"),
		//colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36"),
		colly.UserAgent("Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1"),
		colly.Async(true),
		colly.CacheDir(cacheDir),
	)

	c.CacheExpiration = 24 * time.Hour

	/*c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting:", r.URL.String())
	})*/

	c.OnHTML(".quote", func(e *colly.HTMLElement) {
		//text := strings.TrimSpace(e.ChildText(".quoteBody"))

		text := formatQuote(e)
		author := strings.TrimSpace(e.ChildText(".quoteAuthor"))

		if text != "" {
			quotes = append(quotes, Quote{Text: text, Author: author})
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error: status=%d, err=%v", r.StatusCode, err)
	})

	if err := c.Visit("https://www.goodreads.com/quotes"); err != nil {
		log.Fatal(err)
	}

	c.Wait()

	if len(quotes) == 0 {
		log.Fatal("Quote not found")
	}

	pick := quotes[rand.IntN(len(quotes))]
	quote := fmt.Sprintf("[Goodreads] %s -- %s", pick.Text, pick.Author)
	fmt.Println(quote)
}

func formatQuote(e *colly.HTMLElement) string {
	sep := ". "
	tags := regexp.MustCompile(`<[^>]*>`)
	raw, _ := e.DOM.Find("blockquote.quoteBody").Html()

	// Replace <br> by separator
	replacer := strings.NewReplacer(
		"<br />", sep,
		"<br/>", sep,
		"<br>", sep,
	)
	raw = replacer.Replace(raw)

	// Strip remain tags
	text := tags.ReplaceAllString(raw, "")
	text = html.UnescapeString(text)
	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}

func main() {
	functions := []func(){
		rndHN,
		rndQU,
		rndUD,
	}

	functions[rand.IntN(len(functions))]()
}
