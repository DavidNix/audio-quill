package main

import (
	"fmt"
	"os"
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
		Short: "Transcribe WAV files and generate markdown files",
		Long: `AudioQuill is a CLI tool that transcribes WAV files using Whisper (locally) and generates markdown files with titles using Ollama. 
No data leaves your machine.`,
	}

	var cfg config

	cmd.Flags().StringVarP(&cfg.SourceDir, "source", "s", "", "source directory with WAV files")
	_ = cmd.MarkFlagRequired("source")

	cmd.Flags().StringVarP(&cfg.DestDir, "dest", "d", "", "destination directory for generated markdown files")
	_ = cmd.MarkFlagRequired("dest")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		files, err := findWAVFiles(cfg.SourceDir)
		if err != nil {
			return fmt.Errorf("unable to find WAV source files: %w", err)
		}
		fmt.Println("found", files)
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
