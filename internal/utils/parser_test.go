package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSitemap_LocalFile(t *testing.T) {
	xmlContent := `
		<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
			<url>
				<loc>https://example.com/page1</loc>
			</url>
			<url>
				<loc>https://example.com/page2</loc>
			</url>
		</urlset>
	`
	tmpFile, err := os.CreateTemp("", "sitemap-*.xml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(xmlContent)
	assert.NoError(t, err)
	tmpFile.Close()

	urls, err := ParseSitemap(tmpFile.Name())
	assert.NoError(t, err)
	assert.Len(t, urls, 2)
	assert.Equal(t, "https://example.com/page1", urls[0])
	assert.Equal(t, "https://example.com/page2", urls[1])
}

func TestParseSitemap_RemoteFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, err := w.Write([]byte(`
            <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
                <url>
                    <loc>http://example.com/remote1</loc>
                </url>
            </urlset>
        `))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	urls, err := ParseSitemap(server.URL)
	assert.NoError(t, err)
	assert.Len(t, urls, 1)
	assert.Equal(t, "http://example.com/remote1", urls[0])
}
