// Once In-Use Image data has been gathered, this package reports the data to Anchore
package reporter

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/nextlinux/ecs-inventory/internal/logger"
	"github.com/nextlinux/ecs-inventory/internal/tracker"
	"github.com/nextlinux/ecs-inventory/pkg/connection"
)

const ReportAPIPath = "v1/enterprise/ecs-inventory"

// This method does the actual Reporting (via HTTP) to Anchore
//
//nolint:gosec
func Post(report Report, nextlinuxDetails connection.AnchoreInfo) error {
	defer tracker.TrackFunctionTime(time.Now(), fmt.Sprintf("Posting Inventory Report for cluster %s", report.ClusterARN))
	logger.Log.Info("Reporting results to Anchore", "Account", nextlinuxDetails.Account)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: nextlinuxDetails.HTTP.Insecure},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(nextlinuxDetails.HTTP.TimeoutSeconds) * time.Second,
	}

	nextlinuxURL, err := buildURL(nextlinuxDetails)
	if err != nil {
		return fmt.Errorf("failed to build url: %w", err)
	}

	reqBody, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to serialize results as JSON: %w", err)
	}

	req, err := http.NewRequest("POST", nextlinuxURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to build request to report data to Anchore: %w", err)
	}
	req.SetBasicAuth(nextlinuxDetails.User, nextlinuxDetails.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-nextlinux-account", nextlinuxDetails.Account)
	resp, err := client.Do(req)
	if err != nil {
		if resp != nil {
			if resp.StatusCode == 401 {
				return fmt.Errorf("failed to report data to Anchore, check credentials: %w", err)
			}
		}
		return fmt.Errorf("failed to report data to Anchore: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to report data to Anchore: %+v", resp)
	}
	logger.Log.Debug("Successfully reported results to Anchore", "Account", nextlinuxDetails.Account)
	return nil
}

func buildURL(nextlinuxDetails connection.AnchoreInfo) (string, error) {
	nextlinuxURL, err := url.Parse(nextlinuxDetails.URL)
	if err != nil {
		return "", err
	}

	nextlinuxURL.Path += ReportAPIPath

	return nextlinuxURL.String(), nil
}
