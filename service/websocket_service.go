package service

import (
	"chat-server/constant"
	"chat-server/global"
	"chat-server/model"
	"chat-server/utils"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// WebSocket消息结构
type WebSocketMessage struct {
	Type      string      `json:"type"`
	RoomId    string      `json:"room_id"`
	SenderId  string      `json:"sender_id"`
	Content   interface{} `json:"content"`
	CreatedAt int64       `json:"created_at"`
}

// 客户端
type Client struct {
	Conn     *websocket.Conn
	UserId   string
	RoomId   string
	Send     chan *WebSocketMessage
	LastPing time.Time
	Manager  *WebSocketManager
	mu       sync.Mutex
}

// WebSocket管理器
type WebSocketManager struct {
	Rooms      map[string]map[*Client]bool
	Clients    map[string][]*Client // 按用户ID组织的客户端映射（一个用户可能有多个连接，多平台）
	Broadcast  chan *WebSocketMessage
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

// NewWebSocketManager 创建一个新的WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		Rooms:      make(map[string]map[*Client]bool),
		Clients:    make(map[string][]*Client),
		Broadcast:  make(chan *WebSocketMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (manager *WebSocketManager) Run(ctx context.Context) {
	for {
		select {
		// 接收关闭信号
		case <-ctx.Done():
			global.CHAT_LOG.Info("WebSocket管理器收到关闭信号，正在关闭...")
			return
		// 注册用户
		case client := <-manager.Register:
			manager.mu.Lock()
			// 判断房间是否存在
			if _, ok := manager.Rooms[client.RoomId]; !ok {
				manager.Rooms[client.RoomId] = make(map[*Client]bool)
			}
			// 添加客户端
			manager.Rooms[client.RoomId][client] = true
			// 将客户端添加到用户映射
			manager.Clients[client.UserId] = append(manager.Clients[client.UserId], client)
			manager.mu.Unlock()

			// 发送加入消息
			joinMsg := &WebSocketMessage{
				Type:      constant.MessageTypeJoin,
				RoomId:    client.RoomId,
				SenderId:  client.UserId,
				Content:   map[string]interface{}{constant.MessageTypeJoin: constant.JoinMessageContent, constant.MessageTypeLeave: nil, constant.MessageTypeSystem: nil},
				CreatedAt: utils.GetUTCMillisTimestamp(),
			}
			manager.BroadcastToRoom(client.RoomId, joinMsg)

			// 将用户加入到redis
			pipeline := global.CHAT_REDIS.TxPipeline()
			ctx := context.Background()
			cacheKey := fmt.Sprintf("online_users:%s", client.RoomId)
			addCmd := pipeline.SAdd(ctx, cacheKey, client.UserId)
			expireCmd := pipeline.Expire(ctx, cacheKey, constant.OnlineUserExpire)
			if _, err := pipeline.Exec(ctx); err != nil {
				global.CHAT_LOG.Error("WebSocket Run----->添加在线用户失败", "err", err.Error())
				return
			}
			if err := addCmd.Err(); err != nil {
				global.CHAT_LOG.Error("WebSocket Run----->添加在线用户失败", "err", err.Error())
				return
			}
			if err := expireCmd.Err(); err != nil {
				global.CHAT_LOG.Error("WebSocket Run----->添加在线用户过期时间失败", "err", err.Error())
				return
			}
		// 注销用户
		case client := <-manager.Unregister:
			manager.mu.Lock()
			// 从房间中移除客户端
			if _, roomExist := manager.Rooms[client.RoomId]; roomExist {
				if _, clientExist := manager.Rooms[client.RoomId][client]; clientExist {
					delete(manager.Rooms[client.RoomId], client)
					close(client.Send)
					// 房间没人则删除房间
					if len(manager.Rooms[client.RoomId]) == 0 {
						delete(manager.Rooms, client.RoomId)
					}
				}
			}

			// 从用户映射中移除客户端
			if clients, exists := manager.Clients[client.UserId]; exists {
				var newClients []*Client
				for _, c := range clients {
					if c != client {
						newClients = append(newClients, c)
					}
				}
				// 客户端全部离线
				if len(newClients) == 0 {
					delete(manager.Clients, client.UserId)
					// 从redis中移除用户
					cacheKey := fmt.Sprintf("online_users:%s", client.RoomId)
					global.CHAT_REDIS.SRem(context.Background(), cacheKey, client.UserId)
				} else {
					manager.Clients[client.UserId] = newClients
				}
			}
			manager.mu.Unlock()

			// 发送离线消息
			leaveMsg := &WebSocketMessage{
				Type:      constant.MessageTypeLeave,
				RoomId:    client.RoomId,
				SenderId:  client.UserId,
				Content:   map[string]interface{}{constant.MessageTypeLeave: constant.LeaveMessageContent, constant.MessageTypeJoin: nil, constant.MessageTypeSystem: nil},
				CreatedAt: utils.GetUTCMillisTimestamp(),
			}
			manager.BroadcastToRoom(client.RoomId, leaveMsg)
		// 广播消息
		case message := <-manager.Broadcast:
			manager.BroadcastToRoom(message.RoomId, message)
		}
	}
}

func (manager *WebSocketManager) BroadcastToRoom(roomId string, message *WebSocketMessage) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// 向指定房间发送消息
	if clients, exists := manager.Rooms[roomId]; exists {
		for client := range clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(manager.Rooms[roomId], client)
			}
		}
	}

	// 保存用户消息到mongoDB
	if constant.UserMessageType[message.Type] {
		go func(message *WebSocketMessage) {
			// 验证消息内容
			content, isValid := validateUserMessage(message)
			if !isValid {
				global.CHAT_LOG.Error("WebSocket BroadcastToRoom----->消息内容验证失败", "message", message)
			}
			// 保存消息到mongoDB
			mongoMsg := model.UserMessages{
				ID:        bson.NewObjectID(),
				RoomId:    message.RoomId,
				SenderId:  message.SenderId,
				Type:      message.Type,
				Content:   content,
				CreatedAt: message.CreatedAt,
			}
			if _, err := global.CHAT_MONGODB.Collection("user_messages").InsertOne(context.Background(), mongoMsg); err != nil {
				global.CHAT_LOG.Error("WebSocket BroadcastToRoom----->保存消息到MongoDB失败", "err", err.Error())
			}

		}(message)
	}
	// 保存系统消息到mongoDB
	if constant.SystemMessageType[message.Type] {
		go func(message *WebSocketMessage) {
			// 验证消息内容
			content, isValid := validateSystemMessage(message)
			if !isValid {
				global.CHAT_LOG.Error("WebSocket BroadcastToRoom----->消息内容验证失败", "message", message)
			}
			// 保存消息到mongoDB
			mongoMsg := model.SystemMessages{
				ID:        bson.NewObjectID(),
				RoomId:    message.RoomId,
				SenderId:  message.SenderId,
				Type:      message.Type,
				Content:   content,
				CreatedAt: message.CreatedAt,
			}
			if _, err := global.CHAT_MONGODB.Collection("system_messages").InsertOne(context.Background(), mongoMsg); err != nil {
				global.CHAT_LOG.Error("WebSocket BroadcastToRoom----->保存消息到MongoDB失败", "err", err.Error())
			}

		}(message)
	}

}

