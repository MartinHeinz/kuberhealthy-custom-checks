package main

import (
	"encoding/json"
	"github.com/itchyny/gojq"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kuberhealthy/kuberhealthy/v2/pkg/checks/external"
	checkclient "github.com/kuberhealthy/kuberhealthy/v2/pkg/checks/external/checkclient"
)

// Inspired by https://github.com/kuberhealthy/kuberhealthy/blob/master/cmd/http-content-check/main.go
func main() {

	log.Println("Using kuberhealthy reporting url", os.Getenv(external.KHReportingURL))

	var err error
	if doCheck() {
		log.Println("Reporting success...")
		err = checkclient.ReportSuccess()
		if err != nil {
			log.Println(err.Error())
		}
	} else {
		log.Println("Reporting failure...")
		err = checkclient.ReportFailure([]string{"Test has failed!"})
		if err != nil {
			log.Println(err.Error())
		}
	}

	if err != nil {
		log.Println("Error reporting to Kuberhealthy servers:", err)
		return
	}
	log.Println("Successfully reported to Kuberhealthy servers")
}

func doCheck() bool {
	jqQuery := os.Getenv("JQ_QUERY")
	expectedResult := os.Getenv("EXPECTED_RESULT")
	targetURL := os.Getenv("TARGET_URL")

	query, err := gojq.Parse(jqQuery)
	if err != nil {
		log.Fatalln(err)
		return false
	}
	data, err := getURLContent(targetURL)
	log.Println("Attempting to fetch content from: " + targetURL)
	if err != nil {
		log.Fatalln(err)
		return false
	}
	iter := query.Run(data)
	for {
		v, ok := iter.Next()
		if !ok {
			log.Println("No match found")
			return false
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
			return false
		}
		if v == expectedResult {
			log.Println("Found match")
			return true
		}
	}
}

func getURLContent(url string) (map[string]any, error) {
	timeoutDuration := os.Getenv("TIMEOUT_DURATION")

	dur, err := time.ParseDuration(timeoutDuration)
	if err != nil {
		return nil, err
	}
	client := http.Client{Timeout: dur}
	resp, err := client.Get(url)
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
