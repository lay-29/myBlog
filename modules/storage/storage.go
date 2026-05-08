package storage

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type AttachmentStore struct {
	Root string
}

type StoredFile struct {
	UUID   string
	Name   string
	Path   string
	Size   int64
	SHA256 string
}

func NewAttachmentStore(root string) *AttachmentStore {
	return &AttachmentStore{Root: root}
}

func (s *AttachmentStore) Save(name string, r io.Reader) (*StoredFile, error) {
	if err := os.MkdirAll(s.Root, 0o755); err != nil {
		return nil, err
	}
	uuid, err := randomID()
	if err != nil {
		return nil, err
	}
	cleanName := filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	if cleanName == "." || cleanName == string(filepath.Separator) || cleanName == "" {
		cleanName = "attachment"
	}

	tmp, err := os.CreateTemp(s.Root, "upload-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())

	hasher := sha256.New()
	size, err := io.Copy(io.MultiWriter(tmp, hasher), r)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}
	sum := hex.EncodeToString(hasher.Sum(nil))
	relDir := filepath.Join(sum[:2], sum[2:4])
	targetDir := filepath.Join(s.Root, relDir)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, err
	}
	relPath := filepath.Join(relDir, uuid+"-"+cleanName)
	target := filepath.Join(s.Root, relPath)
	if err := os.Rename(tmp.Name(), target); err != nil {
		return nil, err
	}
	return &StoredFile{
		UUID:   uuid,
		Name:   cleanName,
		Path:   filepath.ToSlash(relPath),
		Size:   size,
		SHA256: sum,
	}, nil
}

func randomID() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
