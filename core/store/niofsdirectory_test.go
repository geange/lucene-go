package store

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestNIOFSDirectory(t *testing.T) {

	dir, err := NewNIOFSDirectory(os.TempDir())
	if assert.Nil(t, err) {
		defer dir.Close()
	}

	dirPath, err := dir.GetDirectory()
	assert.Nil(t, err)

	assert.Equal(t, os.TempDir(), dirPath+"/")
}

func TestNIOFSDirectory_ObtainLock(t *testing.T) {
	dir, err := NewNIOFSDirectory(os.TempDir())
	if assert.Nil(t, err) {
		defer dir.Close()
	}

	lock, err := dir.ObtainLock("xxxx")
	assert.Nil(t, err)

	_, err = os.Stat(filepath.Join(os.TempDir(), "xxxx"))
	assert.Nil(t, err)

	err = lock.EnsureValid()
	assert.Nil(t, err)

	err = lock.Close()
	assert.Nil(t, err)

	_, err = os.Stat(filepath.Join(os.TempDir(), "xxxx"))
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestNIOFSDirectory_ListAll(t *testing.T) {
	{
		dirPath := filepath.Join(os.TempDir(), "demo")
		err := os.MkdirAll(dirPath, 0755)
		assert.Nil(t, err)

		fileA, err := os.Create(filepath.Join(dirPath, "a"))
		assert.Nil(t, err)
		defer fileA.Close()
		fileB, err := os.Create(filepath.Join(dirPath, "b"))
		assert.Nil(t, err)
		defer fileB.Close()
		fileC, err := os.Create(filepath.Join(dirPath, "c"))
		assert.Nil(t, err)
		defer fileC.Close()

		defer func() {
			os.Remove(filepath.Join(dirPath, "a"))
			os.Remove(filepath.Join(dirPath, "b"))
			os.Remove(filepath.Join(dirPath, "c"))
			os.RemoveAll(dirPath)
		}()

		dir, err := NewNIOFSDirectory(dirPath)
		if assert.Nil(t, err) {
			defer dir.Close()
		}

		items, err := dir.ListAll(context.Background())
		assert.Nil(t, err)

		assert.ElementsMatch(t, []string{"a", "b", "c"}, items)
	}

	// dir not found
	{
		dirPath := filepath.Join(os.TempDir(), "nodir")

		defer func() {
			os.RemoveAll(dirPath)
		}()

		dir, err := NewNIOFSDirectory(dirPath)
		assert.Nil(t, err)

		items, err := dir.ListAll(context.Background())
		assert.Nil(t, err)

		assert.ElementsMatch(t, []string{}, items)
	}

	// dir is a file
	{
		dirPath := filepath.Join(os.TempDir(), "test")
		file, err := os.Create(dirPath)
		assert.Nil(t, err)
		file.Close()

		defer func() {
			os.Remove(dirPath)
		}()

		_, err = NewNIOFSDirectory(dirPath)
		assert.NotNil(t, err)
	}
}

func TestNIOFSDirectory_NewFSIndexOutput(t *testing.T) {
	dir, err := NewNIOFSDirectory(os.TempDir())
	if assert.Nil(t, err) {
		defer dir.Close()
	}
	output, err := dir.NewFSIndexOutput("xxxx")
	assert.Nil(t, err)

	n, err := output.Write([]byte{2, 0, 4, 8})
	assert.Nil(t, err)
	assert.EqualValues(t, 4, n)
	output.Close()

	filePath := filepath.Join(os.TempDir(), "xxxx")

	stat, err := os.Stat(filePath)
	assert.Nil(t, err)
	assert.EqualValues(t, 4, stat.Size())

	content, err := os.ReadFile(filePath)
	assert.Nil(t, err)
	assert.EqualValues(t, []byte{2, 0, 4, 8}, content)

	err = os.Remove(filePath)
	assert.Nil(t, err)
}

func TestNIOFSDirectory_DeleteFile(t *testing.T) {
	dirPath := filepath.Join(os.TempDir(), "deleteDir")

	dir, err := NewNIOFSDirectory(dirPath)
	if assert.Nil(t, err) {
		defer dir.Close()
	}
	defer os.RemoveAll(dirPath)

	output, err := dir.CreateOutput(context.Background(), "x")
	assert.Nil(t, err)
	err = output.Close()
	assert.Nil(t, err)

	err = dir.DeleteFile(context.Background(), "x")
	assert.Nil(t, err)

	files, err := dir.ListAll(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{}, files)
}

func TestNIOFSDirectory_OpenInput(t *testing.T) {
	dirPath := filepath.Join(os.TempDir(), "openInput")

	_ = os.RemoveAll(dirPath)

	dir, err := NewNIOFSDirectory(dirPath)
	if assert.Nil(t, err) {
		defer dir.Close()
	}
	defer os.RemoveAll(dirPath)

	output, err := dir.CreateOutput(context.Background(), "x")
	assert.Nil(t, err)

	_, err = output.Write([]byte{1, 2, 3})
	assert.Nil(t, err)

	err = output.Close()
	assert.Nil(t, err)

	input, err := dir.OpenInput(context.Background(), "x")
	assert.Nil(t, err)
	bs := make([]byte, 3)
	_, err = input.Read(bs)
	assert.Nil(t, err)
	assert.Equal(t, []byte{1, 2, 3}, bs)
	input.Close()
}

func TestNIOFSDirectory_Rename(t *testing.T) {
	dirPath := filepath.Join(os.TempDir(), "rename")

	_ = os.RemoveAll(dirPath)

	dir, err := NewNIOFSDirectory(dirPath)
	if assert.Nil(t, err) {
		defer dir.Close()
	}
	defer os.RemoveAll(dirPath)

	output, err := dir.CreateOutput(context.Background(), "x")
	assert.Nil(t, err)

	_, err = output.Write([]byte{1, 2, 3})
	assert.Nil(t, err)

	err = output.Close()
	assert.Nil(t, err)

	err = dir.Rename(context.Background(), "x", "y")
	assert.Nil(t, err)

	input, err := dir.OpenInput(context.Background(), "y")
	assert.Nil(t, err)
	bs := make([]byte, 3)
	_, err = input.Read(bs)
	assert.Nil(t, err)
	assert.Equal(t, []byte{1, 2, 3}, bs)
	input.Close()
}

func TestNIOFSDirectory_FileLength(t *testing.T) {
	dirPath := filepath.Join(os.TempDir(), "fileLength")

	_ = os.RemoveAll(dirPath)

	dir, err := NewNIOFSDirectory(dirPath)
	if assert.Nil(t, err) {
		defer dir.Close()
	}
	defer os.RemoveAll(dirPath)

	output, err := dir.CreateOutput(context.Background(), "x")
	assert.Nil(t, err)

	_, err = output.Write([]byte{1, 2, 3, 7})
	assert.Nil(t, err)

	err = output.Close()
	assert.Nil(t, err)

	fileLength, err := dir.FileLength(context.Background(), "x")
	assert.Nil(t, err)

	assert.EqualValues(t, 4, fileLength)
}

func TestNIOFSDirectory_CreateTempOutput(t *testing.T) {
	dirPath := filepath.Join(os.TempDir(), "createTempOutput")

	_ = os.RemoveAll(dirPath)

	dir, err := NewNIOFSDirectory(dirPath)
	if assert.Nil(t, err) {
		defer dir.Close()
	}
	defer os.RemoveAll(dirPath)

	output, err := dir.CreateTempOutput(context.Background(), "x", "z")
	assert.Nil(t, err)
	output.Close()
	file1Name := filepath.Join(dirPath, fmt.Sprintf("%s_%s_%d.tmp", "x", "z", 1))
	_, err = os.Stat(file1Name)
	assert.Nil(t, err)

	output1, err := dir.CreateTempOutput(context.Background(), "x", "z")
	assert.Nil(t, err)
	output1.Close()
	file2Name := filepath.Join(dirPath, fmt.Sprintf("%s_%s_%d.tmp", "x", "z", 2))
	_, err = os.Stat(file2Name)
	assert.Nil(t, err)

	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("%s_%s_%d.tmp", "x", "z", 3)))
	assert.Nil(t, err)
	file.Close()

	output2, err := dir.CreateTempOutput(context.Background(), "x", "z")
	assert.Nil(t, err)
	output2.Close()
	file3Name := filepath.Join(dirPath, fmt.Sprintf("%s_%s_%d.tmp", "x", "z", 4))
	_, err = os.Stat(file3Name)
	assert.Nil(t, err)
}
