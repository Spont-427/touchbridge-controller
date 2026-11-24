package web

import (
	"testing"
)

func TestSharedDesktopInputStaysHeldUntilLastControllerReleases(t *testing.T) {
	backend := &testBackend{}
	ownership := newInputOwnership(backend)
	playerOne := newControllerSession(backend, 0, ownership)
	playerTwo := newControllerSession(backend, 1, ownership)

	if err := playerOne.Key("W", true); err != nil {
		t.Fatal(err)
	}
	if err := playerTwo.Key("W", true); err != nil {
		t.Fatal(err)
	}
	if err := playerOne.MouseButton("left", true); err != nil {
		t.Fatal(err)
	}
	if err := playerTwo.MouseButton("left", true); err != nil {
		t.Fatal(err)
	}

	if got := backend.keyDowns.Load(); got != 1 {
		t.Fatalf("key down events=%d, want 1", got)
	}
	if got := backend.buttonDowns.Load(); got != 1 {
		t.Fatalf("mouse button down events=%d, want 1", got)
	}

	if err := playerTwo.ReleaseAll(); err != nil {
		t.Fatal(err)
	}
	keyDown, buttonDown := backend.desktopState("W", "left")
	if !keyDown || !buttonDown {
		t.Fatalf("shared input released while player one still holds it: key=%t button=%t", keyDown, buttonDown)
	}
	if got := backend.keyUps.Load(); got != 0 {
		t.Fatalf("key up events after player two disconnect=%d, want 0", got)
	}
	if got := backend.buttonUps.Load(); got != 0 {
		t.Fatalf("mouse button up events after player two disconnect=%d, want 0", got)
	}

	if err := playerOne.ReleaseAll(); err != nil {
		t.Fatal(err)
	}
	keyDown, buttonDown = backend.desktopState("W", "left")
	if keyDown || buttonDown {
		t.Fatalf("shared input still held after final owner released: key=%t button=%t", keyDown, buttonDown)
	}
	if got := backend.keyUps.Load(); got != 1 {
		t.Fatalf("key up events after final release=%d, want 1", got)
	}
	if got := backend.buttonUps.Load(); got != 1 {
		t.Fatalf("mouse button up events after final release=%d, want 1", got)
	}
}

func TestControllerCannotReleaseAnotherPlayersUnownedKey(t *testing.T) {
	backend := &testBackend{}
	ownership := newInputOwnership(backend)
	playerOne := newControllerSession(backend, 0, ownership)
	playerTwo := newControllerSession(backend, 1, ownership)

	if err := playerOne.Key("SPACE", true); err != nil {
		t.Fatal(err)
	}
	if err := playerTwo.Key("SPACE", false); err != nil {
		t.Fatal(err)
	}
	keyDown, _ := backend.desktopState("SPACE", "")
	if !keyDown {
		t.Fatal("player two released a key it did not own")
	}
	if got := backend.keyUps.Load(); got != 0 {
		t.Fatalf("key up events=%d, want 0", got)
	}
}
