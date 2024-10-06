# go-audio2org

`go-audio2org` is a Go utility that transcribes audio files using OpenAI's Whisper API and generates Emacs org notes from the transcription. This can be particularly useful for creating organized notes from recorded meetings, lectures, or any other audio content.

## Features

- Transcribe audio files to text using the Whisper API.
- Generate Emacs org-mode formatted notes from the transcriptions.
- Allows for custom post-processing commands.

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/YOUR_USERNAME/go-audio2org.git
   cd go-audio2org
   ```

2. Install the required Go packages:

   ```sh
   go mod tidy
   ```

3. Create a `.env` file in the root directory with your OpenAI API key:

   ```env
   OPENAI_API_KEY=your_openai_api_key
   ```

## Usage

1. Transcribe an audio file and optionally generate Emacs org notes:

   ```sh
   go run main.go -file path/to/your/audiofile.mp3 -post create_emacs_org_notes
   ```

2. Use an existing transcription file to generate Emacs org notes:

   ```sh
   go run main.go -transcription path/to/your/transcription.txt -post create_emacs_org_notes
   ```

### Command-line Flags

- `-file`: Path to the audio file to transcribe (optional if `-transcription` is provided).
- `-transcription`: Path to the existing transcription file (optional).
- `-output`: Name of the output transcription file (optional, will include a timestamp if not provided).
- `-post`: Post-processing command to run ("create_emacs_org_notes" is available).

### Example Commands

- Transcribe an audio file and save the transcription with a custom name:

  ```sh
  go run main.go -file path/to/audio.mp3 -output transcription.txt
  ```

- Transcribe an audio file and save the transcription with a default name (timestamp will be included):

  ```sh
  go run main.go -file path/to/audio.mp3
  ```

- Generate Emacs org notes from an existing transcription:

  ```sh
  go run main.go -transcription path/to/transcription.txt -post create_emacs_org_notes
  ```

## Development

### Prerequisites

- [Go](https://golang.org/) installed on your machine (version 1.21 or later recommended).
- An OpenAI API key. Sign up at [OpenAI](https://openai.com/) if you don't have one.

### Building the Project

Compile the project with the following command:

```sh
go build
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgements

- [OpenAI](https://openai.com/) for their fantastic AI models and APIs.
- The Go community for their contributions to the language and ecosystem.

## Contact

For any questions or suggestions, please reach out or create an issue in this repository.
