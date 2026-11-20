package web

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTrustStorePersistsHashAndDeviceIdentity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trust", "devices.json")
	store, err := OpenTrustStore(path)
	if err != nil {
		t.Fatal(err)
	}
	serverID := store.ServerID()
	token, issuedID, err := store.Issue("My Android", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if issuedID != serverID || !store.Verify(token) {
		t.Fatal("issued token or server identity did not validate")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" || string(data) == token {
		t.Fatal("trust store is empty or contains the clear token")
	}
	var saved trustStoreFile
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if len(saved.Devices) != 1 || saved.Devices[0].Name != "My Android" {
		t.Fatalf("unexpected saved devices: %+v", saved.Devices)
	}
	loaded, err := OpenTrustStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ServerID() != serverID || !loaded.Verify(token) {
		t.Fatal("token did not survive trust-store reload")
	}
}

func TestTrustStoreRejectsMalformedAndUnknownTokens(t *testing.T) {
	store, err := OpenTrustStore(filepath.Join(t.TempDir(), "devices.json"))
	if err != nil {
		t.Fatal(err)
	}
	if store.Verify("123456") || store.Verify(string(make([]byte, 64))) {
		t.Fatal("malformed token accepted")
	}
	token, _, err := store.Issue("phone", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if store.Verify(token[:62] + "00") {
		t.Fatal("modified token accepted")
	}
}
