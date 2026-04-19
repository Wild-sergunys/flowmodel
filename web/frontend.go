package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed templates/*.html templates/pages/*.html templates/partials/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

type PageData struct {
	Title string
}

type Handler struct {
	pages  map[string]*template.Template
	static http.Handler
}

func NewHandler() (*Handler, error) {
	files, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, err
	}

	pages, err := parsePages()
	if err != nil {
		return nil, err
	}

	return &Handler{
		pages:  pages,
		static: http.FileServer(http.FS(files)),
	}, nil
}

func parsePages() (map[string]*template.Template, error) {
	entries, err := fs.ReadDir(templateFS, "templates/pages")
	if err != nil {
		return nil, err
	}

	pages := make(map[string]*template.Template, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		pageName := strings.TrimSuffix(entry.Name(), ".html")
		tmpl, err := template.ParseFS(
			templateFS,
			"templates/layout.html",
			"templates/partials/*.html",
			"templates/pages/"+entry.Name(),
		)
		if err != nil {
			return nil, err
		}

		pages[pageName] = tmpl
	}

	return pages, nil
}

func (h *Handler) Static() http.Handler {
	return http.StripPrefix("/static/", h.static)
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	h.render(w, "home", PageData{
		Title: "FlowModel",
	})
}

func (h *Handler) render(w http.ResponseWriter, page string, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, ok := h.pages[page]
	if !ok {
		http.Error(w, "page template not found", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "template render error", http.StatusInternalServerError)
	}
}
