package store

import (
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

func MinioInit(c *gin.Context, err error) (string, *minio.Client, bool) {
	endpoint, accessKeyID, secretAccessKey, useSSL := MinioProfile()
	if endpoint == "" || accessKeyID == "" || secretAccessKey == "" {
		log.Println("MinIO client configuration not set")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MinIO client configuration not set"})
		return "", nil, true
	}

	log.Println("MinIO client configuration set")

	// 初始化MinIO客户端
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Printf("Error initializing MinIO client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize MinIO client"})
		return "", nil, true
	}
	log.Println("MinIO client initialized successfully")
	return endpoint, minioClient, false
}

func MinioProfile() (string, string, string, bool) {
	// MinIO配置
	endpoint := viper.GetString("minio.endpoint")               // 替换为你的MinIO服务器地址
	accessKeyID := viper.GetString("minio.accessKeyID")         // 替换为你的Access Key
	secretAccessKey := viper.GetString("minio.secretAccessKey") // 替换为你的Secret Key
	useSSL := viper.GetBool("minio.useSSL")                     // 如果使用https，设置为true
	return endpoint, accessKeyID, secretAccessKey, useSSL
}

func MinioDownloadFile(c *gin.Context, client *minio.Client, bucketName string, objectName string, filePath string) error {
	// 从MinIO存储桶中下载文件
	err := client.FGetObject(c, bucketName, objectName, filePath, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("Error downloading file from MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file from MinIO"})
		return err
	}
	log.Println("File downloaded from MinIO successfully")
	return nil
}
