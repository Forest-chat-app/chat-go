package service

import (
	"bytes"
	"chat-server/global"
	"chat-server/model"
	"chat-server/model/common"
	"chat-server/model/request/chat"
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/v2/bson"
	"io"
)

type ChatService struct{}

func (chatService *ChatService) GetHistoryMsg(req chat.HistoryMsgRequest) (chat.HistoryMsgResponse, error) {
	// 构建 Elasticsearch 查询
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"room_id": req.RoomId,
			},
		},
		"sort": []map[string]interface{}{
			{
				"created_at": map[string]interface{}{
					"order": "desc", // 按时间倒序
				},
			},
		},
		"from": (req.PageNum - 1) * req.PageSize, // 分页起始位置
		"size": req.PageSize,                     // 每页大小
	}

	// 将查询转换为 JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		global.CHAT_LOG.Error("构建 ES 查询失败", "err", err)
		return chat.HistoryMsgResponse{}, common.NewServiceError(common.ERROR)
	}

	// 执行 Elasticsearch 搜索
	res, err := global.CHAT_ES.Search(
		global.CHAT_ES.Search.WithContext(context.Background()),
		global.CHAT_ES.Search.WithIndex("user_messages"),
		global.CHAT_ES.Search.WithBody(&buf),
		global.CHAT_ES.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		global.CHAT_LOG.Error("ES 搜索失败", "err", err)
		return chat.HistoryMsgResponse{}, common.NewServiceError(common.ERROR)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		global.CHAT_LOG.Error("ES 搜索返回错误", "status", res.Status(), "response", string(bodyBytes))
		return chat.HistoryMsgResponse{}, common.NewServiceError(common.ERROR)
	}

	// 解析响应
	var esResponse struct {
		Hits struct {
			Hits []struct {
				ID     string                 `json:"_id"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		global.CHAT_LOG.Error("解析 ES 响应失败", "err", err)
		return chat.HistoryMsgResponse{}, common.NewServiceError(common.ERROR)
	}

	// 将 ES 结果转换为 UserMessages 结构
	messages := make([]model.UserMessages, 0, len(esResponse.Hits.Hits))
	for _, hit := range esResponse.Hits.Hits {
		// 将 _id 添加回 source
		hit.Source["_id"] = hit.ID

		// 将 map 转换为 BSON 再转换为 UserMessages
		bsonBytes, err := bson.Marshal(hit.Source)
		if err != nil {
			global.CHAT_LOG.Error("转换为 BSON 失败", "err", err, "id", hit.ID)
			continue
		}

		var msg model.UserMessages
		if err := bson.Unmarshal(bsonBytes, &msg); err != nil {
			global.CHAT_LOG.Error("解析消息失败", "err", err, "id", hit.ID)
			continue
		}

		messages = append(messages, msg)
	}

	return chat.HistoryMsgResponse{
		Messages: messages,
	}, nil
}
