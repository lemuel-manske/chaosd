// Package session provides APIs to store and retrieve sessions.
// Fault identification is generated on this layer.
package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"chaosd/cli/internal/storage"

	"github.com/google/uuid"
)

const (
	activeStatus = "active"
	healedStatus = "healed"

	partitionFaultType = "partition"

	sessionFileExt = ".json"

	sessionIDLength = 6
	sessionIDPrefix = "sess-"

	faultIDLength = 4
	faultIDPrefix = "fault-"
)

type FaultID string

func NewFaultID() FaultID {
	id := uuid.NewString()
	return FaultID(faultIDPrefix + id[:faultIDLength])
}

type Fault struct {
	ID     FaultID `json:"id"`
	Type   string  `json:"type"`
	NodeA  string  `json:"node_a"`
	NodeB  string  `json:"node_b"`
	Status string  `json:"status"`
}

func (f *Fault) IsPartition() bool {
	return f.Type == partitionFaultType
}

func (f *Fault) IsHealed() bool {
	return f.Status == healedStatus
}

type SessionID string

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

func (s *Session) GetFault(nodeAName string, nodeBName string) *Fault {
	for i := range s.Faults {
		fault := &s.Faults[i]
		if fault.NodeA == nodeAName && fault.NodeB == nodeBName {
			return fault
		}
		if fault.NodeA == nodeBName && fault.NodeB == nodeAName {
			return fault
		}
	}

	return nil
}

// SessionStore is responsible for managing sessions and their subordinates.
type SessionStore interface {
	AddPartitionFault(sessionID SessionID, nodeAName string, nodeBName string) (FaultID, error)
	Create(projectName string, composeFileAbsPath string) (*Session, error)
	Delete(id SessionID) error
	Get(id SessionID) (*Session, error)
	HealPartitionFault(sessionID SessionID, nodeAName string, nodeBName string) error
}

type FileSessionStore struct {
	writer storage.AtomicWriter
	dir    string
}

func NewFileSessionStore(dir string) SessionStore {
	writer := storage.NewAtomicFileWriter()

	return &FileSessionStore{dir: dir, writer: writer}
}

func (s *FileSessionStore) HealPartitionFault(
	sessionID SessionID,
	nodeAName string,
	nodeBName string,
) error {
	_session, err := s.Get(sessionID)
	if err != nil {
		return err
	}

	fault := _session.GetFault(nodeAName, nodeBName)
	if fault == nil {
		return fmt.Errorf("no partition fault found between %s and %s", nodeAName, nodeBName)
	}

	if fault.IsHealed() {
		return fmt.Errorf("partition fault between %s and %s is already healed", nodeAName, nodeBName)
	}

	// control fault healing, to now expose as a public API
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

func (s *FileSessionStore) AddPartitionFault(
	sessionID SessionID,
	nodeAName string,
	nodeBName string,
) (FaultID, error) {
	session, err := s.Get(sessionID)

	if err != nil {
		return "", err
	}

	fault := Fault{
		ID:     NewFaultID(),
		Type:   partitionFaultType,
		NodeA:  nodeAName,
		NodeB:  nodeBName,
		Status: activeStatus,
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
