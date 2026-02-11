//go:build mage

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	// mage:import
	"github.com/grafana/grafana-plugin-sdk-go/build"
	"github.com/magefile/mage/mg"
)

// Default configures the default target.
var Default = BuildAndSign

// BuildAndSign runs the standard Grafana build and then generates .sig files.
func BuildAndSign() {
	mg.Deps(build.BuildAll) // This runs the standard Grafana build first
	mg.Deps(Sign)           // Then this runs the custom signature logic
}

// Sign generates .sig (sha256) files for all binaries in the build directory.
func Sign() error {
	fmt.Println(">> Generating .sig files for binaries...")

	files, err := filepath.Glob("dist/gpx_*") // Grafana SDK puts binaries in 'dist/'
	if err != nil {
		return err
	}

	for _, file := range files {
		// Skip files that are already .sig files
		if filepath.Ext(file) == ".sig" {
			continue
		}

		fmt.Printf("Signing %s...\n", file)
		hash, err := hashFile(file)
		if err != nil {
			return fmt.Errorf("failed to hash %s: %w", file, err)
		}

		sigFile := file + ".sig"
		if err := os.WriteFile(sigFile, []byte(hash), 0644); err != nil {
			return fmt.Errorf("failed to write sig file %s: %w", sigFile, err)
		}
	}
	return nil
}

// Helper to calculate SHA256 hex string
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
