package htmlpages

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

// Template variables to store HTML content
var (
	Login            string
	Playlist         string
	GeneratePlaylist string
	errorTemplate    string
)

func init() {
	// Load all templates during initialization
	loadTemplates()
}

// loadTemplates reads all HTML template files and stores their content
func loadTemplates() {
	// Define templates to load with their corresponding variable
	templates := map[string]*string{
		"login.html":    &Login,
		"playlist.html": &Playlist,
		"forms.html":    &GeneratePlaylist,
		"error.html":    &errorTemplate,
	}

	// Load each template
	for filename, variable := range templates {
		templatePath := filepath.Join("htmlpages", "html", filename)
		content, err := os.ReadFile(templatePath)
		if err != nil {
			panic("Failed to read " + filename + " template: " + err.Error())
		}
		*variable = string(content)
	}
}

// RenderErrorPage renders the error template with the provided error message
func RenderErrorPage(w http.ResponseWriter, errorMessage string) {
	tmpl, err := template.New("error").Parse(errorTemplate)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		ErrorMessage string
	}{
		ErrorMessage: errorMessage,
	}

	w.WriteHeader(http.StatusInternalServerError)
	tmpl.Execute(w, data)
}
