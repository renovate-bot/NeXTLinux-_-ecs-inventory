package reporter

import (
	"testing"

	"github.com/nextlinux/ecs-inventory/pkg/connection"
)

func TestBuildUrl(t *testing.T) {
	nextlinuxDetails := connection.AnchoreInfo{
		URL:      "https://ancho.re",
		User:     "admin",
		Password: "foobar",
	}

	expectedURL := "https://ancho.re/v1/enterprise/ecs-inventory"
	actualURL, err := buildURL(nextlinuxDetails)
	if err != nil || expectedURL != actualURL {
		t.Errorf("Failed to build URL:\nexpected=%s\nactual=%s", expectedURL, actualURL)
	}
}
