package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"regexp"
	"strings"
	"time"
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

var (
	quoteTextRe     = regexp.MustCompile(`(?s)<div[^>]*class="[^"]*\bquoteText\b[^"]*"[^>]*>(.*?)</div>`)
	tagRe           = regexp.MustCompile(`<[^>]*>`)
	wsRe            = regexp.MustCompile(`\s+`)
	dashRe          = regexp.MustCompile(`\s*[―—–-]+\s*`)
	leadingQuoteRe  = regexp.MustCompile(`^[\s"“”‘’'«»]+`)
	trailingQuoteRe = regexp.MustCompile(`[\s"“”‘’'«»]+$`)
)

func fetchData(url string, data any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Request %q: %w", url, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,application/json;q=0.8,*/*;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Fetch %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Fetch %q: status %d", url, resp.StatusCode)
	}

	switch v := data.(type) {
	case *string:
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Read %q body: %w", url, err)
		}
		*v = string(bodyBytes)
		return nil
	default:
		return json.NewDecoder(resp.Body).Decode(data)
	}
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

func getQU() {
	url := "https://www.goodreads.com/quotes"

	var body string
	if err := fetchData(url, &body); err != nil {
		log.Fatal(err)
	}

	raw := extractQuote(body, quoteTextRe)
	if raw == "" {
		log.Fatal("Quote not found")
	}

	quote, author := parseQuoteText(raw)
	text := fmt.Sprintf("[Goodreads] %s -- %s", quote, author)
	fmt.Println(text)
}

func extractQuote(raw string, re *regexp.Regexp) string {
	/*q := re.FindStringSubmatch(raw)
	if len(q) < 2 {
		return ""
	}
	return q[1]*/

	matches := re.FindAllStringSubmatch(raw, -1)
	l := len(matches)
	if l == 0 {
		return ""
	}

	return matches[rand.IntN(l)][1]
}

func parseQuoteText(rawHTML string) (string, string) {
	text := tagRe.ReplaceAllString(rawHTML, " ")
	text = html.UnescapeString(text)
	text = wsRe.ReplaceAllString(text, " ")

	idx := dashRe.FindStringIndex(text)
	var quotePart, authorPart string

	if idx != nil {
		quotePart = text[:idx[0]]
		authorPart = text[idx[1]:]
	} else {
		quotePart = text
	}

	quotePart = leadingQuoteRe.ReplaceAllString(quotePart, "")
	quotePart = trailingQuoteRe.ReplaceAllString(quotePart, "")
	quotePart = strings.TrimSpace(quotePart)
	authorPart = strings.TrimSpace(authorPart)

	return quotePart, authorPart
}

func main() {
	functions := []func(){
		rndHN,
		getQU,
		rndUD,
	}

	functions[rand.IntN(len(functions))]()
}
