package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
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

		for _, wav := range files {
			cmd.Printf("Processing %s ...\n", wav)
			if err = processFile(ctx, cfg.DestDir, wav); err != nil {
				return fmt.Errorf("failed to process %s: %w", wav, err)
			}

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

func processFile(ctx context.Context, destDir string, waveFile string) error {
	// TODO: temporary
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randomString := make([]byte, 10)
	for i := range randomString {
		randomString[i] = charset[rand.Intn(len(charset))]
	}

	transcribed, err := transcribe(ctx, waveFile)
	if err != nil {
		return err
	}

	fpath := filepath.Join(destDir, string(randomString))

	return os.WriteFile(fpath, transcribed, 0664)
}

func transcribe(ctx context.Context, waveFile string) ([]byte, error) {
	// Call bash because whisperfile is a script that switches on the arch
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
