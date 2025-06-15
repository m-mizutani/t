package file

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/config"
)

// CopyAction implements file.copy action
type CopyAction struct{}

// Name returns the action name
func (a *CopyAction) Name() string {
	return "file.copy"
}

// Description returns a human-readable description
func (a *CopyAction) Description() string {
	return "Copy a file or directory"
}

// Execute runs the file.copy action
func (a *CopyAction) Execute(ctx context.Context, actx *action.Context, step config.LegacyStepConfig) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing file.copy action")

	// Get source path
	srcArg, exists := step.Args["src"]
	if !exists {
		return nil, goerr.New("src argument is required for file.copy action")
	}

	srcStr, ok := srcArg.(string)
	if !ok {
		return nil, goerr.New("src argument must be string")
	}

	// Get destination path
	dstArg, exists := step.Args["dst"]
	if !exists {
		return nil, goerr.New("dst argument is required for file.copy action")
	}

	dstStr, ok := dstArg.(string)
	if !ok {
		return nil, goerr.New("dst argument must be string")
	}

	// Process templates
	src, err := action.ProcessTemplate(srcStr, actx, "file.copy.src")
	if err != nil {
		return nil, goerr.Wrap(err, "failed to process src template")
	}

	dst, err := action.ProcessTemplate(dstStr, actx, "file.copy.dst")
	if err != nil {
		return nil, goerr.Wrap(err, "failed to process dst template")
	}

	logger.Debug("Copying file/directory",
		slog.String("src", src),
		slog.String("dst", dst),
	)

	// Check if source exists
	srcInfo, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, goerr.New("source does not exist", goerr.Value("src", src))
		}
		return nil, goerr.Wrap(err, "failed to stat source", goerr.Value("src", src))
	}

	var bytesCopied int64
	if srcInfo.IsDir() {
		// Copy directory recursively
		bytesCopied, err = a.copyDir(src, dst)
	} else {
		// Copy single file
		bytesCopied, err = a.copyFile(src, dst)
	}

	if err != nil {
		return nil, goerr.Wrap(err, "failed to copy", goerr.Value("src", src), goerr.Value("dst", dst))
	}

	logger.Debug("Copy completed",
		slog.String("src", src),
		slog.String("dst", dst),
		slog.Int64("bytes_copied", bytesCopied),
		slog.Bool("is_directory", srcInfo.IsDir()),
	)

	return &action.Result{
		Output: dst,
		Metadata: map[string]interface{}{
			"src":          src,
			"dst":          dst,
			"bytes_copied": bytesCopied,
			"is_directory": srcInfo.IsDir(),
		},
	}, nil
}

// copyFile copies a single file
func (a *CopyAction) copyFile(src, dst string) (int64, error) {
	// Create destination directory if needed
	if dir := filepath.Dir(dst); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return 0, goerr.Wrap(err, "failed to create destination directory")
		}
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return 0, goerr.Wrap(err, "failed to open source file")
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, goerr.Wrap(err, "failed to create destination file")
	}
	defer dstFile.Close()

	copied, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return 0, goerr.Wrap(err, "failed to copy file content")
	}

	// Copy file permissions
	if srcInfo, err := srcFile.Stat(); err == nil {
		os.Chmod(dst, srcInfo.Mode())
	}

	return copied, nil
}

// copyDir recursively copies a directory
func (a *CopyAction) copyDir(src, dst string) (int64, error) {
	var totalBytes int64

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return goerr.Wrap(err, "failed to calculate relative path")
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// Copy file
			copied, err := a.copyFile(path, dstPath)
			if err != nil {
				return err
			}
			totalBytes += copied
			return nil
		}
	})

	return totalBytes, err
}
