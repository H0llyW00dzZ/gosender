# GoSender (WIP TODO)

This is a Go-based server side utility for sending emails using Google's Cloud Gmail API.

## Configuration

1. Obtain Gmail API credentials: 
   - Go to the [Google Cloud Console](https://console.cloud.google.com/).
   - Create a new project or select an existing one.
   - Enable the Gmail API for your project.
   - Create credentials (OAuth 2.0 client ID) for a web application.
   - Download the JSON file containing your credentials.
   - Rename the downloaded file to `credentials.json` and place it in the project directory.

## Usage

1. Build the application:

   ```shell
   go build
   ```

2. Run the application:

   ```shell
   ./gosender
   ```

3. Send a request to the application's HTTP endpoint:
   - Method: POST
   - URL: http://localhost:8080/send
   - Parameters:
     - `payload`: Base64-encoded JSON payload containing the necessary information.

     Example payload:
     ```json
     {
       "credentials": "<base64-encoded credentials JSON>",
       "token": "<base64-encoded token JSON>",
       "messageBody": "<base64-encoded message body>"
     }
     ```

4. The application will send the email message using the Gmail API and perform additional actions on existing messages in the user's Gmail account.

## License

This project is licensed under the [MIT License](LICENSE).
