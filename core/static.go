package core

import (
	"net/http"
	"path"
	"strings"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

// Static registers a route to serve static files from a directory.
//
// The prefix is the URL path prefix, and root is the filesystem directory.
// Files are served using http.FileServer with directory listing disabled.
//
// Usage:
//
//	app.Static("/static", "./public")  // serves ./public/style.css at /static/style.css
func (e *Engine) Static(prefix, root string) {
	e.StaticFS(prefix, http.Dir(root))
}

// StaticFile registers a route to serve a single file.
//
// Usage:
//
//	app.StaticFile("/favicon.ico", "./assets/favicon.ico")
func (e *Engine) StaticFile(urlPath, filepath string) {
	e.GET(urlPath, func(c *rudraContext.Context) error {
		http.ServeFile(c.Writer(), c.Request(), filepath)
		return nil
	})
	e.HEAD(urlPath, func(c *rudraContext.Context) error {
		http.ServeFile(c.Writer(), c.Request(), filepath)
		return nil
	})
}

// StaticFS registers a route to serve files from an http.FileSystem.
//
// This supports custom file systems (embed.FS, afero, etc.) via http.FS().
//
// Usage:
//
//	app.StaticFS("/assets", http.FS(embedFS))
//	app.StaticFS("/static", http.Dir("./public"))
func (e *Engine) StaticFS(prefix string, fs http.FileSystem) {
	// Ensure prefix ends without trailing slash for consistency.
	prefix = strings.TrimRight(prefix, "/")

	fileServer := http.StripPrefix(prefix, http.FileServer(&noListingFS{fs}))

	handler := func(c *rudraContext.Context) error {
		// Prevent directory traversal.
		file := c.Param("filepath")
		if strings.Contains(file, "..") {
			return c.String(http.StatusForbidden, "403 forbidden")
		}
		fileServer.ServeHTTP(c.Writer(), c.Request())
		return nil
	}

	// Register both the prefix and the wildcard route.
	pattern := prefix + "/*filepath"
	e.GET(pattern, handler)
	e.HEAD(pattern, handler)
}

// noListingFS wraps http.FileSystem to disable directory listing.
// When a directory is requested, it returns 404 instead of listing contents.
type noListingFS struct {
	fs http.FileSystem
}

func (nfs *noListingFS) Open(name string) (http.File, error) {
	f, err := nfs.fs.Open(name)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	// If it's a directory, check for index.html.
	if info.IsDir() {
		indexPath := path.Join(name, "index.html")
		if _, err := nfs.fs.Open(indexPath); err != nil {
			f.Close()
			return nil, err
		}
	}

	return f, nil
}
