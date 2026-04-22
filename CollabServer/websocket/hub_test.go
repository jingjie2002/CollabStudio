package websocket

import (
	"encoding/json"
	"testing"
)

func TestHostUnregisterDissolvesRoom(t *testing.T) {
	hub := NewHub()
	host := testClient("room-1", "111", "host-uuid")
	guest := testClient("room-1", "222", "guest-uuid")
	hub.rooms["room-1"] = &RoomData{
		Clients:      map[*Client]bool{host: true, guest: true},
		HostUUID:     host.UUID,
		HostUsername: host.Username,
	}

	hub.handleUnregister(host)

	if _, ok := hub.rooms["room-1"]; ok {
		t.Fatal("expected host unregister to dissolve and remove room")
	}

	msg := readWSMessage(t, guest.Send)
	if msg.Type != "room_closed" {
		t.Fatalf("expected room_closed, got %q", msg.Type)
	}
	if msg.RoomID != "room-1" {
		t.Fatalf("expected roomId room-1, got %q", msg.RoomID)
	}
}

func TestHostDissolveRoomBroadcastRemovesRoom(t *testing.T) {
	hub := NewHub()
	host := testClient("room-2", "111", "host-uuid")
	guest := testClient("room-2", "222", "guest-uuid")
	hub.rooms["room-2"] = &RoomData{
		Clients:      map[*Client]bool{host: true, guest: true},
		HostUUID:     host.UUID,
		HostUsername: host.Username,
	}

	hub.handleBroadcast(BroadcastMessage{
		RoomID:  "room-2",
		Message: []byte(`{"type":"dissolve_room"}`),
		Sender:  host,
	})

	if _, ok := hub.rooms["room-2"]; ok {
		t.Fatal("expected dissolve_room to remove room")
	}

	msg := readWSMessage(t, guest.Send)
	if msg.Type != "room_closed" {
		t.Fatalf("expected room_closed, got %q", msg.Type)
	}
}

func TestNonHostCannotDissolveRoom(t *testing.T) {
	hub := NewHub()
	host := testClient("room-3", "111", "host-uuid")
	guest := testClient("room-3", "222", "guest-uuid")
	hub.rooms["room-3"] = &RoomData{
		Clients:      map[*Client]bool{host: true, guest: true},
		HostUUID:     host.UUID,
		HostUsername: host.Username,
	}

	hub.handleBroadcast(BroadcastMessage{
		RoomID:  "room-3",
		Message: []byte(`{"type":"dissolve_room"}`),
		Sender:  guest,
	})

	if _, ok := hub.rooms["room-3"]; !ok {
		t.Fatal("expected non-host dissolve request to keep room")
	}

	msg := readWSMessage(t, guest.Send)
	if msg.Type != "error" {
		t.Fatalf("expected error message, got %q", msg.Type)
	}
}

func testClient(roomID, username, uuid string) *Client {
	return &Client{
		RoomID:   roomID,
		Username: username,
		UUID:     uuid,
		Send:     make(chan []byte, 4),
	}
}

func readWSMessage(t *testing.T, ch <-chan []byte) WSMessage {
	t.Helper()

	raw, ok := <-ch
	if !ok {
		t.Fatal("expected websocket message before channel close")
	}

	var msg WSMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal websocket message: %v", err)
	}
	return msg
}
