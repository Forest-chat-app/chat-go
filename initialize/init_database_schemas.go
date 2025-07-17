package initialize

import (
	"chat-server/global"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"chat-server/config"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// InitDatabaseSchemas Orchestrates schema initialization for all configured databases.
// 接收一个 config.DBSchemaConfig 实例作为参数
func InitDatabaseSchemas(ctx context.Context, schemaCfg config.DBSchemaConfig) error { // <-- 修改函数签名
	global.CHAT_LOG.Info("开始初始化数据库结构...")

	// MySQL Schema Initialization
	if schemaCfg.MySQL != nil { // <-- 使用传入的 schemaCfg
		if err := initMySQLSchemaFromFile(ctx, schemaCfg.MySQL.ScriptFile); err != nil { // <-- 使用传入的 schemaCfg
			return fmt.Errorf("初始化 MySQL 结构失败: %w", err)
		}
	}

	// MongoDB Schema Initialization
	if schemaCfg.MongoDB != nil { // <-- 使用传入的 schemaCfg
		if err := initMongoDBSchema(ctx, schemaCfg.MongoDB); err != nil { // <-- 使用传入的 schemaCfg
			return fmt.Errorf("初始化 MongoDB 结构失败: %w", err)
		}
	}

	// Elasticsearch Schema Initialization
	if schemaCfg.Elasticsearch != nil { // <-- 使用传入的 schemaCfg
		if err := initElasticsearchSchema(ctx, schemaCfg.Elasticsearch); err != nil { // <-- 使用传入的 schemaCfg
			return fmt.Errorf("初始化 Elasticsearch 结构失败: %w", err)
		}
	}

	global.CHAT_LOG.Info("数据库结构初始化完成。")
	return nil
}

// initMySQLSchemaFromFile 从SQL文件执行MySQL Schema初始化
func initMySQLSchemaFromFile(ctx context.Context, scriptPath string) error {
	if global.CHAT_MYSQL == nil { // 使用全局变量
		global.CHAT_LOG.Warn("MySQL 客户端未初始化，跳过 MySQL Schema引导。")
		return nil
	}
	if scriptPath == "" {
		global.CHAT_LOG.Info("MySQL Schema脚本路径未配置，跳过MySQL Schema初始化。")
		return nil
	}
	global.CHAT_LOG.Info(fmt.Sprintf("开始执行 MySQL Schema脚本: %s", scriptPath))

	sqlBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		global.CHAT_LOG.Error(fmt.Sprintf("读取 MySQL Schema文件 '%s' 失败: %w", scriptPath, err))
		return err
	}
	sqlContent := string(sqlBytes)

	sqls := strings.Split(sqlContent, ";")

	tx := global.CHAT_MYSQL.WithContext(ctx).Begin()
	if tx.Error != nil {
		global.CHAT_LOG.Error(fmt.Sprintf("开启MySQL事务失败: %w", tx.Error))
		return tx.Error
	}

	for _, sql := range sqls {
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}
		if strings.HasPrefix(sql, "--") || strings.HasPrefix(sql, "#") {
			continue
		}
		if err := tx.Exec(sql).Error; err != nil {
			tx.Rollback()
			global.CHAT_LOG.Error(fmt.Sprintf("执行 MySQL SQL语句失败: %w\nSQL: %s", err, sql))
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		global.CHAT_LOG.Error(fmt.Sprintf("提交MySQL事务失败: %w", err))
		return err
	}

	global.CHAT_LOG.Info("MySQL Schema脚本执行成功。")
	return nil
}

