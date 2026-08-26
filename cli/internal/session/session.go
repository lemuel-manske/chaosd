package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Session struct {
	ID          string `json:"id"`
	Project     string `json:"project"`
	ComposeFile string `json:"compose_file"`
}

const sessionFileExt = ".json"

type Store struct {
	dir string
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

func DefaultStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get user home dir: %w", err)
	}

	return NewStore(filepath.Join(home, ".chaosd", "sessions")), nil
}

func (s *Store) sessionPath(id string) (string, error) {
	if id == "" {
		return "", errors.New("session id cannot be empty")
	}

	return filepath.Join(s.dir, id+sessionFileExt), nil
}

func (s *Store) Get(id string) (*Session, error) {
	path, err := s.sessionPath(id)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read session %q: %w", id, err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("decode session %q: %w", id, err)
	}

	return &session, nil
}

func (s *Store) Create(
	projectName string,
	composeFileAbsPath string,
) (*Session, error) {
	id := uuid.NewString()

	path, err := s.sessionPath(id)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return nil, fmt.Errorf("create sessions directory: %w", err)
	}

	session := &Session{
		ID:          id,
		Project:     projectName,
		ComposeFile: composeFileAbsPath,
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode session: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return nil, fmt.Errorf("write session: %w", err)
	}

	return session, nil
}

func (s *Store) Delete(id string) error {
	path, err := s.sessionPath(id)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("delete session %q: %w", id, err)
	}

	return nil
}
