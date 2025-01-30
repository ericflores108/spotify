package htmlpages

import (
	"html/template"
	"net/http"
)

const errorTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Error - Titled</title>
    <link rel="icon" href="/static/favicon.ico" type="image/x-icon">
    <link href="https://fonts.googleapis.com/css2?family=Raleway:wght@400;700&display=swap" rel="stylesheet">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: 'Raleway', Arial, sans-serif;
            margin: 0;
            padding: 0;
            background-color: #ffffff;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            grid-template-rows: repeat(3, 1fr);
            gap: 10px;
            width: 80vw;
            height: 80vh;
            background-color: black;
            padding: 10px;
        }
        .block {
            display: flex;
            justify-content: center;
            align-items: center;
            font-size: 24px;
            font-weight: bold;
            color: black;
            border: 8px solid black;
        }
        .red { background-color: #ff0000; grid-column: span 2; }
        .blue { background-color: #0000ff; grid-column: span 1; }
        .yellow { background-color: #ffff00; grid-column: span 1; }
        .white { background-color: #ffffff; grid-column: span 2; }
        .message {
            background-color: white;
            text-align: center;
            font-size: 20px;
            grid-column: span 4;
            padding: 20px;
        }
        a {
            text-decoration: none;
            font-weight: bold;
            color: black;
        }
        a:hover {
            color: #555;
        }
    </style>
</head>
<body>
    <div class="grid">
        <div class="block red">Error</div>
        <div class="block blue"></div>
        <div class="block yellow"></div>
        <div class="block white"></div>
        <div class="block white"></div>
        <div class="block blue"></div>
        <div class="block red"></div>
        <div class="block yellow"></div>
        <div class="message">{{.ErrorMessage}} <br>
            <a href="/">Return to Home</a>
        </div>
    </div>
</body>
</html>
`

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
