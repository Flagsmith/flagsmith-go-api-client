package flagsmithapi_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flagsmithapi "github.com/Flagsmith/flagsmith-go-api-client"
)

// projectLookupCounter counts requests to each project lookup endpoint.
type projectLookupCounter struct {
	mu     sync.Mutex
	byID   int
	byUUID int
}

func (c *projectLookupCounter) counts() (int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.byID, c.byUUID
}

// projectCacheTestServer serves both project lookups plus the feature and segment
// get-by-uuid endpoints, counting how often each project lookup is hit.
func projectCacheTestServer(t *testing.T) (*httptest.Server, *projectLookupCounter) {
	t.Helper()

	mux := http.NewServeMux()
	counter := &projectLookupCounter{}

	writeProject := func(rw http.ResponseWriter) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	}

	// Registered as a subtree pattern, so it also receives requests to paths beneath
	// it. Only exact matches are counted, so the count stays a project lookup count
	// whatever else a caller requests under /projects/{id}/.
	projectByIDPath := fmt.Sprintf("/api/v1/projects/%d/", ProjectID)
	mux.HandleFunc(projectByIDPath, func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == projectByIDPath {
			counter.mu.Lock()
			counter.byID++
			counter.mu.Unlock()
		}
		writeProject(rw)
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/get-by-uuid/%s/", ProjectUUID), func(rw http.ResponseWriter, req *http.Request) {
		counter.mu.Lock()
		counter.byUUID++
		counter.mu.Unlock()
		writeProject(rw)
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/features/get-by-uuid/%s/", FeatureUUID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, CreateFeatureResponseJson)
		assert.NoError(t, err)
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/segments/get-by-uuid/%s/", SegmentUUID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetSegmentResponseJson)
		assert.NoError(t, err)
	})

	return httptest.NewServer(mux), counter
}

func TestGetFeatureCachesProjectLookup(t *testing.T) {
	// Given
	server, counter := projectCacheTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	for i := 0; i < 3; i++ {
		feature, err := client.GetFeature(FeatureUUID)
		require.NoError(t, err)
		assert.Equal(t, ProjectUUID, feature.ProjectUUID)
	}

	// Then
	byID, _ := counter.counts()
	assert.Equal(t, 1, byID)
}

// GetFeature and GetSegment resolve the same project, so the second one is free.
func TestGetSegmentSharesTheProjectCacheWithGetFeature(t *testing.T) {
	// Given
	server, counter := projectCacheTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	_, err := client.GetFeature(FeatureUUID)
	require.NoError(t, err)
	segment, err := client.GetSegment(SegmentUUID)
	require.NoError(t, err)

	// Then
	assert.Equal(t, ProjectUUID, segment.ProjectUUID)
	byID, _ := counter.counts()
	assert.Equal(t, 1, byID)
}

func TestGetProjectIDByUUIDCaches(t *testing.T) {
	// Given
	server, counter := projectCacheTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	for i := 0; i < 3; i++ {
		projectID, err := client.GetProjectIDByUUID(ProjectUUID)
		require.NoError(t, err)
		assert.Equal(t, ProjectID, projectID)
	}

	// Then
	_, byUUID := counter.counts()
	assert.Equal(t, 1, byUUID)
}

// Resolving a UUID also populates the by-ID cache, since one response carries both.
func TestGetProjectIDByUUIDPopulatesTheProjectRefCache(t *testing.T) {
	// Given
	server, counter := projectCacheTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	_, err := client.GetProjectIDByUUID(ProjectUUID)
	require.NoError(t, err)
	_, err = client.GetFeature(FeatureUUID)
	require.NoError(t, err)

	// Then
	byID, byUUID := counter.counts()
	assert.Equal(t, 1, byUUID)
	assert.Equal(t, 0, byID)
}

// The public project reads must stay live, so that a caller reading a project after
// updating it does not see stale values.
func TestGetProjectIsNotCached(t *testing.T) {
	// Given
	server, counter := projectCacheTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	for i := 0; i < 2; i++ {
		_, err := client.GetProject(ProjectUUID)
		require.NoError(t, err)
		_, err = client.GetProjectByID(ProjectID)
		require.NoError(t, err)
	}

	// Then
	byID, byUUID := counter.counts()
	assert.Equal(t, 2, byID)
	assert.Equal(t, 2, byUUID)
}

func TestProjectCacheIsConcurrencySafe(t *testing.T) {
	// Given
	server, counter := projectCacheTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			feature, err := client.GetFeature(FeatureUUID)
			assert.NoError(t, err)
			if err == nil {
				assert.Equal(t, ProjectUUID, feature.ProjectUUID)
			}
		}()
	}
	wg.Wait()

	// Then
	byID, _ := counter.counts()
	assert.Equal(t, 1, byID)
}
