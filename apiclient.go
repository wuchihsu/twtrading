package twtrading

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const (
	urlFutContracts = "https://www.taifex.com.tw/enl/eng3/futContractsDateDown"
	dateFormat      = "2006/01/02"
)

var (
	rgx = regexp.MustCompile(`alert\("(.*?)"\)`)
)

// APIClient represents an HTTP API client for Taiwan stocks.
type APIClient struct {
	HTTPCli *http.Client
}

// MTXFutContracts gets statistics of Mini-TAIEX Futures Contracts.
func (c *APIClient) MTXFutContracts(startDate, endDate *time.Time) ([][]string, error) {
	resp, err := c.HTTPCli.PostForm(
		urlFutContracts,
		url.Values{
			"firstDate":      {"2017/06/14 00:00"},
			"lastDate":       {"2020/06/14 00:00"},
			"queryStartDate": {startDate.Format(dateFormat)},
			"queryEndDate":   {endDate.Format(dateFormat)},
			"commodityId":    {"MXF"},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("post form: %s", err)
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK {
		return nil, fmt.Errorf("status code: %d %s", code, http.StatusText(code))
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read all: %s", err)
	}

	csvReader := csv.NewReader(bytes.NewReader(respBody))
	records, err := csvReader.ReadAll()
	if err != nil {
		alert := rgx.FindSubmatch(respBody)
		if alert == nil {
			return nil, fmt.Errorf("response body is not CSV and has no alert: %s", string(respBody))
		}
		if bytes.Equal(alert[1], []byte("no data")) {
			return nil, ErrNoData
		}
		return nil, fmt.Errorf(string(alert[1]))
	}

	return records, nil
}
