package integration_test

import (
	"access-log/test/integration/helper"
	"net/http"
	"testing"
)

func TestNoS3EntriesAreWrittenForIgnoredPaths(t *testing.T) {
	helper.WaitForHealthcheck(t)
	helper.WaitForLocalstack(t)
	helper.ResetLocalstack(t)

	_, err := http.Get("http://localhost:8080/ignored/3")
	if err != nil {
		t.Fatalf("Couldn't call endpoint: %s", err)
	}

	//TODO
}
