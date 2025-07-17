package model

import "go.mongodb.org/mongo-driver/v2/bson"

// Messages 消息模型
type UserMessages struct {
	ID        bson.ObjectID      `bson:"_id" json:"_id"`               // 同时支持BSON和JSON
	RoomId    string             `bson:"room_id" json:"room_id"`       // 同时支持BSON和JSON
	SenderId  string             `bson:"sender_id" json:"sender_id"`   // 同时支持BSON和JSON
	Type      string             `bson:"type" json:"type"`             // 同时支持BSON和JSON
	Content   UserMessageContent `bson:"content" json:"content"`       // 同时支持BSON和JSON
	CreatedAt int64              `bson:"created_at" json:"created_at"` // 同时支持BSON和JSON
}

// 内容模型
type UserMessageContent struct {
	Text  *string                  `bson:"text" json:"text"`
	Image *UserMessageContentImage `bson:"image" json:"image"`
	File  *UserMessageContentFile  `bson:"file" json:"file"`
	Voice *UserMessageContentVoice `bson:"voice" json:"voice"`
	Video *UserMessageContentVideo `bson:"video" json:"video"`
	Reply *UserMessageContentReply `bson:"reply" json:"reply"`
}

// 对应的不同内容模型
// size都是字节、duration都是秒
type UserMessageContentImage struct {
	URL    string `bson:"url" json:"url"`
	Name   string `bson:"name" json:"name"`
	Size   int    `bson:"size" json:"size"`
	Format string `bson:"format" json:"format"`
}
type UserMessageContentFile struct {
	URL    string `bson:"url" json:"url"`
	Name   string `bson:"name" json:"name"`
	Size   int    `bson:"size" json:"size"`
	Format string `bson:"format" json:"format"`
}
type UserMessageContentVoice struct {
	URL      string  `bson:"url" json:"url"`
	Name     string  `bson:"name" json:"name"`
	Size     int     `bson:"size" json:"size"`
	Format   string  `bson:"format" json:"format"`
	Duration float64 `bson:"duration" json:"duration"`
}
type UserMessageContentVideo struct {
	URL      string  `bson:"url" json:"url"`
	Name     string  `bson:"name" json:"name"`
	Size     int     `bson:"size" json:"size"`
	Format   string  `bson:"format" json:"format"`
	Duration float64 `bson:"duration" json:"duration"`
}

type UserMessageContentReply struct {
	Text    string `bson:"text" json:"text"`
	ReplyTo string `bson:"reply_to" json:"reply_to"`
}
