package htmlpages

const GeneratePlaylist = `
<!DOCTYPE html>
<html>
<head>
    <title>Titled - User Form</title>
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
        }
        .container {
            width: 100%;
            max-width: 600px;
            background-color: #ffffff;
            border: 8px solid #000000;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            padding: 20px;
            display: grid;
            grid-template-columns: 1fr;
            gap: 15px;
            border-radius: 10px;
        }
        .container div {
            border: 3px solid #000000;
            padding: 15px;
            border-radius: 5px;
        }
        .container .red {
            background-color: #ff0000;
            text-align: center;
            color: #ffffff;
        }
        .container .yellow {
            background-color: #ffff00;
            text-align: center;
        }
        .container .blue {
            background-color: #0000ff;
            text-align: center;
            color: #ffffff;
            display: none;
            padding: 15px;
            border-radius: 5px;
        }
        form {
            display: flex;
            flex-direction: column;
            gap: 15px;
        }
        label {
            font-weight: bold;
            text-align: left;
        }
        input, button {
            width: 100%;
            padding: 12px;
            font-size: 16px;
            border: 2px solid #000000;
            border-radius: 5px;
            box-sizing: border-box;
        }
        button {
            background-color: #000000;
            color: #ffffff;
            cursor: pointer;
            font-weight: bold;
        }
        button:hover {
            background-color: #555555;
        }
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
            input, button {
                padding: 10px;
                font-size: 14px;
            }
        }
    </style>
    <script>
        function validateInput(event) {
            event.preventDefault(); // Prevent form submission

            const albumInput = document.getElementById('albumURL');
            const albumURL = albumInput.value.trim();
            const validPrefix = "https://open.spotify.com/album/";

            if (!albumURL.startsWith(validPrefix)) {
                alert("Please enter a valid Spotify album link.");
                return false;
            }

            showLoading(event); // Call showLoading if validation passes
            return true;
        }

        function showLoading(event) {
            // Show the blue box and loading message
            document.querySelector('.blue').style.display = 'block';

            // Allow form submission after the loading message is displayed
            setTimeout(() => {
                event.target.submit();
            }, 50);
        }
    </script>
</head>
<body>
    <div class="container">
        <div class="red">
            <h1>Generate Spotify Playlist</h1>
        </div>
        <div class="yellow">
            <form action="/generatePlaylist" method="post" onsubmit="return validateInput(event)">
                <input type="hidden" id="userID" name="userID" value="{{.UserID}}">
                <input type="hidden" id="accessToken" name="accessToken" value="{{.AccessToken}}">
                
                <label for="albumURL">Insert Spotify Album Link:</label>
                <input type="text" id="albumURL" name="albumURL" value="{{.AlbumURL}}" required>

                <button type="submit">Generate</button>
            </form>
        </div>
        <div class="blue">
            <p>Please wait... Generating your Spotify playlist.</p>
        </div>
    </div>
</body>
</html>
`
