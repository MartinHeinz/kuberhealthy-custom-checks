package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/itchyny/gojq"
	"github.com/kuberhealthy/kuberhealthy/v2/pkg/checks/external/nodeCheck"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kuberhealthy/kuberhealthy/v2/pkg/checks/external"
	"github.com/kuberhealthy/kuberhealthy/v2/pkg/checks/external/checkclient"
	log "github.com/sirupsen/logrus"
)

var (
	TargetURL       = os.Getenv("TARGET_URL")
	ExpectedResult  = os.Getenv("EXPECTED_RESULT")
	JqQuery         = os.Getenv("JQ_QUERY")
	TimeoutDuration = os.Getenv("TIMEOUT_DURATION")
	RequestData     = os.Getenv("REQUEST_DATA")
	RequestMethod   = os.Getenv("REQUEST_METHOD")
	RequestHeaders  = make(map[string]string)
)

func init() {
	nodeCheck.EnableDebugOutput()

	if TargetURL == "" {
		reportErrorAndStop("No URL provided in YAML")
	}
	if ExpectedResult == "" {
		reportErrorAndStop("No expected result provided in YAML")
	}
	if JqQuery == "" {
		reportErrorAndStop("No jq query provided in YAML")
	}

	prefix := "KH_REQUEST_HEADER_"
	pattern := `\${(.*)}`
	rgx := regexp.MustCompile(pattern)
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			if strings.HasPrefix(e[:i], prefix) {
				value := e[i+1:]
				match, _ := regexp.MatchString(pattern, value)
				if match {
					rs := rgx.FindStringSubmatch(value)
					envVar := os.Getenv(rs[1])
					value = rgx.ReplaceAllString(value, envVar)
				}
				RequestHeaders[e[len(prefix):i]] = value
			}
		}
	}
}

// Inspired by https://github.com/kuberhealthy/kuberhealthy/blob/master/cmd/http-content-check/main.go
func main() {

	log.Println("Using kuberhealthy reporting url", os.Getenv(external.KHReportingURL))

	checkTimeLimit := time.Minute * 1
	ctx, cancelFunc := context.WithTimeout(context.Background(), checkTimeLimit)
	defer cancelFunc()

	var err error
	err = nodeCheck.WaitForKuberhealthy(ctx)
	if err != nil {
		log.Errorln("Error waiting for kuberhealthy endpoint to be contactable by checker pod with error:" + err.Error())
	}

	ok, err := doCheck()
	if !ok {
		if err != nil {
			reportErrorAndStop(err.Error())
		}

	} else {
		log.Println("Reporting success...")
		err = checkclient.ReportSuccess()
		if err != nil {
			log.Errorln("failed to report success", err)
			os.Exit(1)
		}
		log.Infoln("Successfully reported to Kuberhealthy")
	}
}

func doCheck() (bool, error) {

	query, err := gojq.Parse(JqQuery)
	if err != nil {
		return false, err
	}
	data, err := getURLContent(TargetURL)
	log.Println("Attempting to fetch content from: " + TargetURL)
	if err != nil {
		return false, err
	}
	log.Println("Attempting run query against content")
	iter := query.Run(data)
	for {
		v, ok := iter.Next()
		if !ok {
			log.Println("No match found")
			return false, errors.New("no match found")
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
			return false, err
		}
		if v == ExpectedResult {
			log.Println("Found match")
			return true, nil
		}
	}
}

func getURLContent(url string) (map[string]any, error) {
	dur, err := time.ParseDuration(TimeoutDuration)
	if err != nil {
		return nil, err
	}
	client := http.Client{Timeout: dur}
	req, _ := http.NewRequest(RequestMethod, url, bytes.NewBuffer([]byte(RequestData)))

	for _, header := range RequestHeaders {
		headerParts := strings.Split(header, ": ")
		req.Header.Set(headerParts[0], headerParts[1])
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m := make(map[string]any)

	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// reportErrorAndStop reports to kuberhealthy of error and exits program when called
func reportErrorAndStop(s string) {
	log.Infoln("attempting to report error to kuberhealthy:", s)
	err := checkclient.ReportFailure([]string{s})
	if err != nil {
		log.Errorln("failed to report to kuberhealthy servers:", err)
		os.Exit(1)
	}
	log.Infoln("Successfully reported to Kuberhealthy")
	os.Exit(0)
}
