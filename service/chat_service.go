package service

import (
	"bytes"
	"chat-server/constant"
	"chat-server/global"
	"chat-server/model"
	"chat-server/model/common"
	"chat-server/model/request/chat"
	"chat-server/utils"
	"context"
	"encoding/json"
	"io"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ChatService struct{}

// GetHistoryMsg 根据页码获取历史消息
func (chatService *ChatService) GetHistoryMsg(req chat.HistoryMsgRequest) (chat.HistoryMsgResponse, error) {
	// 1、构建 Elasticsearch 查询
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

	// 1.1 将查询转换为 JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		global.CHAT_LOG.Error("构建 ES 查询失败", "err", err)
		return chat.HistoryMsgResponse{}, common.NewServiceError(common.ERROR)
	}

	// 1.2 执行 Elasticsearch 搜索
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

	// 2、解析响应
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

	// 2.2 将 ES 结果转换为 UserMessages 结构
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
		if msg.Type != constant.MessageTypeText {
			GenerateUrl(&msg)
		}

		messages = append(messages, msg)
	}

	// 3、返回结果
	return chat.HistoryMsgResponse{
		Messages: messages,
	}, nil
}

// SearchChat 搜索所有房间消息
func (chatService *ChatService) SearchChat(req chat.SearchChatRequest, userId string) (chat.SearchChatResponse, error) {
	if req.Type == 1 {
		// 类型1：搜索房间名称
		var roomIds []string
		err := global.CHAT_MYSQL.Table("room_members").
			Select("room.id").
			Joins("JOIN room ON room_members.room_id = room.id").
			Where("room_members.user_id = ? AND room.room_name LIKE ?", userId, "%"+req.Content+"%").
			Pluck("room.id", &roomIds).Error

		if err != nil {
			global.CHAT_LOG.Error("搜索房间名称失败", "err", err)
			return chat.SearchChatResponse{}, common.NewServiceError(common.ERROR)
		}

		return chat.SearchChatResponse{
			RoomIds: roomIds,
		}, nil
	} else if req.Type == 2 {
		// 类型2：搜索消息内容
		// 先获取用户所有的room_id
		var roomIds []string
		err := global.CHAT_MYSQL.Table("room_members").
			Select("room_id").
			Where("user_id = ?", userId).
			Pluck("room_id", &roomIds).Error

		if err != nil {
			global.CHAT_LOG.Error("获取用户房间列表失败", "err", err)
			return chat.SearchChatResponse{}, common.NewServiceError(common.ERROR)
		}

		if len(roomIds) == 0 {
			return chat.SearchChatResponse{
				Messages: []model.UserMessages{},
			}, nil
		}

		// 构建 Elasticsearch 查询，使用wildcard实现精准的子串匹配
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []map[string]interface{}{
						{
							"term": map[string]interface{}{
								"type": "text",
							},
						},
						{
							"wildcard": map[string]interface{}{
								"content.text.keyword": map[string]interface{}{
									"value":            "*" + req.Content + "*",
									"case_insensitive": true,
								},
							},
						},
						{
							"terms": map[string]interface{}{
								"room_id": roomIds,
							},
						},
					},
				},
			},
			"sort": []map[string]interface{}{
				{
					"created_at": map[string]interface{}{
						"order": "desc",
					},
				},
			},
			"size": 100, // 限制返回结果数量
		}

		// 将查询转换为 JSON
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			global.CHAT_LOG.Error("构建 ES 查询失败", "err", err)
			return chat.SearchChatResponse{}, common.NewServiceError(common.ERROR)
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
			return chat.SearchChatResponse{}, common.NewServiceError(common.ERROR)
		}
		defer res.Body.Close()

		if res.IsError() {
			bodyBytes, _ := io.ReadAll(res.Body)
			global.CHAT_LOG.Error("ES 搜索返回错误", "status", res.Status(), "response", string(bodyBytes))
			return chat.SearchChatResponse{}, common.NewServiceError(common.ERROR)
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
			return chat.SearchChatResponse{}, common.NewServiceError(common.ERROR)
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
			if msg.Type != constant.MessageTypeText {
				GenerateUrl(&msg)
			}

			messages = append(messages, msg)
		}

		return chat.SearchChatResponse{
			Messages: messages,
		}, nil
	}

	return chat.SearchChatResponse{}, common.NewServiceError(common.INVALID_PARAMS)
}

// GetRoomMsg 搜索某个房间的消息
func (chatService *ChatService) GetRoomMsg(req chat.GetRoomMsgReq) (chat.GetRoomMsgRsp, error) {
	// 构建 Elasticsearch 查询，使用wildcard实现精准的子串匹配
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"type": "text",
						},
					},
					{
						"wildcard": map[string]interface{}{
							"content.text.keyword": map[string]interface{}{
								"value":            "*" + req.Content + "*",
								"case_insensitive": true,
							},
						},
					},
					{
						"term": map[string]interface{}{
							"room_id": req.RoomId,
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{
				"created_at": map[string]interface{}{
					"order": "desc",
				},
			},
		},
		"size": 100, // 限制返回结果数量
	}

	// 将查询转换为 JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		global.CHAT_LOG.Error("构建 ES 查询失败", "err", err)
		return chat.GetRoomMsgRsp{}, common.NewServiceError(common.ERROR)
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
		return chat.GetRoomMsgRsp{}, common.NewServiceError(common.ERROR)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		global.CHAT_LOG.Error("ES 搜索返回错误", "status", res.Status(), "response", string(bodyBytes))
		return chat.GetRoomMsgRsp{}, common.NewServiceError(common.ERROR)
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
		return chat.GetRoomMsgRsp{}, common.NewServiceError(common.ERROR)
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
		if msg.Type != constant.MessageTypeText {
			GenerateUrl(&msg)
		}

		messages = append(messages, msg)
	}

	return chat.GetRoomMsgRsp{
		Messages: messages,
	}, nil
}

func GenerateUrl(msg *model.UserMessages) {
	switch msg.Type {
	case constant.MessageTypeImage:
		msg.Content.Image.URL = utils.GenerateCdnUrl(msg.Content.Image.URL)
	case constant.MessageTypeVoice:
		msg.Content.Voice.URL = utils.GenerateCdnUrl(msg.Content.Voice.URL)
	case constant.MessageTypeVideo:
		msg.Content.Video.URL = utils.GenerateCdnUrl(msg.Content.Video.URL)
	case constant.MessageTypeFile:
		msg.Content.File.URL = utils.GenerateCdnUrl(msg.Content.File.URL)
	}
}
