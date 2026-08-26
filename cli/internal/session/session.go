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

func sessionsDir() (string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf("get user home dir: %w", err)
	}

	return filepath.Join(home, ".chaosd", "sessions"), nil
}

func sessionPath(id string) (string, error) {
	if id == "" {
		return "", errors.New("session id cannot be empty")
	}

	dir, err := sessionsDir()

	if err != nil {
		return "", err
	}

	return filepath.Join(dir, id+".json"), nil
}

func GetSession(id string) (*Session, error) {
	path, err := sessionPath(id)

	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("read session %q: %w", id, err)
	}

	var s Session

	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("decode session %q: %w", id, err)
	}

	return &s, nil
}

func CreateSession(
	projectName string,
	composeFileAbsPath string,
) (*Session, error) {
	id := uuid.NewString()

	path, err := sessionPath(id)

	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create sessions directory: %w", err)
	}

	s := &Session{
		ID:          id,
		Project:     projectName,
		ComposeFile: composeFileAbsPath,
	}

	data, err := json.MarshalIndent(s, "", "  ")

	if err != nil {
		return nil, fmt.Errorf("encode session: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return nil, fmt.Errorf("write session: %w", err)
	}

	return s, nil
}

func DeleteSession(id string) error {
	path, err := sessionPath(id)

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
