package cli_test

import (
	"testing"

	"github.com/rkriad585/PixalPeek/internal/cli"
)

func TestParseFlags(t *testing.T) {
	args := []string{"-qr", "sample.png", "--multi", "-o", "result.json"}
	cfg, err := cli.ParseFlags(args)
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}
	if cfg.DecodePath != "sample.png" {
		t.Errorf("DecodePath = %q; want sample.png", cfg.DecodePath)
	}
	if !cfg.Multi {
		t.Error("Multi = false; want true")
	}
	if cfg.OutputPath != "result.json" {
		t.Errorf("OutputPath = %q; want result.json", cfg.OutputPath)
	}
}

func TestParseFlagsGenerate(t *testing.T) {
	args := []string{"-g", "hello world", "-t", "text", "--size", "1024", "--ecc", "H"}
	cfg, err := cli.ParseFlags(args)
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}
	if cfg.GenerateText != "hello world" {
		t.Errorf("GenerateText = %q; want 'hello world'", cfg.GenerateText)
	}
	if cfg.TypeHint != "text" {
		t.Errorf("TypeHint = %q; want text", cfg.TypeHint)
	}
	if cfg.Size != 1024 {
		t.Errorf("Size = %d; want 1024", cfg.Size)
	}
	if cfg.ECC != "H" {
		t.Errorf("ECC = %q; want H", cfg.ECC)
	}
}

func TestParseFlagsShortFlags(t *testing.T) {
	args := []string{"-qr", "img.png", "-fg", "#ff0000", "-bg", "#ffffff", "-shape", "dot"}
	cfg, err := cli.ParseFlags(args)
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}
	if cfg.FGColor != "#ff0000" {
		t.Errorf("FGColor = %q; want #ff0000", cfg.FGColor)
	}
	if cfg.BGColor != "#ffffff" {
		t.Errorf("BGColor = %q; want #ffffff", cfg.BGColor)
	}
	if cfg.Shape != "dot" {
		t.Errorf("Shape = %q; want dot", cfg.Shape)
	}
}

func TestExitCodes(t *testing.T) {
	cfg := &cli.Config{
		DecodePath: "nonexistent_file_12345.png",
		ECC:        "M",
		FGColor:    "#000000",
		BGColor:    "#FFFFFF",
		Format:     "png",
		Size:       512,
		Shape:      "square",
	}
	code := cli.Execute(cfg)
	if code != cli.ExitFileNotFound {
		t.Errorf("Execute(non-existent file) exit code = %d; want %d", code, cli.ExitFileNotFound)
	}
}

func TestExitCodeConstants(t *testing.T) {
	if cli.ExitSuccess != 0 {
		t.Errorf("ExitSuccess = %d; want 0", cli.ExitSuccess)
	}
	if cli.ExitGeneralError != 1 {
		t.Errorf("ExitGeneralError = %d; want 1", cli.ExitGeneralError)
	}
	if cli.ExitInvalidArgs != 2 {
		t.Errorf("ExitInvalidArgs = %d; want 2", cli.ExitInvalidArgs)
	}
	if cli.ExitFileNotFound != 3 {
		t.Errorf("ExitFileNotFound = %d; want 3", cli.ExitFileNotFound)
	}
	if cli.ExitNoQRCodeDetected != 4 {
		t.Errorf("ExitNoQRCodeDetected = %d; want 4", cli.ExitNoQRCodeDetected)
	}
	if cli.ExitUnsupportedImage != 5 {
		t.Errorf("ExitUnsupportedImage = %d; want 5", cli.ExitUnsupportedImage)
	}
}
