package admin

import (
	"bytes"
	"chat-server/global"
	"chat-server/model"
	"chat-server/model/common"
	reqAdmin "chat-server/model/request/admin"
	"context"
	"encoding/json"
	"io"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AdminMessageService struct{}

// ListMessages 分页获取某房间消息
func (s *AdminMessageService) ListMessages(req reqAdmin.AdminListMessagesRequest) (map[string]interface{}, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"room_id": req.RoomId,
			},
		},
		"sort": []map[string]interface{}{
			{"created_at": map[string]interface{}{"order": "desc"}},
		},
		"from":             (req.Page - 1) * req.PageSize,
		"size":             req.PageSize,
		"track_total_hits": true,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, common.NewServiceError(common.ERROR)
	}

	res, err := global.CHAT_ES.Search(
		global.CHAT_ES.Search.WithContext(context.Background()),
		global.CHAT_ES.Search.WithIndex("user_messages"),
		global.CHAT_ES.Search.WithBody(&buf),
		global.CHAT_ES.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		global.CHAT_LOG.Error("AdminListMessages-->ES搜索失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		global.CHAT_LOG.Error("AdminListMessages-->ES返回错误", "status", res.Status(), "body", string(bodyBytes))
		return nil, common.NewServiceError(common.ERROR)
	}

	var esResp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string                 `json:"_id"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, common.NewServiceError(common.ERROR)
	}

	messages := make([]model.UserMessages, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		hit.Source["_id"] = hit.ID
		bsonBytes, err := bson.Marshal(hit.Source)
		if err != nil {
			continue
		}
		var msg model.UserMessages
		if err := bson.Unmarshal(bsonBytes, &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return map[string]interface{}{
		"total":     esResp.Hits.Total.Value,
		"page":      req.Page,
		"page_size": req.PageSize,
		"messages":  messages,
	}, nil
}

// SearchMessages 全局搜索消息（支持可选 room_id 过滤）
func (s *AdminMessageService) SearchMessages(req reqAdmin.AdminSearchMessagesRequest) (map[string]interface{}, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	mustClauses := []map[string]interface{}{
		{
			"wildcard": map[string]interface{}{
				"content.text.keyword": map[string]interface{}{
					"value":            "*" + req.Keyword + "*",
					"case_insensitive": true,
				},
			},
		},
	}
	if req.RoomId != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{"room_id": req.RoomId},
		})
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
		"sort": []map[string]interface{}{
			{"created_at": map[string]interface{}{"order": "desc"}},
		},
		"from":             (req.Page - 1) * req.PageSize,
		"size":             req.PageSize,
		"track_total_hits": true,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, common.NewServiceError(common.ERROR)
	}

	res, err := global.CHAT_ES.Search(
		global.CHAT_ES.Search.WithContext(context.Background()),
		global.CHAT_ES.Search.WithIndex("user_messages"),
		global.CHAT_ES.Search.WithBody(&buf),
		global.CHAT_ES.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, common.NewServiceError(common.ERROR)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		global.CHAT_LOG.Error("AdminSearchMessages-->ES返回错误", "status", res.Status(), "body", string(bodyBytes))
		return nil, common.NewServiceError(common.ERROR)
	}

	var esResp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string                 `json:"_id"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, common.NewServiceError(common.ERROR)
	}

	messages := make([]model.UserMessages, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		hit.Source["_id"] = hit.ID
		bsonBytes, err := bson.Marshal(hit.Source)
		if err != nil {
			continue
		}
		var msg model.UserMessages
		if err := bson.Unmarshal(bsonBytes, &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return map[string]interface{}{
		"total":     esResp.Hits.Total.Value,
		"page":      req.Page,
		"page_size": req.PageSize,
		"messages":  messages,
	}, nil
}

// DeleteMessage 删除消息（同时从 MongoDB 和 ES 删除）
func (s *AdminMessageService) DeleteMessage(req reqAdmin.AdminDeleteMessageRequest) error {
	ctx := context.Background()

	// 将字符串ID转为 ObjectID
	objID, err := bson.ObjectIDFromHex(req.MessageId)
	if err != nil {
		return common.NewServiceError(common.MESSAGE_NOT_FOUND)
	}

	// 1、从 MongoDB 删除
	col := global.CHAT_MONGODB.Collection("user_messages")
	result, err := col.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		global.CHAT_LOG.Error("AdminDeleteMessage-->MongoDB删除失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}
	if result.DeletedCount == 0 {
		return common.NewServiceError(common.MESSAGE_NOT_FOUND)
	}

	// 2、从 ES 删除
	esRes, err := global.CHAT_ES.Delete("user_messages", req.MessageId,
		global.CHAT_ES.Delete.WithContext(ctx),
	)
	if err != nil {
		global.CHAT_LOG.Error("AdminDeleteMessage-->ES删除失败", "err", err)
		// ES删除失败不影响主流程，仅记录日志
	}
	if esRes != nil {
		esRes.Body.Close()
	}

	return nil
}
