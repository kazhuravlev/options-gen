//nolint:testpackage
package optionsgen

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRun_Concurrent tests running the generator concurrently.
// This can help find race conditions and thread-safety issues.
func TestRun_Concurrent(t *testing.T) {
	sourceCode := `package test
type Options struct {
	Field1 string
	Field2 int
	Field3 bool
}`

	const numGoroutines = 10

	// Create separate temp directories for each goroutine to avoid file conflicts
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "options.go")
			outputFile := filepath.Join(tmpDir, "options_generated.go")

			if err := os.WriteFile(inputFile, []byte(sourceCode), 0o644); err != nil {
				errors <- fmt.Errorf("goroutine %d: failed to write input: %w", id, err)

				return
			}

			opts := NewOptions(
				WithVersion(fmt.Sprintf("test-v%d", id)),
				WithPackageName("test"),
				WithStructName("Options"),
				WithInFilename(inputFile),
				WithOutFilename(outputFile),
			)
			if err := Run(opts); err != nil {
				errors <- fmt.Errorf("goroutine %d: Run() failed: %w", id, err)

				return
			}

			// Verify output was created
			if _, err := os.Stat(outputFile); err != nil {
				errors <- fmt.Errorf("goroutine %d: output file not created: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		require.NoError(t, err)
	}
}
