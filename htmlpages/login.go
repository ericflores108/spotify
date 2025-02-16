package htmlpages

const Login = `
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

			/* Cookie Banner */
			.cookie-banner {
				position: fixed;
				bottom: 10px;
				left: 50%;
				transform: translateX(-50%);
				background: #000;
				color: #fff;
				padding: 10px 20px;
				border-radius: 5px;
				font-size: 14px;
				text-align: center;
				z-index: 1000;
				display: flex;
				align-items: center;
				gap: 10px;
			}
			.cookie-banner p {
				margin: 0;
				font-size: 14px;
			}
			.cookie-banner button {
				background: #fff;
				color: #000;
				border: none;
				padding: 5px 10px;
				cursor: pointer;
				font-weight: bold;
				border-radius: 5px;
			}
			.cookie-banner button:hover {
				background: #ccc;
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
				.cookie-banner {
					flex-direction: column;
					text-align: center;
					padding: 15px;
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
					<li>
							Any songs it samples. ie,
							<ul>
									<li><i>Song:</i> Squabble Up by Kendrick Lamar</li>
									<li><i>Sampled Song:</i> When I Hear Music by Debbie Deb</li>
							</ul>
					</li>
					<li>If no sampled song, we analyze general inspirations or influences behind its creation</li>
					<li>Based on the analysis, you'll see a generated playlist in your Spotify app.</li>
					<li>Each playlist generated will have the prefix, "Titled - Inspired Songs from [album]"</li>
				</ul>
			</div>
		</div>

		<!-- Cookie Consent Banner -->
		<div id="cookie-banner" class="cookie-banner">
			<p>This site uses necessary cookies for authentication and session management.</p>
			<button id="accept-cookies">Accept</button>
			<button id="reject-cookies">Reject</button>
		</div>

		<script>
			// Initialize cookie to "cookies_accepted=null" if no cookie preference exists
			if (!document.cookie.includes("cookies_accepted") && !document.cookie.includes("cookies_rejected")) {
				document.cookie = "cookies_accepted=null; path=/; max-age=" + (60*60*24*365) + "; Secure; SameSite=Lax";
			}

			document.getElementById("accept-cookies").addEventListener("click", function() {
				document.cookie = "cookies_accepted=true; path=/; max-age=" + (60*60*24*365) + "; Secure; SameSite=Lax";
				document.getElementById("cookie-banner").style.display = "none";
			});

			document.getElementById("reject-cookies").addEventListener("click", function() {
				document.cookie = "cookies_rejected=true; path=/; max-age=" + (60*60*24*365) + "; Secure; SameSite=Lax";
				document.getElementById("cookie-banner").style.display = "none";
			});

			if (document.cookie.includes("cookies_accepted=true") || document.cookie.includes("cookies_rejected=true")) {
				document.getElementById("cookie-banner").style.display = "none";
			}
		</script>
	</body>
	</html>`
