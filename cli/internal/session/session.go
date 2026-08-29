package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type SessionID string

const (
	activeStatus = "active"
	healedStatus = "healed"

	partitionFaultType = "partition"
)

type Fault struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	NodeA  string `json:"node_a"`
	NodeB  string `json:"node_b"`
	Status string `json:"status"`
}

func (f *Fault) IsPartition() bool {
	return f.Type == partitionFaultType
}

func (f *Fault) IsHealed() bool {
	return f.Status == healedStatus
}

func (f *Fault) Heal() bool {
	if f.Status == healedStatus {
		return false
	}

	f.Status = healedStatus

	return true
}

type Session struct {
	ID          SessionID `json:"id"`
	Project     string    `json:"project"`
	ComposeFile string    `json:"compose_file"`
	Faults      []Fault   `json:"faults"`
}

func (s *Session) AddFault(fault Fault) {
	s.Faults = append(s.Faults, fault)
}

func (s *Session) FindFault(nodeAName string, nodeBName string) *Fault {
	for i := range s.Faults {
		fault := &s.Faults[i]
		if fault.NodeA == nodeAName && fault.NodeB == nodeBName {
			return fault
		}
	}

	return nil
}

const sessionFileExt = ".json"

type Store interface {
	AddPartitionFault(sessionID SessionID, nodeAName string, nodeBName string) error
	Create(projectName string, composeFileAbsPath string) (*Session, error)
	Delete(id SessionID) error
	Get(id SessionID) (*Session, error)
	HealPartitionFault(sessionID SessionID, nodeAName string, nodeBName string) error
}

func (s *concreteStore) HealPartitionFault(
	sessionID SessionID,
	nodeAName string,
	nodeBName string,
) error {
	_session, err := s.Get(sessionID)
	if err != nil {
		return err
	}

	fault := _session.FindFault(nodeAName, nodeBName)
	if fault == nil {
		return fmt.Errorf("no partition fault found between %s and %s", nodeAName, nodeBName)
	}

	if !fault.Heal() {
		return fmt.Errorf("partition fault between %s and %s is already healed", nodeAName, nodeBName)
	}

	path, err := s.sessionPath(sessionID)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(_session, "", "  ")
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write session: %w", err)
	}

	return nil
}

type concreteStore struct {
	dir string
}

func NewStore(dir string) Store {
	return &concreteStore{dir: dir}
}

func NewDefaultStore() (Store, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return nil, fmt.Errorf("get user home dir: %w", err)
	}

	return NewStore(filepath.Join(home, ".chaosd", "sessions")), nil
}

func (s *concreteStore) AddPartitionFault(
	sessionID SessionID,
	nodeAName string,
	nodeBName string,
) error {
	session, err := s.Get(sessionID)

	if err != nil {
		return err
	}

	fault := Fault{
		ID:     uuid.NewString(),
		Type:   partitionFaultType,
		NodeA:  nodeAName,
		NodeB:  nodeBName,
		Status: activeStatus,
	}

	session.AddFault(fault)

	path, err := s.sessionPath(sessionID)

	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(session, "", "  ")

	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write session: %w", err)
	}

	return nil
}

func (s *concreteStore) Get(id SessionID) (*Session, error) {
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

func (s *concreteStore) Create(
	projectName string,
	composeFileAbsPath string,
) (*Session, error) {
	id := uuid.NewString()

	path, err := s.sessionPath(SessionID(id))

	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return nil, fmt.Errorf("create sessions directory: %w", err)
	}

	session := &Session{
		ID:          SessionID(id),
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

func (s *concreteStore) Delete(id SessionID) error {
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

func (s *concreteStore) sessionPath(id SessionID) (string, error) {
	if id == "" {
		return "", errors.New("session id cannot be empty")
	}

	return filepath.Join(s.dir, string(id)+sessionFileExt), nil
}
