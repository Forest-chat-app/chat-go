package service

import (
	"chat-server/global"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log/slog"

	"strings"
)

type MongoToEsSync struct{}

func (s *MongoToEsSync) StartMongoToEsSync(ctx context.Context, collectionName string, esIndex string) error {
	slog.Info(fmt.Sprintf("初始化 Mongo-Es 数据同步流程 (Collection: %s, Index: %s)", collectionName, esIndex))
	collection := global.CHAT_MONGODB.Collection(collectionName)
	pipeline := mongo.Pipeline{}

	//启动监听
	watch, err := collection.Watch(ctx, pipeline)
	if err != nil {
		slog.Error("监听 Change Stream 失败: ", "err", err)
		return err
	}
	defer func() {
		if err := watch.Close(ctx); err != nil {
			slog.Error(fmt.Sprintf("关闭 MongoDB Change Stream (Collection: %s) 失败:", collectionName), "err", err)
		} else {
			slog.Info(fmt.Sprintf("关闭 MongoDB Change Stream (Collection: %s) 成功。", collectionName))
		}
	}()
	slog.Info(fmt.Sprintf("开始监听 MongoDB 集合 %s 变更并同步到 ES 索引 %s...", collectionName, esIndex))

	//开始监听
	for watch.Next(ctx) {
		//解码文档 bson.M
		var event bson.M
		if err := watch.Decode(&event); err != nil {
			slog.Error(fmt.Sprintf("解码 Change Stream 事件失败 (Collection: %s):", collectionName), "err", err)
			continue
		}
		operationType, ok := event["operationType"].(string)
		if !ok {
			slog.Error(fmt.Sprintf("无法获取 operationType 或其类型不是 string (Collection: %s)", collectionName), "event", event)
			continue
		}
		operationType = strings.TrimSpace(operationType)

		switch operationType {
		case "insert", "replace", "update":
			//获取文档
			document, ok := event["fullDocument"].(bson.M)
			if !ok {
				// 1. 将 fullDocument (bson.D) 编码回 BSON 字节
				docBytes, err := bson.Marshal(event["fullDocument"])
				if err != nil {
					slog.Error(fmt.Sprintf("fullDocument 编码为 BSON 字节失败 (Collection: %s):", collectionName), "err", err, "fullDocument_type", fmt.Sprintf("%T", event["fullDocument"]))
					continue
				}

				// 2. 将 BSON 字节解码到 bson.M
				var docMap bson.M
				err = bson.Unmarshal(docBytes, &docMap)
				if err != nil {
					slog.Error(fmt.Sprintf("BSON 字节解码为 bson.M 失败 (Collection: %s):", collectionName), "err", err)
					continue
				}
				document = docMap
			}

			//先获取id，再把文档里的_id删掉，不然写入es的时候就会多一个_id
			var id string
			if mongoID, ok := document["_id"].(bson.ObjectID); ok {
				id = mongoID.Hex() // 使用 Hex() 方法获取十六进制字符串
			} else {
				id = fmt.Sprintf("%v", document["_id"])
				slog.Warn(fmt.Sprintf("警告: 文档 _id 类型不是 bson.ObjectID,使用通用字符串格式化 (Collection: %s):", collectionName), "id", id)
			}
			delete(document, "_id")

			//把文档转换为json，为了写入es
			jsonDocument, err := json.Marshal(document)
			if err != nil {
				slog.Error(fmt.Sprintf("document 到 JSON 失败 (Collection: %s):", collectionName), "err", err)
				continue
			}
			if len(jsonDocument) == 0 {
				slog.Error(fmt.Sprintf("JSON 结果为空，跳过写入ES (Collection: %s)。原始 document:", collectionName), "document", document)
				continue
			}

			//写入es
			res, err := global.CHAT_ES.Index(
				esIndex,
				strings.NewReader(string(jsonDocument)),
				global.CHAT_ES.Index.WithDocumentID(id),
				global.CHAT_ES.Index.WithContext(ctx),
			)
			if err != nil {
				slog.Error(fmt.Sprintf("写入 ES 索引 %s 失败 (Collection: %s):", esIndex, collectionName), "err", err)
			} else {
				if res.IsError() {
					slog.Error(fmt.Sprintf("ES 索引 %s 文档失败 (Collection: %s):", esIndex, collectionName), "文档", res.Status(), "文档ID:", id, "响应:", res.String())
				} else {
					slog.Info(fmt.Sprintf("ES 索引 %s 文档同步成功 (Collection: %s):", esIndex, collectionName), "文档", res.Status(), "文档ID:", id)
				}

				if closeErr := res.Body.Close(); closeErr != nil {
					slog.Error(fmt.Sprintf("关闭 ES Index 响应体失败 (Collection: %s, ID: %s):", collectionName, id), "err", closeErr)
				}
			}

		case "delete":
			//解码获取文档
			documentKey, ok := event["documentKey"].(bson.M)
			if !ok {
				// 1. 将 documentKey (bson.D) 编码回 BSON 字节
				docBytes, err := bson.Marshal(event["documentKey"])
				if err != nil {
					slog.Error(fmt.Sprintf("documentKey 编码为 BSON 字节失败 (Collection: %s):", collectionName), "err", err, "fullDocument_type", fmt.Sprintf("%T", event["fullDocument"]))
					continue
				}

				// 2. 将 BSON 字节解码到 bson.M
				var docMap bson.M
				err = bson.Unmarshal(docBytes, &docMap)
				if err != nil {
					slog.Error(fmt.Sprintf("BSON 字节解码为 bson.M 失败 (Collection: %s):", collectionName), "err", err)
					continue
				}
				documentKey = docMap
			}
			//获取16进制id
			var id string
			if mongoID, ok := documentKey["_id"].(bson.ObjectID); ok {
				id = mongoID.Hex() // 使用 Hex() 方法获取十六进制字符串
			} else {
				id = fmt.Sprintf("%v", documentKey["_id"])
				slog.Warn(fmt.Sprintf("警告: 文档 _id 类型不是 bson.ObjectID,使用通用字符串格式化 (Collection: %s):", collectionName), "id", id)
			}
			//删除文档
			res, err := global.CHAT_ES.Delete(esIndex, id, global.CHAT_ES.Delete.WithContext(ctx))
			if err != nil {
				slog.Error(fmt.Sprintf("删除 ES 索引 %s 文档失败 (Collection: %s):", esIndex, collectionName), "err", err)
			} else {
				if res.IsError() {
					slog.Error(fmt.Sprintf("ES 索引 %s 删除文档失败 (Collection: %s):", esIndex, collectionName), "文档", res.Status(), "文档ID:", id, "响应:", res.String())
				} else {
					slog.Info(fmt.Sprintf("ES 索引 %s 文档同步成功 (删除) (Collection: %s):", esIndex, collectionName), "文档", res.Status(), "文档ID:", id)
				}
				if closeErr := res.Body.Close(); closeErr != nil {
					slog.Error(fmt.Sprintf("关闭 ES Delete 响应体失败 (Collection: %s, ID: %s):", collectionName, id), "err", closeErr)
				}
			}

		default:
			slog.Error(fmt.Sprintf("未匹配到任何支持的操作类型 (Collection: %s):", collectionName), "operationType", operationType, "event", event)
		}

		if watch.Err() != nil && !errors.Is(watch.Err(), context.Canceled) {
			return fmt.Errorf("change Stream 监听 (Collection: %s) 意外终止: %w", collectionName, watch.Err())
		}
	}
	slog.Info(fmt.Sprintf("MongoDB Change Stream (Collection: %s) 监听结束。", collectionName))
	return nil
}
