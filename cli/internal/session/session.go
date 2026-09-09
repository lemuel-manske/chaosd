// Package session provides APIs to store and retrieve sessions.
// Fault identification is generated on this layer.
package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"chaosd/cli/internal/network"
	"chaosd/cli/internal/storage"

	"github.com/google/uuid"
)

const (
	activeStatus = "active"
	healedStatus = "healed"

	partitionFaultType = "partition"
	delayFaultType     = "delay"

	sessionFileExt = ".json"

	sessionIDLength = 6
	sessionIDPrefix = "sess-"

	faultIDLength = 4
	faultIDPrefix = "fault-"
)

type FaultID string

func ParseFaultID(id string) FaultID {
	return FaultID(strings.TrimSpace(id))
}

func NewFaultID() FaultID {
	id := uuid.NewString()
	return FaultID(faultIDPrefix + id[:faultIDLength])
}

type Fault struct {
	ID      FaultID                 `json:"id"`
	Type    string                  `json:"type"`
	NodeA   string                  `json:"node_a"`
	NodeB   string                  `json:"node_b"`
	Status  string                  `json:"status"`
	Effects []network.AppliedEffect `json:"effects,omitempty"`
}

func NewPartitionFault(nodeA, nodeB string) Fault {
	return Fault{
		ID:      NewFaultID(),
		Type:    partitionFaultType,
		NodeA:   nodeA,
		NodeB:   nodeB,
		Status:  activeStatus,
	}
}

func NewDelayFault(nodeA, nodeB string) Fault {
	return Fault{
		ID:      NewFaultID(),
		Type:    delayFaultType,
		NodeA:   nodeA,
		NodeB:   nodeB,
		Status:  activeStatus,
	}
}

func (f *Fault) SetEffects(effects []network.AppliedEffect) {
	f.Effects = effects
}

func (f *Fault) IsDelay() bool {
	return f.Type == delayFaultType
}

func (f *Fault) IsPartition() bool {
	return f.Type == partitionFaultType
}

func (f *Fault) IsHealed() bool {
	return f.Status == healedStatus
}

type SessionID string

func ParseSessionID(id string) SessionID {
	return SessionID(strings.TrimSpace(id))
}

func NewSessionID() SessionID {
	id := uuid.NewString()
	return SessionID(sessionIDPrefix + id[:sessionIDLength])
}

// Session is only responsible to list faults, not add them.
// SessionStore API should be used to add faults to a session.
type Session struct {
	ID          SessionID `json:"id"`
	Project     string    `json:"project"`
	ComposeFile string    `json:"compose_file"`
	Faults      []Fault   `json:"faults"`
}

func (s *Session) GetFault(ID FaultID) *Fault {
	for i := range s.Faults {
		if s.Faults[i].ID == ID {
			return &s.Faults[i]
		}
	}

	return nil
}

// SessionStore is responsible for managing sessions and their subordinates.
type SessionStore interface {
	AddFault(sessionID SessionID, fault Fault) (FaultID, error)
	Create(projectName string, composeFileAbsPath string) (*Session, error)
	Delete(id SessionID) error
	Get(id SessionID) (*Session, error)
	HealFault(sessionID SessionID, faultID FaultID) error
}

type FileSessionStore struct {
	writer storage.AtomicWriter
	dir    string
}

func NewFileSessionStore(dir string) SessionStore {
	writer := storage.NewAtomicFileWriter()

	return &FileSessionStore{dir: dir, writer: writer}
}

func (s *FileSessionStore) AddFault(
	sessionID SessionID,
	fault Fault,
) (FaultID, error) {
	session, err := s.Get(sessionID)

	if err != nil {
		return "", err
	}

	session.Faults = append(session.Faults, fault)

	path, err := s.createPathToSession(sessionID)

	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(session, "", "  ")

	if err != nil {
		return "", fmt.Errorf("encode session: %w", err)
	}

	if err := s.writer.Write(path, data, 0600); err != nil {
		return "", fmt.Errorf("write session: %w", err)
	}

	return fault.ID, nil
}

func (s *FileSessionStore) HealFault(
	sessionID SessionID,
	faultID FaultID,
) error {
	_session, err := s.Get(sessionID)
	if err != nil {
		return err
	}

	fault := _session.GetFault(faultID)
	if fault == nil {
		return fmt.Errorf("fault %s not found", faultID)
	}

	if fault.IsHealed() {
		return fmt.Errorf("fault %s is already healed", faultID)
	}

	// control fault healing, to not expose as a public API
	fault.Status = healedStatus

	path, err := s.createPathToSession(sessionID)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(_session, "", "  ")
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}

	if err := s.writer.Write(path, data, 0600); err != nil {
		return fmt.Errorf("write session: %w", err)
	}

	return nil
}

func (s *FileSessionStore) Get(id SessionID) (*Session, error) {
	path, err := s.createPathToSession(id)

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

func (s *FileSessionStore) Create(
	projectName string,
	composeFileAbsPath string,
) (*Session, error) {
	id := NewSessionID()

	path, err := s.createPathToSession(id)

	if err != nil {
		return nil, err
	}

	if err = os.MkdirAll(s.dir, 0700); err != nil {
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

	if err := s.writer.Write(path, data, 0600); err != nil {
		return nil, fmt.Errorf("write session: %w", err)
	}

	return session, nil
}

func (s *FileSessionStore) Delete(id SessionID) error {
	path, err := s.createPathToSession(id)

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

func (s *FileSessionStore) createPathToSession(id SessionID) (string, error) {
	if id == "" {
		return "", errors.New("session id cannot be empty")
	}

	stringID := string(id)

	if !s.isValidSessionID(stringID) {
		return "", fmt.Errorf("invalid session id: %q", id)
	}

	return filepath.Join(s.dir, stringID+sessionFileExt), nil
}

func (s *FileSessionStore) isValidSessionID(id string) bool {
	return len(id) == len(sessionIDPrefix)+sessionIDLength && id[:len(sessionIDPrefix)] == sessionIDPrefix
}
