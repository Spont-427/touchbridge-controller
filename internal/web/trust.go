package web

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxTrustedDevices = 64

// TrustStore persists hashes of per-device bearer tokens. Raw tokens are only
// returned once to the pairing client; the server never writes them to disk.
type TrustStore struct {
	mu       sync.Mutex
	path     string
	serverID string
	devices  []trustedDevice
}

type trustedDevice struct {
	TokenHash string    `json:"token_hash"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type trustStoreFile struct {
	Version  int             `json:"version"`
	ServerID string          `json:"server_id"`
	Devices  []trustedDevice `json:"devices"`
}

// OpenTrustStore loads or securely creates a persistent server identity and
// trusted-device database. The data directory and file are owner-only.
func OpenTrustStore(path string) (*TrustStore, error) {
	if strings.TrimSpace(path) == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("resolve user config directory: %w", err)
		}
		path = filepath.Join(base, "TouchBridge", "trusted-devices.json")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve trust store path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create trust store directory: %w", err)
	}
	store := &TrustStore{path: path}
	data, err := os.ReadFile(path)
	if err == nil {
		var saved trustStoreFile
		if err := json.Unmarshal(data, &saved); err != nil {
			return nil, fmt.Errorf("decode trust store: %w", err)
		}
		if saved.Version != 1 || !validServerID(saved.ServerID) || len(saved.Devices) > maxTrustedDevices {
			return nil, errors.New("trust store has an unsupported or invalid format")
		}
		for _, device := range saved.Devices {
			if len(device.TokenHash) != sha256.Size*2 {
				return nil, errors.New("trust store contains an invalid token hash")
			}
			if _, err := hex.DecodeString(device.TokenHash); err != nil {
				return nil, errors.New("trust store contains an invalid token hash")
			}
		}
		store.serverID = saved.ServerID
		store.devices = saved.Devices
		return store, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read trust store: %w", err)
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return nil, fmt.Errorf("generate server identity: %w", err)
	}
	store.serverID = hex.EncodeToString(id[:])
	if err := store.persistLocked(); err != nil {
		return nil, err
	}
	return store, nil
}

func validServerID(value string) bool {
	if len(value) != 32 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func (s *TrustStore) ServerID() string {
	if s == nil {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.serverID
}

// Issue creates a 256-bit token, stores only its SHA-256 hash, and returns the
// clear token once for secure storage by the client. Oldest devices are evicted
// only when the explicit trust limit is reached.
func (s *TrustStore) Issue(name string, now time.Time) (string, string, error) {
	if s == nil {
		return "", "", errors.New("trusted-device store is unavailable")
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", "", fmt.Errorf("generate device token: %w", err)
	}
	token := hex.EncodeToString(raw[:])
	hash := sha256.Sum256([]byte(token))
	name = strings.TrimSpace(name)
	if name == "" {
		name = "TouchBridge device"
	}
	if len(name) > 64 {
		name = name[:64]
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.devices) >= maxTrustedDevices {
		s.devices = append([]trustedDevice(nil), s.devices[1:]...)
	}
	s.devices = append(s.devices, trustedDevice{TokenHash: hex.EncodeToString(hash[:]), Name: name, CreatedAt: now.UTC()})
	if err := s.persistLocked(); err != nil {
		s.devices = s.devices[:len(s.devices)-1]
		return "", "", err
	}
	return token, s.serverID, nil
}

func (s *TrustStore) Verify(token string) bool {
	if s == nil || len(token) != 64 {
		return false
	}
	if _, err := hex.DecodeString(token); err != nil {
		return false
	}
	hash := sha256.Sum256([]byte(token))
	candidate := []byte(hex.EncodeToString(hash[:]))
	s.mu.Lock()
	defer s.mu.Unlock()
	valid := 0
	for _, device := range s.devices {
		valid |= subtle.ConstantTimeCompare(candidate, []byte(device.TokenHash))
	}
	return valid == 1
}

func (s *TrustStore) persistLocked() error {
	data, err := json.MarshalIndent(trustStoreFile{Version: 1, ServerID: s.serverID, Devices: s.devices}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode trust store: %w", err)
	}
	tmp := s.path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create trust store temp file: %w", err)
	}
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("write trust store: %w", err)
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("secure trust store: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace trust store: %w", err)
	}
	return nil
}