func (client *Client) ReadPump() {
	global.CHAT_LOG.Info("ReadPump 开始读取消息")
	defer func() {
		client.Manager.Unregister <- client
		client.Conn.Close()
		global.CHAT_LOG.Info("ReadPump 读取消息结束")
	}()

	// 设置读取时长和心跳
	client.Conn.SetReadDeadline(time.Now().Add(time.Second * 60))
	client.Conn.SetPongHandler(func(string) error {
		client.mu.Lock()
		client.LastPing = time.Now()
		client.mu.Unlock()
		client.Conn.SetReadDeadline(time.Now().Add(time.Second * 60))
		return nil
	})

	// 读取消息
	for {
		// 获取消息
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				global.CHAT_LOG.Error("ReadPump WebSocket读取消息错误", "err", err)
			}
			break
		}
		// 接收json格式的message，解析成WebSocketMessage类型。
		var wsMessage WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			global.CHAT_LOG.Error("WebSocket解析消息错误", "err", err, "message", message)
			// 解析错误则默认当作文本处理
			wsMessage.Type = constant.MessageTypeText
			wsMessage.Content = map[string]interface{}{"text": string(message)}
		}
		// 解析后json后，设置基本信息
		wsMessage.RoomId = client.RoomId
		wsMessage.SenderId = client.UserId
		wsMessage.CreatedAt = utils.GetUTCMillisTimestamp()
		// 发送消息
		client.Manager.Broadcast <- &wsMessage
	}
}

