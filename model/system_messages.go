package model

import "go.mongodb.org/mongo-driver/v2/bson"

// SystemMessages 消息模型
type SystemMessages struct {
	ID        bson.ObjectID        `bson:"_id" json:"_id"`               // 同时支持BSON和JSON
	RoomId    string               `bson:"room_id" json:"room_id"`       // 同时支持BSON和JSON
	SenderId  string               `bson:"sender_id" json:"sender_id"`   // 同时支持BSON和JSON
	Type      string               `bson:"type" json:"type"`             // 同时支持BSON和JSON
	Content   SystemMessageContent `bson:"content" json:"content"`       // 同时支持BSON和JSON
	CreatedAt int64                `bson:"created_at" json:"created_at"` // 同时支持BSON和JSON
}

// 内容模型
type SystemMessageContent struct {
	Join   *string `bson:"join" json:"join"`
	Leave  *string `bson:"leave" json:"leave"`
	System *string `bson:"system" json:"system"`
}