// initMongoDBSchema handles MongoDB collection schema initialization (indexes and validation).
// 接收一个 *config.MongoDBClusterSchemaConfig 实例作为参数
func initMongoDBSchema(ctx context.Context, cfg *config.MongoDBClusterSchemaConfig) error { // <-- 修改函数签名
	if global.CHAT_MONGODB == nil { // 使用全局变量
		return errors.New("MongoDB 客户端未初始化")
	}

	db := global.CHAT_MONGODB // 获取 MongoDB 数据库客户端

	for _, collectionCfg := range cfg.Collections { // <-- 使用传入的 cfg
		global.CHAT_LOG.Info(fmt.Sprintf("开始处理 MongoDB 集合 '%s' 的Schema...", collectionCfg.Name))
		collection := db.Collection(collectionCfg.Name)

		// --- 1. 处理索引文件 ---
		if collectionCfg.IndexFile != "" {
			global.CHAT_LOG.Info(fmt.Sprintf("检查集合 '%s' 的索引...", collectionCfg.Name))
			indexBytes, err := os.ReadFile(collectionCfg.IndexFile)
			if err != nil {
				return fmt.Errorf("读取 MongoDB 索引文件 '%s' 失败: %w", collectionCfg.IndexFile, err)
			}
			var indexDefs []config.MongoDBIndexSchema // <-- 使用 config.MongoDBIndexSchema
			if err := json.Unmarshal(indexBytes, &indexDefs); err != nil {
				return fmt.Errorf("解析 MongoDB 索引文件 '%s' (JSON) 失败: %w", collectionCfg.IndexFile, err)
			}

			var indexModels []mongo.IndexModel
			for _, indexCfg := range indexDefs {
				keysDoc := bson.D{} // 重新初始化一个有序的 BSON 文档
				if indexCfg.Options["name"] == "room_timestamp" {
					keysDoc = append(keysDoc, bson.E{Key: "room_id", Value: indexCfg.Keys["room_id"]})
					keysDoc = append(keysDoc, bson.E{Key: "timestamp", Value: indexCfg.Keys["timestamp"]})
				} else if indexCfg.Options["name"] == "_id_" {
					keysDoc = append(keysDoc, bson.E{Key: "_id", Value: indexCfg.Keys["_id"]})
				} else if indexCfg.Options["name"] == "timestamp_desc" {
					keysDoc = append(keysDoc, bson.E{Key: "timestamp", Value: indexCfg.Keys["timestamp"]})
				} else {
					// 对于其他通用索引，或者不关心顺序的单字段索引，可以继续遍历
					for k, v := range indexCfg.Keys {
						keysDoc = append(keysDoc, bson.E{Key: k, Value: v})
					}
				}

				opts := options.Index()
				for optKey, optVal := range indexCfg.Options {
					switch optKey {
					case "unique":
						if val, ok := optVal.(bool); ok {
							opts.SetUnique(val)
						}
					case "name":
						if val, ok := optVal.(string); ok {
							opts.SetName(val)
						}
					case "sparse":
						if val, ok := optVal.(bool); ok {
							opts.SetSparse(val)
						}
					case "expireAfterSeconds":
						if val, ok := optVal.(float64); ok {
							opts.SetExpireAfterSeconds(int32(val))
						}
					default:
						global.CHAT_LOG.Warn(fmt.Sprintf("未识别的 MongoDB 索引选项 '%s' for collection '%s'", optKey, collectionCfg.Name))
					}
				}
				indexModels = append(indexModels, mongo.IndexModel{Keys: keysDoc, Options: opts})
			}
			if len(indexModels) > 0 {
				_, err := collection.Indexes().CreateMany(ctx, indexModels)
				if err != nil {
					return fmt.Errorf("为集合 '%s' 创建索引失败: %w", collectionCfg.Name, err)
				}
				global.CHAT_LOG.Info(fmt.Sprintf("为集合 '%s' 成功创建/确认所有配置的索引。", collectionCfg.Name))
			}
		}

		// --- 2. 处理 Schema Validation 命令文件 ---
		if collectionCfg.ValidatorCommandFile != "" {
			global.CHAT_LOG.Info(fmt.Sprintf("应用集合 '%s' 的 Schema Validation 规则...", collectionCfg.Name))
			commandBytes, err := os.ReadFile(collectionCfg.ValidatorCommandFile)
			if err != nil {
				return fmt.Errorf("读取 MongoDB Schema Validation 文件 '%s' 失败: %w", collectionCfg.ValidatorCommandFile, err)
			}

			var commandDoc bson.D
			if err := bson.UnmarshalExtJSON(commandBytes, true, &commandDoc); err != nil {
				return fmt.Errorf("解析 MongoDB Schema Validation 命令文件 '%s' (JSON) 失败: %w", collectionCfg.ValidatorCommandFile, err)
			}

			// 动态更新 collMod 字段以匹配集合名称
			foundCollMod := false
			for i, elem := range commandDoc {
				if elem.Key == "collMod" {
					commandDoc[i].Value = collectionCfg.Name
					foundCollMod = true
					break
				}
			}
			if !foundCollMod {
				return fmt.Errorf("MongoDB Schema Validation 命令文件 '%s' 缺少 'collMod' 键", collectionCfg.ValidatorCommandFile)
			}

			var result bson.M
			err = db.RunCommand(ctx, commandDoc).Decode(&result)
			if err != nil {
				return fmt.Errorf("执行 MongoDB collMod 命令失败: %w", err)
			}

			if ok, found := result["ok"]; !found || ok.(float64) != 1.0 {
				return fmt.Errorf("执行 MongoDB collMod 命令返回错误: %v", result)
			}
			global.CHAT_LOG.Info(fmt.Sprintf("集合 '%s' 的 Schema Validation 规则应用成功。", collectionCfg.Name))
		}
	}
	return nil
}

