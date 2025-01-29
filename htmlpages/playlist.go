package htmlpages

const Playlist = `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Titled - Playlist</title>
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
				height: 100vh;
				padding: 10px;
				overflow-x: hidden; /* Prevent horizontal scrolling */
			}
			.container {
				width: 100%%;
				max-width: 600px;
				background-color: #ffffff;
				border: 8px solid #000000;
				box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
				padding: 20px;
				display: flex;
				flex-direction: column;
				gap: 15px;
				border-radius: 10px;
				box-sizing: border-box; /* Include padding and border in width/height */
			}
			.container div {
				border: 3px solid #000000;
				padding: 15px;
				border-radius: 5px;
				box-sizing: border-box;
			}
			.container .red {
				background-color: #ff0000;
				text-align: center;
				color: #ffffff;
			}
			.container .white {
				background-color: #ffffff;
				text-align: center;
			}
			.container .yellow {
				background-color: #FFFF00;
				border: 2px solid #000000;
				border-radius: 5px;
				padding: 15px;
				text-align: center;
				box-sizing: border-box;
			}
			.container .yellow-box a {
				color: #000000;
				text-decoration: none;
				font-weight: bold;
			}
			.container .yellow-box a:hover {
				text-decoration: underline;
			}
			a {
				color: #ffffff;
				text-decoration: none;
				font-weight: bold;
				background-color: #000000;
				padding: 10px 20px;
				border-radius: 5px;
				display: inline-block;
			}
			a:hover {
				background-color: #555555;
			}
			iframe {
				border-radius: 12px;
				width: 100%%;
				height: 352px;
				border: 0;
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
				a {
					padding: 8px 15px;
					font-size: 14px;
				}
				iframe {
					height: 280px;
				}
				.container .yellow {
					padding: 10px;
				}
			}
		</style>
	</head>
	<body>
		<div class="container">
			<div class="red">
				<p>Your Spotify playlist is ready!</p>
				<a href="%s">Click here</a> to open it.
			</div>
			<div class="white">
				<iframe src="https://open.spotify.com/embed/playlist/%s?utm_source=generator" frameborder="0" allowfullscreen allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture" loading="lazy"></iframe>
			</div>
			<div class="yellow">
				<a href="/">Go Back to Home</a>
			</div>
		</div>
	</body>
	</html>`
