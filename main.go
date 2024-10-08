package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	if err := root().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type config struct {
	SourceDir string
	DestDir   string
}

func root() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audioquill",
		Short: "Transcribe WAV files and generate text files",
		Long: `AudioQuill is a CLI tool that transcribes WAV files using Whisper (locally) and generates text files with titles using Ollama. 
No data leaves your machine.`,
	}

	var cfg config

	cmd.Flags().StringVarP(&cfg.SourceDir, "source", "s", "", "source directory with WAV files")
	_ = cmd.MarkFlagRequired("source")

	cmd.Flags().StringVarP(&cfg.DestDir, "dest", "d", "", "destination directory for generated text files")
	_ = cmd.MarkFlagRequired("dest")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := os.MkdirAll(cfg.DestDir, 0777); err != nil {
			return err
		}

		files, err := findWAVFiles(cfg.SourceDir)
		if err != nil {
			return fmt.Errorf("unable to find WAV source files: %w", err)
		}
		cmd.Printf("Found %d WAV files\n", len(files))

		ctx := cmd.Context()

		for i, wav := range files {
			cmd.Printf("Processing %d. %s ...\n", i+1, wav)
			fname, err := processFile(ctx, cfg.DestDir, wav)
			if err != nil {
				return fmt.Errorf("failed to process %s: %w", wav, err)
			}
			cmd.Printf("\tSaved file %s\n", fname)
		}
		return nil
	}

	return cmd
}

func findWAVFiles(source string) ([]string, error) {
	var wavFiles []string
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".wav" {
			wavFiles = append(wavFiles, path)
		}
		return nil
	})
	return wavFiles, err
}

func processFile(ctx context.Context, destDir string, waveFile string) (string, error) {
	transcribed, err := transcribe(ctx, waveFile)
	if err != nil {
		return "", err
	}

	summary, err := ollamaTitleSummary(ctx, string(transcribed))
	if err != nil {
		return "", fmt.Errorf("failed to create summary: %w", err)
	}

	fname := cleanFilename(summary) + ".md"
	fpath := filepath.Join(destDir, fname)
	return fname, os.WriteFile(fpath, transcribed, 0664)
}

var re = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func cleanFilename(s string) string {
	cleaned := strings.TrimSpace(strings.ToLower(s))
	cleaned = re.ReplaceAllString(cleaned, "")
	cleaned = strings.Join(strings.Split(cleaned, " "), "-")
	if len(cleaned) > 50 {
		cleaned = cleaned[:50]
	}
	return cleaned
}

func transcribe(ctx context.Context, waveFile string) ([]byte, error) {
	// Call through bash because whisperfile is a script that switches on the arch, otherwise you get an exec format error
	cmd := exec.CommandContext(ctx, "bash", "-c", `./whisper-tiny.en.llamafile -f "`+waveFile+`" --no-prints`)
	cmd.Stderr = os.Stderr
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("whisper failed %w", err)
	}
	b := removeTimestamps(buf.Bytes())
	return b, nil
}

func removeTimestamps(input []byte) []byte {
	lines := bytes.Split(input, []byte("\n"))
	var result [][]byte

	for _, line := range lines {
		// Remove timestamp at the beginning of each line
		if idx := bytes.Index(line, []byte("]")); idx != -1 && idx < len(line)-1 {
			result = append(result, bytes.TrimSpace(line[idx+1:]))
		}
	}

	return bytes.Join(result, []byte("\n"))
}

func ollamaTitleSummary(ctx context.Context, content string) (string, error) {
	prompt := fmt.Sprintf(`Summarize the following content. Your summary will be used in a file name. Keep it short with 3 to 7 words. Only respond with the summary. Do not elaborate.
	CONTENT:
	%s`, content)

	const model = "llama3.1"
	rbody, err := json.Marshal(map[string]any{
		"system": "You are a helpful summarizer.",
		"model":  model,
		"prompt": prompt,
		"stream": false,
	})
	if err != nil {
		panic(err)
	}

	const endpoint = "http://localhost:11434/api/generate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(rbody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	result := struct {
		Response string `json:"response"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode Ollama response: %w", err)
	}

	return result.Response, nil
}
