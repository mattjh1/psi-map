package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattjh1/psi-map/internal/types"
)

func TestHandleAPIResults(t *testing.T) {
	results := []*types.PageResult{
		createMockResult("https://example1.com", 90, 85, 80, 95, false),
		createMockResult("https://example2.com", 70, 75, 65, 80, false),
	}

	server := &Server{results: results}

	req := httptest.NewRequest(http.MethodGet, "/api/results", http.NoBody)
	w := httptest.NewRecorder()

	server.handleAPIResults(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var responseResults []types.PageResult
	err := json.Unmarshal(w.Body.Bytes(), &responseResults)
	require.NoError(t, err)

	assert.Len(t, responseResults, 2)
	assert.Equal(t, "https://example1.com", responseResults[0].URL)
	assert.Equal(t, "https://example2.com", responseResults[1].URL)
}

func TestHandleAPIResult_ValidIndex(t *testing.T) {
	results := []*types.PageResult{
		createMockResult("https://example1.com", 90, 85, 80, 95, false),
		createMockResult("https://example2.com", 70, 75, 65, 80, false),
	}

	server := &Server{results: results}

	req := httptest.NewRequest(http.MethodGet, "/api/results/1", http.NoBody)
	w := httptest.NewRecorder()

	server.handleAPIResult(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result types.PageResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "https://example2.com", result.URL)
	assert.Equal(t, 70.0, result.Mobile.Scores.Performance)
}

func TestHandleAPIResult_InvalidIndex(t *testing.T) {
	results := []*types.PageResult{
		createMockResult("https://example1.com", 90, 85, 80, 95, false),
	}

	server := &Server{results: results}

	testCases := []struct {
		name string
		path string
	}{
		{"negative index", "/api/results/-1"},
		{"out of bounds", "/api/results/5"},
		{"non-numeric", "/api/results/abc"},
		{"empty", "/api/results/"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, http.NoBody)
			w := httptest.NewRecorder()

			server.handleAPIResult(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "Invalid result index")
		})
	}
}