func (client *Client) WritePump() {
	// 设置心跳定时器
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			// 设置写入时长
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// 通道已关闭
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// 将消息编码为JSON
			jsonMessage, err := json.Marshal(message)
			if err != nil {
				global.CHAT_LOG.Error("WritePump 编码WebSocket消息失败", "err", err)
				return
			}
			// 设置websocket写入器
			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				global.CHAT_LOG.Error("WritePump 设置写入器错误", "err", err)
			}
			w.Write(jsonMessage)
			// 检查是否还有别的消息
			n := len(client.Send)
			for i := 0; i < n; i++ {
				jsonMessage, err = json.Marshal(<-client.Send)
				if err != nil {
					global.CHAT_LOG.Error("WritePump 编码WebSocket消息失败", "err", err)
					return
				}
				w.Write([]byte("\n"))
				w.Write(jsonMessage)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// 发送ping消息保持连接活跃
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func validateUserMessage(message *WebSocketMessage) (model.UserMessageContent, bool) {
	// 验证id
	if message.RoomId == "" {
		global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息id不能为空", "message", message)
		return model.UserMessageContent{}, false
	}
	if message.SenderId == "" {
		global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息发送者id不能为空", "message", message)
		return model.UserMessageContent{}, false
	}
	if message.CreatedAt == 0 {
		global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息创建时间不能为空", "message", message)
		return model.UserMessageContent{}, false
	}
	if message.Type == "" {
		global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息类型不能为空", "message", message)
		return model.UserMessageContent{}, false
	}
	// 验证消息内容
	content := model.UserMessageContent{}
	switch message.Type {
	case constant.MessageTypeText:
		if contentMap, ok := message.Content.(map[string]interface{}); ok {
			text := utils.GetStringValue(contentMap, "text")
			content.Text = &text
			if !validateUserContent(content, message.Type) {
				global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息内容验证失败", "message", message)
				return model.UserMessageContent{}, false
			}
		} else {
			global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息内容验证失败", "message", message)
			return model.UserMessageContent{}, false
		}
	case constant.MessageTypeImage:
		if contentMap, ok := message.Content.(map[string]interface{}); ok {
			imageMap := utils.GetMapValue(contentMap, "image")
			if imageMap == nil {
				global.CHAT_LOG.Error("WebSocket validateUserMessage----->图片消息内容无效", "message", message)
				return model.UserMessageContent{}, false
			}
			image := model.UserMessageContentImage{
				URL:    utils.GetStringValue(imageMap, "url"),
				Name:   utils.GetStringValue(imageMap, "name"),
				Format: utils.GetStringValue(imageMap, "format"),
				Size:   utils.GetIntValue(imageMap, "size"),
			}
			content.Image = &image
			if !validateUserContent(content, message.Type) {
				global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息内容验证失败", "message", message)
				return model.UserMessageContent{}, false
			}
		} else {
			global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息内容验证失败", "message", message)
			return model.UserMessageContent{}, false
		}
	case constant.MessageTypeFile:
		if contentMap, ok := message.Content.(map[string]interface{}); ok {
			fileMap := utils.GetMapValue(contentMap, "file")
			if fileMap == nil {
				global.CHAT_LOG.Error("WebSocket validateUserMessage----->文件消息内容无效", "message", message)
				return model.UserMessageContent{}, false
			}

			file := model.UserMessageContentFile{
				URL:    utils.GetStringValue(fileMap, "url"),
				Name:   utils.GetStringValue(fileMap, "name"),
				Format: utils.GetStringValue(fileMap, "format"),
				Size:   utils.GetIntValue(fileMap, "size"),
			}
			content.File = &file
			if !validateUserContent(content, message.Type) {
				global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息内容验证失败", "message", message)
				return model.UserMessageContent{}, false
			}
		} else {
			global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息内容验证失败", "message", message)
			return model.UserMessageContent{}, false
		}
	case constant.MessageTypeVoice:
		if contentMap, ok := message.Content.(map[string]interface{}); ok {
			voiceMap := utils.GetMapValue(contentMap, "voice")
			if voiceMap == nil {
				global.CHAT_LOG.Error("WebSocket validateUserMessage----->语音消息内容无效", "message", message)
				return model.UserMessageContent{}, false
			}

			voice := model.UserMessageContentVoice{
				URL:      utils.GetStringValue(voiceMap, "url"),
				Name:     utils.GetStringValue(voiceMap, "name"),
				Format:   utils.GetStringValue(voiceMap, "format"),
				Size:     utils.GetIntValue(voiceMap, "size"),
				Duration: utils.GetFloatValue(voiceMap, "duration"),
			}
			content.Voice = &voice
			if !validateUserContent(content, message.Type) {
				global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息内容验证失败", "message", message)
				return model.UserMessageContent{}, false
			}
		} else {
			global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息内容验证失败", "message", message)
			return model.UserMessageContent{}, false
		}
	case constant.MessageTypeVideo:
		if contentMap, ok := message.Content.(map[string]interface{}); ok {
			videoMap := utils.GetMapValue(contentMap, "video")
			if videoMap == nil {
				global.CHAT_LOG.Error("WebSocket validateUserMessage----->视频消息内容无效", "message", message)
				return model.UserMessageContent{}, false
			}

			video := model.UserMessageContentVideo{
				URL:      utils.GetStringValue(videoMap, "url"),
				Name:     utils.GetStringValue(videoMap, "name"),
				Format:   utils.GetStringValue(videoMap, "format"),
				Size:     utils.GetIntValue(videoMap, "size"),
				Duration: utils.GetFloatValue(videoMap, "duration"),
			}
			content.Video = &video
			if !validateUserContent(content, message.Type) {
				global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息内容验证失败", "message", message)
				return model.UserMessageContent{}, false
			}
		} else {
			global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息内容验证失败", "message", message)
			return model.UserMessageContent{}, false
		}
	case constant.MessageTypeReply:
		if contentMap, ok := message.Content.(map[string]interface{}); ok {
			replyMap := utils.GetMapValue(contentMap, "reply")
			if replyMap == nil {
				global.CHAT_LOG.Error("WebSocket validateUserMessage----->回复消息内容无效", "message", message)
				return model.UserMessageContent{}, false
			}

			reply := model.UserMessageContentReply{
				Text:    utils.GetStringValue(replyMap, "text"),
				ReplyTo: utils.GetStringValue(replyMap, "reply_to"),
			}
			content.Reply = &reply
			if !validateUserContent(content, message.Type) {
				global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息内容验证失败", "message", message)
				return model.UserMessageContent{}, false
			}
		} else {
			global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息内容验证失败", "message", message)
			return model.UserMessageContent{}, false
		}
	default:
		global.CHAT_LOG.Error("WebSocket validateUserMessage----->未知消息类型", "message", message)
		return model.UserMessageContent{}, false
	}
	return content, true
}
func validateUserContent(content model.UserMessageContent, contentType string) bool {
	switch contentType {
	case constant.MessageTypeText:
		if content.Text == nil {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->文本消息内容为空", "content", content)
			return false
		}
		if *content.Text == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->文本消息内容为空", "content", content)
			return false
		}
	case constant.MessageTypeImage:
		if content.Image == nil {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->图片消息内容为空", "content", content)
			return false
		}
		if content.Image.URL == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->图片URL为空", "content", content)
			return false
		}
		if content.Image.Name == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->图片名称不能为空", "content", content)
			return false
		}
		if content.Image.Format == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->图片格式不能为空", "content", content)
			return false
		}
		if content.Image.Size == 0 {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->图片大小不能为0", "content", content)
			return false
		}
	case constant.MessageTypeFile:
		if content.File == nil {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->文件消息内容为空", "content", content)
			return false
		}
		if content.File.URL == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->文件URL为空", "content", content)
			return false
		}
		if content.File.Name == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->文件名称不能为空", "content", content)
			return false
		}
		if content.File.Format == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->文件格式不能为空", "content", content)
			return false
		}
		if content.File.Size == 0 {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->文件大小不能为0", "content", content)
			return false
		}
	case constant.MessageTypeVoice:
		if content.Voice == nil {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->语音消息内容为空", "content", content)
			return false
		}
		if content.Voice.URL == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->语音URL为空", "content", content)
			return false
		}
		if content.Voice.Name == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->语音名称不能为空", "content", content)
			return false
		}
		if content.Voice.Format == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->语音格式不能为空", "content", content)
			return false
		}
		if content.Voice.Size == 0 {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->语音大小不能为0", "content", content)
			return false
		}
		if content.Voice.Duration == 0 {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->语音时长不能为0", "content", content)
			return false
		}
	case constant.MessageTypeVideo:
		if content.Video == nil {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->视频消息内容为空", "content", content)
			return false
		}
		if content.Video.URL == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->视频URL为空", "content", content)
			return false
		}
		if content.Video.Name == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->视频名称不能为空", "content", content)
			return false
		}
		if content.Video.Format == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->视频格式不能为空", "content", content)
			return false
		}
		if content.Video.Size == 0 {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->视频大小不能为0", "content", content)
			return false
		}
		if content.Video.Duration == 0 {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->视频时长不能为0", "content", content)
			return false
		}
	case constant.MessageTypeReply:
		if content.Reply == nil {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->回复消息内容为空", "content", content)
			return false
		}
		if content.Reply.Text == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->回复消息内容为空", "content", content)
			return false
		}
		if content.Reply.ReplyTo == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->回复消息内容为空", "content", content)
			return false
		}
	}
	return true
}

