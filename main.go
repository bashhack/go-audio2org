package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"flag"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

type Config struct {
	AudioFilePath         string
	TranscriptionFilePath string
	OutputFileName        string
	PostProcessCmd        string
	OpenAIAPIKey          string
}

type TranscriptionResponse struct {
	Text string `json:"text"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func main() {
	config := parseFlags()

	loadEnv()

	config.OpenAIAPIKey = getEnv("OPENAI_API_KEY")

	if config.AudioFilePath == "" && config.TranscriptionFilePath == "" {
		log.Fatal("The -file or -transcription argument is required.")
	}

	transcriptionText, outputFilePath := processTranscription(config)

	if config.PostProcessCmd == "create_emacs_org_notes" {
		createEmacsOrgNotes(transcriptionText, config.OpenAIAPIKey, outputFilePath)
	}
}

func parseFlags() Config {
	config := Config{}

	flag.StringVar(&config.AudioFilePath, "file", "", "Path to the audio file to transcribe (required)")
	flag.StringVar(&config.TranscriptionFilePath, "transcription", "", "Path to the existing transcription file (optional)")
	flag.StringVar(&config.OutputFileName, "output", "", "Name of the output transcription file (optional)")
	flag.StringVar(&config.PostProcessCmd, "post", "", "Post-processing command to run after transcription (optional)")

	flag.Parse()
	return config
}

func loadEnv() {
	log.Println("Loading environment variables...")
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v\n", err)
	}
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s not set in environment", key)
	}
	return value
}

func processTranscription(config Config) (string, string) {
	var transcriptionText string
	var outputFilePath string

	if config.AudioFilePath != "" {
		log.Printf("Reading audio file: %s\n", config.AudioFilePath)

		audioBytes, err := os.ReadFile(config.AudioFilePath)
		if err != nil {
			log.Fatalf("Error reading audio file: %v", err)
		}
		log.Println("Transcribing audio file...")
		transcriptionText = transcribeAudio(config.AudioFilePath, audioBytes, config.OpenAIAPIKey)

		outputDir := createOutputDir()
		outputFileName := config.OutputFileName
		if outputFileName == "" {
			outputFileName = "transcription.txt"
		}
		outputFilePath = generateTimestampedFilePath(outputDir, outputFileName)
		writeToFile(outputFilePath, transcriptionText)
	} else if config.TranscriptionFilePath != "" {
		transcriptionText = readExistingTranscription(config.TranscriptionFilePath)
		outputFilePath = config.TranscriptionFilePath
	}

	return transcriptionText, outputFilePath
}

func transcribeAudio(filePath string, audioBytes []byte, apiKey string) string {
	client := resty.New()
	client.SetTimeout(10 * time.Minute)

	log.Println("Sending request to Whisper API...")
	request := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey)).
		SetFileReader("file", filepath.Base(filePath), bytes.NewReader(audioBytes)).
		SetFormData(map[string]string{
			"model": "whisper-1",
		})

	resp, err := request.Post("https://api.openai.com/v1/audio/transcriptions")
	if err != nil {
		log.Fatalf("Error sending request to Whisper API: %v", err)
	}

	if resp.IsError() {
		log.Fatalf("Error response from Whisper API: %v", resp.String())
	}

	var transcriptionResp TranscriptionResponse
	if err := json.Unmarshal(resp.Body(), &transcriptionResp); err != nil {
		log.Fatalf("Error unmarshalling JSON response: %v", err)
	}

	return transcriptionResp.Text
}

func createOutputDir() string {
	outputDir := "output"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.Mkdir(outputDir, 0755); err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}
	}
	return outputDir
}

func generateTimestampedFilePath(outputDir, baseFileName string) string {
	timestamp := time.Now().Format("20060102_150405")
	ext := filepath.Ext(baseFileName)
	name := baseFileName[:len(baseFileName)-len(ext)]
	return filepath.Join(outputDir, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
}

func writeToFile(filePath, content string) {
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}
	log.Printf("Content successfully written to %s\n", filePath)
}

func readExistingTranscription(filePath string) string {
	log.Printf("Reading existing transcription file: %s\n", filePath)

	transcriptionBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading transcription file: %v", err)
	}

	return string(transcriptionBytes)
}

func createEmacsOrgNotes(transcriptionText, apiKey, baseFilePath string) {
	log.Println("Starting post-processing with create_emacs_org_notes command...")

	message := map[string]string{
		"role":    "user",
		"content": createPrompt(transcriptionText),
	}

	client := resty.New()
	client.SetTimeout(10 * time.Minute)

	reqBody := map[string]interface{}{
		"model":       "gpt-4",
		"messages":    []map[string]string{message},
		"max_tokens":  1500,
		"temperature": 0.7,
	}

	log.Println("Sending request to OpenAI API...")
	resp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey)).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post("https://api.openai.com/v1/chat/completions")
	if err != nil {
		log.Fatalf("Error sending request to OpenAI API: %v", err)
	}

	if resp.IsError() {
		log.Fatalf("Error response from OpenAI API: %v", resp.Error())
	}

	log.Println("Parsing OpenAI API response...")
	var aiResponse OpenAIResponse
	if err := json.Unmarshal(resp.Body(), &aiResponse); err != nil {
		log.Fatalf("Error unmarshalling OpenAI response: %v", err)
	}

	outputFilePath := generateOrgFilePath(baseFilePath)
	writeToFile(outputFilePath, aiResponse.Choices[0].Message.Content)
}

func generateOrgFilePath(baseFilePath string) string {
	dir := filepath.Dir(baseFilePath)
	baseName := strings.TrimSuffix(filepath.Base(baseFilePath), filepath.Ext(baseFilePath))
	orgFileName := fmt.Sprintf("%s_emacs_org_notes.org", baseName)
	return filepath.Join(dir, orgFileName)
}

func createPrompt(transcriptionText string) string {
	return fmt.Sprintf(`I need you to summarize the following content and convert it into an Emacs Org file format.
Please do not include any extra commentary or explanations.
Make sure the summary is detailed but concise, capturing the key points and providing enough explanation for each section. Avoid being too brief or overly terse.

The response should only contain the Emacs Org formatted output.

Use the following structure:

1. The file should have a #+title: and #+author: and #+date: header using today's date in the format like "<1999-10-04 Fri>"
2. Include a "Summary" section that gives a brief overview of the key points, try
3. Include a "Notes" section, with **subsections** that organize the content logically.

Here is the content to summarize:

%s

Please format the response as a valid Emacs Org file.`, transcriptionText)
}
