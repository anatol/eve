// Copyright (c) 2022 Zededa, Inc.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

// MoveDir recursively move a directory tree, attempting to preserve permissions.
// if file exists in dst (sub)directory it will be replaced
func MoveDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error accessing src directory: %v", err)
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error accessing dst directory: %v", err)
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return fmt.Errorf("error creating dst directory: %v", err)
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return fmt.Errorf("error read dst directory: %v", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = MoveDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			if entry.Mode()&os.ModeSymlink != 0 {
				link, err := os.Readlink(srcPath)
				if err != nil {
					return fmt.Errorf("error readlink file: %v", err)
				}
				err = os.Symlink(link, dstPath)
				if err != nil {
					return fmt.Errorf("error symlink file: %v", err)
				}
				err = os.Remove(srcPath)
				if err != nil {
					return fmt.Errorf("error delete symlink: %v", err)
				}
				continue
			}
			_, err = os.Stat(dstPath)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("error accessing dst file: %v", err)
			}
			err = os.Remove(dstPath)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("error removing dst file: %v", err)
			}
			err = os.Rename(srcPath, dstPath)
			if err != nil {
				if linkErr, ok := err.(*os.LinkError); !ok || linkErr.Err != syscall.EXDEV {
					return fmt.Errorf("error renaming to new file: %v", err)
				}
				// special case for moving to another device
				err = CopyFile(srcPath, dstPath)
				if err != nil {
					return fmt.Errorf("error copy file: %v", err)
				}
				err = os.Remove(srcPath)
				if err != nil {
					return fmt.Errorf("error removing file: %v", err)
				}
			}
		}
	}
	return os.RemoveAll(src)
}
