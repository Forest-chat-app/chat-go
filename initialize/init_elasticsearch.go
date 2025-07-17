package initialize

import (
	"chat-server/global"
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/elastic/go-elasticsearch/v9"
	"net/http"
	"time"
)

func InitElasticSearch() error {
	global.CHAT_LOG.Info("初始化elasticsearch")
	EsConfig := global.CHAT_CONFIG.ElasticSearch

	// 加载CA证书
	caCert, err := os.ReadFile(EsConfig.CaFile)
	if err != nil {
		global.CHAT_LOG.Error("InitElasticSearch-->无法读取CA证书", "err", err)
	}
	// 创建证书池并添加CA证书
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cfg := elasticsearch.Config{
		Addresses: []string{
			EsConfig.Address,
		},
		Username: EsConfig.Username,
		Password: EsConfig.Password,
		Transport: &http.Transport{
			MaxIdleConns:          100,              // 保持的最大空闲连接数
			IdleConnTimeout:       90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,

			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: EsConfig.InsecureSkipVerify,
				RootCAs:            caCertPool,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // 15秒超时
	defer cancel()

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		global.CHAT_LOG.Error("Elasticsearch 连接失败: ", "err", err)
		return err
	}

	res, err := client.Info(client.Info.WithContext(ctx))
	if err != nil {
		global.CHAT_LOG.Error("Elasticsearch Ping 失败: ", "err", err)
		closeErr := res.Body.Close()
		if closeErr != nil {
			global.CHAT_LOG.Error("Elasticsearch Ping 失败后，，响应数据流关闭失败: ", "closeErr", closeErr)
		}
		return err
	}

	if res.IsError() {
		global.CHAT_LOG.Error("Elasticsearch 返回错误: ", "err", err)
		return err
	}
	global.CHAT_ES = client
	global.CHAT_LOG.Info("Elasticsearch连接成功")
	return nil
}
