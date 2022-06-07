package integration_test

import (
	"access-log/src/firehose"
	"access-log/test/integration/helper"
	"encoding/json"
	uuid2 "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestNoS3EntryIsWrittenForIgnoredPath(t *testing.T) {
	helper.WaitForHealthcheck(t)
	helper.WaitForLocalstack(t)
	helper.ResetLocalstack(t)

	_, err := http.Get("http://localhost:8080/ignored/3")
	if err != nil {
		t.Fatalf("Couldn't call endpoint: %s", err)
	}

	//Buffer time for Firehose is set to 60 seconds, buffer timeout in gateway set to 10s - so worst case should be
	//70 seconds + processing time on Firehose mock for objects to appear in S3.
	time.Sleep(75 * time.Second)

	keys := helper.ListObjects(t, "factory-access-log-bucket")

	require.Empty(t, keys)
}

func TestS3EntryWrittenForNonIgnoredPath(t *testing.T) {
	helper.WaitForHealthcheck(t)
	helper.WaitForLocalstack(t)
	helper.ResetLocalstack(t)

	request, _ := http.NewRequest("GET", "http://localhost:8080/example/3", nil)
	request.Header.Add("Authorization", "Bearer eyJhbGciOiAiSFMyNTYiLCJ0eXAiOiAiSldUIn0.eyJzdWIiOiAic3ViLWJsYWJsYSIsIm5hbWUiOiAiU29tZWJvZHkiLCJpYXQiOiAxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ICJhOWRlOTNmYy0yZDEzLTQ0ZGQtOTI3Mi1kYTdmOGMxN2QxNTUiLCJhbmRyb2lkX2lkIjogIjA3ZmYwMGU0LWMxYTUtNDY4My05ZmNiLTYxM2E3MzRkOGQzZiJ9.aW52YWxpZCBzaWduYXR1cmU")
	_, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("Couldn't call endpoint: %s", err)
	}

	//Buffer time for Firehose is set to 60 seconds, buffer timeout in gateway set to 10s - so worst case should be
	//70 seconds + processing time on Firehose mock for objects to appear in S3.
	time.Sleep(75 * time.Second)

	keys := helper.ListObjects(t, "factory-access-log-bucket")

	require.Len(t, keys, 1)

	content := helper.GetObjectContent(t, "factory-access-log-bucket", keys[0])

	var accessLog firehose.AccessLog
	err = json.Unmarshal(content, &accessLog)
	if err != nil {
		t.Fatalf("Could not unmarshall contents of object: %s", err)
	}

	assert.Equal(t, "Test Product", accessLog.Product)
	assert.Equal(t, "GET", accessLog.Method)
	assert.Equal(t, "/example/3", accessLog.Path)
	assert.Equal(t, "07ff00e4-c1a5-4683-9fcb-613a734d8d3f", accessLog.AndroidId)
	assert.Equal(t, uuid2.MustParse("a9de93fc-2d13-44dd-9272-da7f8c17d155"), accessLog.SubscriptionId)
}
