package websocket

import (
	"collab-server/database"
	"collab-server/models"
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm/clause"
)

type RoomData struct {
	Clients map[*Client]bool
	Content string
}

type BroadcastMessage struct {
	RoomID  string
	Message []byte
	Sender  *Client
}

type WSMessage struct {
	Type    string           `json:"type"`
	Content string           `json:"content,omitempty"`
	Message string           `json:"message,omitempty"`
	Sender  string           `json:"sender,omitempty"`
	Users   []string         `json:"users,omitempty"`
	History []models.Message `json:"history,omitempty"`
	Cursor  int              `json:"cursor,omitempty"`
}

type Hub struct {
	rooms      map[string]*RoomData
	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastMessage
	dirtyRooms map[string]bool
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan BroadcastMessage, 1024), // 加大广播通道，应对高并发
		register:   make(chan *Client, 100),
		unregister: make(chan *Client, 100),
		rooms:      make(map[string]*RoomData),
		dirtyRooms: make(map[string]bool),
	}
}

func (h *Hub) saveDocumentToDB(roomID string, content string) {
	if roomID == "" {
		return
	}
	database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "room_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"content", "updated_at"}),
	}).Create(&models.Document{RoomID: roomID, Content: content})
}

func (h *Hub) saveVisitHistory(username, roomID string) {
	var history models.History
	result := database.DB.Where("username = ? AND room_id = ?", username, roomID).First(&history)
	if result.Error == nil {
		history.UpdatedAt = time.Now()
		database.DB.Save(&history)
	} else {
		database.DB.Create(&models.History{Username: username, RoomID: roomID, UpdatedAt: time.Now()})
	}
}

func (h *Hub) loadDocumentFromDB(roomID string) string {
	var doc models.Document
	if err := database.DB.Where("room_id = ?", roomID).First(&doc).Error; err != nil {
		return ""
	}
	return doc.Content
}

func (h *Hub) saveChatToDB(roomID, sender, message string) {
	database.DB.Create(&models.Message{RoomID: roomID, Sender: sender, Content: message})
}

func (h *Hub) loadChatHistory(roomID string) []models.Message {
	var messages []models.Message
	database.DB.Where("room_id = ?", roomID).Order("id desc").Limit(50).Find(&messages)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages
}

func (h *Hub) Run() {
	saveTicker := time.NewTicker(5 * time.Second)
	defer saveTicker.Stop()

	for {
		select {
		case client := <-h.register:
			roomID := client.RoomID
			if _, ok := h.rooms[roomID]; !ok {
				h.rooms[roomID] = &RoomData{Clients: make(map[*Client]bool), Content: h.loadDocumentFromDB(roomID)}
			}
			room := h.rooms[roomID]

			// 简单的踢人逻辑 (防止多开)
			for existingClient := range room.Clients {
				if existingClient.Username == client.Username {
					close(existingClient.Send)
					delete(room.Clients, existingClient)
				}
			}

			room.Clients[client] = true
			log.Printf("Join: %s (Room: %s)", client.Username, roomID)

			// 初始数据发送 (尽力而为)
			h.sendJSONToClient(client, "user_list", nil, h.getUserList(roomID))
			go h.saveVisitHistory(client.Username, roomID)

			if room.Content != "" {
				h.sendJSONToClient(client, "doc_update", room.Content, "System")
			}

			history := h.loadChatHistory(roomID)
			if len(history) > 0 {
				b, _ := json.Marshal(WSMessage{Type: "chat_history", History: history})
				select {
				case client.Send <- b:
				default:
				}
			}
			h.broadcastUserList(roomID)

		case client := <-h.unregister:
			roomID := client.RoomID
			if room, ok := h.rooms[roomID]; ok {
				if _, ok := room.Clients[client]; ok {
					delete(room.Clients, client)
					close(client.Send)
					h.broadcastUserList(roomID)
				}
			}

		case message := <-h.broadcast:
			if room, ok := h.rooms[message.RoomID]; ok {
				// 🟢 核心修复：绝对非阻塞广播
				// 我们宁愿丢弃某条消息，也不愿卡住整个 Hub
				for client := range room.Clients {
					select {
					case client.Send <- message.Message:
						// 发送成功
					default:
						// 缓冲区满，静默丢弃。
						// 因为文档是全量同步的，丢一包没关系，下一包会修正。
						// 只要不阻塞 Hub，其他人的体验就是流畅的。
					}
				}

				// 处理数据持久化
				var tmpMsg WSMessage
				if err := json.Unmarshal(message.Message, &tmpMsg); err == nil {
					if tmpMsg.Type == "doc_update" {
						room.Content = tmpMsg.Content
						h.dirtyRooms[message.RoomID] = true
					} else if tmpMsg.Type == "chat" {
						go h.saveChatToDB(message.RoomID, tmpMsg.Sender, tmpMsg.Message)
					}
				}
			}

		case <-saveTicker.C:
			for rID := range h.dirtyRooms {
				if room, ok := h.rooms[rID]; ok {
					go h.saveDocumentToDB(rID, room.Content)
				}
			}
			h.dirtyRooms = make(map[string]bool)
		}
	}
}

// 辅助：获取用户列表
func (h *Hub) getUserList(roomID string) []string {
	var list []string
	if room, ok := h.rooms[roomID]; ok {
		for c := range room.Clients {
			list = append(list, c.Username)
		}
	}
	return list
}

// 辅助：构建并发送JSON
func (h *Hub) sendJSONToClient(client *Client, msgType string, content interface{}, data interface{}) {
	msg := WSMessage{Type: msgType}
	if str, ok := content.(string); ok {
		msg.Content = str
	}
	if users, ok := data.([]string); ok {
		msg.Users = users
	}
	b, _ := json.Marshal(msg)

	select {
	case client.Send <- b:
	default:
	}
}

func (h *Hub) broadcastUserList(roomID string) {
	list := h.getUserList(roomID)
	b, _ := json.Marshal(WSMessage{Type: "user_list", Users: list})
	if room, ok := h.rooms[roomID]; ok {
		for c := range room.Clients {
			select {
			case c.Send <- b:
			default:
			}
		}
	}
}
