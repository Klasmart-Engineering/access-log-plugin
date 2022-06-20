package config_test

import (
	"access-log/src/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPatternsAndPathsProduceExpectedOutput(t *testing.T) {
	tests := map[string]struct {
		ignoredPath string
		requestPath string
		want        bool
	}{
		"literal root match":            {ignoredPath: "/", requestPath: "/", want: true},
		"literal root path match":       {ignoredPath: "/path", requestPath: "/path", want: true},
		"literal root path no match":    {ignoredPath: "/path", requestPath: "/paath", want: false},
		"wildcard root path match":      {ignoredPath: "/*", requestPath: "/path", want: true},
		"wildcard nested path match":    {ignoredPath: "/*/*", requestPath: "/path/path2", want: true},
		"wildcard nested path no match": {ignoredPath: "/path/*", requestPath: "/path2/path3", want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := config.NewIgnoredPaths([]string{tc.ignoredPath}).AnyMatch(tc.requestPath)

			require.Equal(t, tc.want, got)
		})
	}
}