// initElasticsearchSchema handles Elasticsearch index creation from a request file.
// 接收一个 *config.ElasticsearchClusterSchemaConfig 实例作为参数
func initElasticsearchSchema(ctx context.Context, cfg *config.ElasticsearchClusterSchemaConfig) error { // <-- 修改函数签名
	if global.CHAT_ES == nil { // 使用全局变量
		return errors.New("Elasticsearch 客户端未初始化")
	}

	for _, indexCfg := range cfg.Indices { // <-- 使用传入的 cfg
		global.CHAT_LOG.Info(fmt.Sprintf("检查 Elasticsearch 索引 '%s'...", indexCfg.Name))

		existsRes, err := global.CHAT_ES.Indices.Exists([]string{indexCfg.Name}, global.CHAT_ES.Indices.Exists.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("检查 ES 索引 '%s' 是否存在失败: %w", indexCfg.Name, err)
		}
		defer existsRes.Body.Close()

		if existsRes.IsError() {
			if existsRes.StatusCode == 404 {
				global.CHAT_LOG.Info(fmt.Sprintf("ES 索引 '%s' 不存在，开始创建...", indexCfg.Name))

				requestBytes, err := os.ReadFile(indexCfg.RequestFile)
				if err != nil {
					return fmt.Errorf("读取 ES 索引 '%s' 请求体文件 '%s' 失败: %w", indexCfg.Name, indexCfg.RequestFile, err)
				}

				createRes, createErr := global.CHAT_ES.Indices.Create(indexCfg.Name,
					global.CHAT_ES.Indices.Create.WithBody(strings.NewReader(string(requestBytes))),
					global.CHAT_ES.Indices.Create.WithContext(ctx),
				)
				if createErr != nil {
					return fmt.Errorf("创建 ES 索引 '%s' 失败: %w", indexCfg.Name, createErr)
				}
				defer createRes.Body.Close()
				if createRes.IsError() {
					return fmt.Errorf("创建 ES 索引 '%s' 失败 (ES 响应错误): %s", indexCfg.Name, createRes.String())
				}
				global.CHAT_LOG.Info(fmt.Sprintf("ES 索引 '%s' 创建成功。", indexCfg.Name))
			} else {
				return fmt.Errorf("检查 ES 索引 '%s' 存在性返回非 404 错误: %s", indexCfg.Name, existsRes.String())
			}
		} else {
			global.CHAT_LOG.Info(fmt.Sprintf("ES 索引 '%s' 已存在。", indexCfg.Name))
		}
	}
	return nil
}
