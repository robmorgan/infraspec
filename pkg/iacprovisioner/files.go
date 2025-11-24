package iacprovisioner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyTerraformFolderToTemp copies the given folder to a temp folder and returns the path to the copied folder.
// This is useful for running Terraform operations in isolation without modifying the original source directory.
// It filters out state files, tfvars files, and hidden files (except .terraform-version and .terraform.lock.hcl).
func CopyTerraformFolderToTemp(folderPath string, tempFolderPrefix string) (string, error) {
	tmpDir := os.TempDir()
	destFolder, err := os.MkdirTemp(tmpDir, tempFolderPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	if err := CopyTerraformFolderToDest(folderPath, destFolder); err != nil {
		return "", fmt.Errorf("failed to copy terraform folder to temp: %w", err)
	}

	return destFolder, nil
}

// CopyTerraformFolderToDest copies the contents of the source folder to the destination folder.
// It filters out files that shouldn't be copied for clean Terraform testing:
// - Hidden files and directories (except .terraform-version and .terraform.lock.hcl)
// - Terraform state files (terraform.tfstate, terraform.tfstate.backup)
// - Terraform variable files (terraform.tfvars, terraform.tfvars.json)
func CopyTerraformFolderToDest(src, dest string) error {
	// Get the absolute path of the source folder
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of source: %w", err)
	}

	// Create the destination folder if it doesn't exist
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Walk through the source directory
	return filepath.Walk(srcAbs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path from the source
		relPath, err := filepath.Rel(srcAbs, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Check if this file/directory should be filtered out
		if shouldFilterTerraformPath(relPath, info) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Construct the destination path
		destPath := filepath.Join(dest, relPath)

		// Handle directories
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			return copySymlink(path, destPath)
		}

		// Handle regular files
		return copyFile(path, destPath, info.Mode())
	})
}

// shouldFilterTerraformPath returns true if the given path should be filtered out when copying Terraform folders.
func shouldFilterTerraformPath(relPath string, info os.FileInfo) bool {
	baseName := filepath.Base(relPath)

	// Filter out hidden files and directories, except for special Terraform files
	if strings.HasPrefix(baseName, ".") {
		// Allow .terraform-version and .terraform.lock.hcl
		if baseName == ".terraform-version" || baseName == ".terraform.lock.hcl" {
			return false
		}
		return true
	}

	// Filter out Terraform state files
	if baseName == "terraform.tfstate" || baseName == "terraform.tfstate.backup" {
		return true
	}

	// Filter out Terraform variable files
	if baseName == "terraform.tfvars" || baseName == "terraform.tfvars.json" {
		return true
	}

	return false
}

// copyFile copies a file from src to dest with the given file mode.
func copyFile(src, dest string, mode os.FileMode) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	// Create the destination file
	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dest, err)
	}
	defer destFile.Close()

	// Copy the contents
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents from %s to %s: %w", src, dest, err)
	}

	return nil
}

// copySymlink copies a symlink from src to dest.
func copySymlink(src, dest string) error {
	// Read the symlink target
	target, err := os.Readlink(src)
	if err != nil {
		return fmt.Errorf("failed to read symlink %s: %w", src, err)
	}

	// Create the symlink at the destination
	if err := os.Symlink(target, dest); err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", dest, target, err)
	}

	return nil
}
