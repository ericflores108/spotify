package httpserver

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/service"
)

type Server struct {
	Service *service.Service
	Ctx     context.Context
}

func NewServer(ctx context.Context, service *service.Service) *Server {
	return &Server{
		Service: service,
		Ctx:     ctx,
	}
}

func (s *Server) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files (e.g., favicon.ico) from the ./static directory
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	tmpl := template.Must(template.New("index").Parse(`
	<!doctype html>
	<html>
	<head>
		<title>Titled - Generate a Spotify playlist based off an album</title>
		<link rel="icon" href="/static/favicon.ico" type="image/x-icon">
		<link href="https://fonts.googleapis.com/css2?family=Raleway:wght@400;700&display=swap" rel="stylesheet">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<style>
			body {
				font-family: 'Raleway', Arial, sans-serif;
				margin: 0;
				padding: 0;
				background-color: #ffffff;
				color: #000000;
				display: flex;
				justify-content: center;
				align-items: center;
				min-height: 100vh;
				padding: 10px;
				overflow-x: hidden;
				box-sizing: border-box;
			}
			.container {
				width: 100%;
				max-width: 600px;
				background-color: #ffffff;
				border: 8px solid #000000;
				box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
				padding: 20px;
				display: grid;
				grid-template-columns: repeat(2, 1fr);
				gap: 15px;
				border-radius: 10px;
				box-sizing: border-box;
			}
			.container div {
				border: 3px solid #000000;
				padding: 15px;
				border-radius: 5px;
				box-sizing: border-box;
			}
			.container .red {
				background-color: #ff0000;
				grid-column: span 2;
				text-align: center;
				color: #ffffff;
			}
			.container .blue {
				background-color: #0000ff;
				grid-column: span 2;
				text-align: center;
				color: #ffffff;
			}
			.container .white {
				background-color: #ffffff;
				text-align: left;
				grid-column: span 2;
			}
			.container .white ul {
				text-align: left;
				padding-left: 20px;
				margin: 0;
				list-style: none;
			}
			.container .white ul li::before {
				content: "\266A "; /* Unicode for music note ♪ */
				color: #1D1D1F; /* Blue color for the music note */
				font-size: 1.2em;
				margin-right: 10px;
			}
			a {
				display: inline-block;
				margin: 10px auto;
				padding: 10px 20px;
				background-color: #000000;
				color: #ffffff;
				text-decoration: none;
				font-weight: bold;
				border-radius: 5px;
				border: 2px solid #ffffff;
			}
			a:hover {
				background-color: #555555;
			}
			/* Responsive Design */
			@media (max-width: 480px) {
				body {
					padding: 5px;
				}
				.container {
					border-width: 5px;
					padding: 15px;
				}
				.container div {
					border-width: 2px;
					padding: 10px;
				}
				.container .white ul {
					padding-left: 15px;
				}
				a {
					padding: 8px 15px;
					font-size: 14px;
				}
			}
		</style>
	</head>
	<body>
		<div class="container">
			<div class="red">
				<h1>Titled</h1>
			</div>
			<div class="blue">
				<a href="/login">Log in with Spotify</a>
			</div>
			<div class="white">
				<h2>Generate a custom Spotify playlist</h2>
				<h3>How It Works</h3>
				<p>For each song in the selected album, we analyze its inspirations and influences:</p>
				<ul>
					<li>Any songs it samples</li>
					<li>General inspirations or influences behind its creation</li>
					<li>Based on the analysis, you'll see a generated playlist in your Spotify app.</li>
					<li>Each playlist generated will have the prefix, "Titled - Inspired Songs from [album]"</li>
				</ul>
			</div>
		</div>
	</body>
	</html>`))

	// Serve the root page with the template
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/login", s.Service.LoginHandler)

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		s.Service.CallbackHandler(w, s.Ctx, r)
	})

	mux.HandleFunc("/generatePlaylist", func(w http.ResponseWriter, r *http.Request) {
		// Ensure the request method is POST
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Parse the form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Extract the userID and album name from the form
		userID := r.FormValue("userID")
		albumURL := r.FormValue("albumURL")

		if userID == "" {
			http.Error(w, "User ID cannot be empty", http.StatusBadRequest)
			return
		}

		// Set the logger prefix to the user ID
		logger.InfoLogger.SetPrefix(fmt.Sprintf("UserID: %s", userID))
		logger.DebugLogger.SetPrefix(fmt.Sprintf("UserID: %s", userID))
		logger.ErrorLogger.SetPrefix(fmt.Sprintf("UserID: %s", userID))

		if albumURL == "" {
			http.Error(w, "Album link cannot be empty", http.StatusBadRequest)
			return
		}
		// Process the form submission
		logger.LogInfo("User ID submitted: %s", userID)
		logger.LogInfo("Album link submitted: %s", albumURL)

		parts := strings.Split(albumURL, "/album/")
		if len(parts) < 2 {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}
		logger.LogDebug("Album link submitted: %s", albumURL)

		s.Service.GeneratePlaylistHandler(w, s.Ctx, parts[1], userID, r)
	})

	return mux
}