func validateSystemMessage(message *WebSocketMessage) (model.SystemMessageContent, bool) {
	// 验证id
	if message.RoomId == "" {
		global.CHAT_LOG.Error("WebSocket validateSystemMessage----->消息id不能为空", "message", message)
		return model.SystemMessageContent{}, false
	}
	if message.SenderId == "" {
		global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息发送者id不能为空", "message", message)
		return model.SystemMessageContent{}, false
	}
	if message.CreatedAt == 0 {
		global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息创建时间不能为空", "message", message)
		return model.SystemMessageContent{}, false
	}
	if message.Type == "" {
		global.CHAT_LOG.Error("WebSocket validateUserMessage----->消息类型不能为空", "message", message)
		return model.SystemMessageContent{}, false
	}
	// 验证消息内容
	content := model.SystemMessageContent{}
	// 验证系统消息
	switch message.Type {
	case constant.MessageTypeJoin:
		if contentMap, ok := message.Content.(map[string]interface{}); ok {
			join := utils.GetStringValue(contentMap, "join")
			content.Join = &join
			if !validateSystemContent(content, message.Type) {
				global.CHAT_LOG.Error("WebSocket validateSystemMessage----->消息内容验证失败", "message", message)
				return model.SystemMessageContent{}, false
			}
		} else {
			global.CHAT_LOG.Error("WebSocket validateSystemMessage----->消息内容验证失败", "message", message)
			return model.SystemMessageContent{}, false
		}
	case constant.MessageTypeLeave:
		if contentMap, ok := message.Content.(map[string]interface{}); ok {
			leave := utils.GetStringValue(contentMap, "leave")
			content.Leave = &leave
			if !validateSystemContent(content, message.Type) {
				global.CHAT_LOG.Error("WebSocket validateSystemMessage----->消息内容验证失败", "message", message)
				return model.SystemMessageContent{}, false
			}
		} else {
			global.CHAT_LOG.Error("WebSocket validateSystemMessage----->消息内容验证失败", "message", message)
			return model.SystemMessageContent{}, false
		}
	case constant.MessageTypeSystem:
		if contentMap, ok := message.Content.(map[string]interface{}); ok {
			system := utils.GetStringValue(contentMap, "system")
			content.System = &system
			if !validateSystemContent(content, message.Type) {
				global.CHAT_LOG.Error("WebSocket validateSystemMessage----->消息内容验证失败", "message", message)
				return model.SystemMessageContent{}, false
			}
		} else {
			global.CHAT_LOG.Error("WebSocket validateSystemMessage----->消息内容验证失败", "message", message)
			return model.SystemMessageContent{}, false
		}
	default:
		return model.SystemMessageContent{}, false
	}
	return content, true
}
func validateSystemContent(content model.SystemMessageContent, contentType string) bool {
	switch contentType {
	case constant.MessageTypeJoin:
		if content.Join == nil {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->加入消息内容为空", "content", content)
			return false
		}
		if *content.Join == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->加入消息内容为空", "content", content)
			return false
		}
	case constant.MessageTypeLeave:
		if content.Leave == nil {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->离开消息内容为空", "content", content)
			return false
		}
		if *content.Leave == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->离开消息内容为空", "content", content)
			return false
		}
	case constant.MessageTypeSystem:
		if content.System == nil {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->系统消息内容为空", "content", content)
			return false
		}
		if *content.System == "" {
			global.CHAT_LOG.Error("WebSocket validateUserContent----->系统消息内容为空", "content", content)
			return false
		}
	}
	return true
}
