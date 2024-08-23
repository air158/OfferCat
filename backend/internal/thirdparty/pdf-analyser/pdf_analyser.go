package pdf_analyser

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"offercat/v0/internal/store"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-tika/tika"
)

// // 想要使用把package改成main
//
//	func main() {
//		// 在这里明文指定PDF文件路径
//		filePath := "S:\\11953\\Desk\\李鹤鹏-简历.pdf"
//
//		res := getStringFromPDF(filePath)
//		fmt.Println(res)
//	}
func GetStringFromPDF(c *gin.Context, filePath string) string {
	var err error
	_, client, b := store.MinioInit(c, err)
	if b {
		return "something wrong"
	}

	babies := strings.Split(filePath, "/")
	bucketName := babies[1]
	objectName := fmt.Sprintf("%s/%s", babies[2], babies[3])
	filePath = babies[3] // 保存原始文件名

	// 下载文件到 tmp 目录
	tmpFilePath := fmt.Sprintf("tmp/%s", filePath)
	err = store.MinioDownloadFile(c, client, bucketName, objectName, filePath)
	if err != nil {
		log.Printf("Error downloading file from MinIO: %v", err)
		return "something wrong"
	}

	// 检查下载的文件
	fileInfo, err := os.Stat(tmpFilePath) // 使用下载后的文件路径
	if err != nil || fileInfo.Size() == 0 {
		log.Printf("File download failed or file is empty: %v", err)
		return "something wrong"
	}
	time.Sleep(3 * time.Second)

	// 传递正确的路径给 getStringFromPDF
	res := getStringFromPDF(tmpFilePath)
	return res
}

// GetStringFromPDF 从PDF文件中提取文本内容
func getStringFromPDF(filePath string) string {
	content, err := readPdf(filePath) // 读取本地PDF文件
	if err != nil {
		panic(err)
	}

	text, err := extractTextFromHTML(content)
	if err != nil {
		panic(err)
	}

	res := compactText(text)
	return res
}

// TODO:记录一下这里debug的痛苦经历和经验，发一篇博客贴
func readPdf(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("无法打开文件: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Printf("无法关闭文件: %v\n", err)
		}
	}(f)
	// 自定义HTTP客户端，设置超时时间
	httpClient := &http.Client{
		Timeout: 10 * time.Minute, // 设置为10分钟，你可以根据需要调整
	}
	client := tika.NewClient(httpClient, "http://117.72.35.68:9998/")
	//client := tika.NewClient(httpClient, "http://116.198.207.159:9998")
	//content, err := client.Parse(contextTODO(), f)
	// 创建带超时的context
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()

	content, err := client.Parse(ctx, f)
	if err != nil {
		return "", fmt.Errorf("无法解析PDF文件: %v", err)
	}

	return content, nil
}

func extractTextFromHTML(htmlContent string) (string, error) {
	// 使用 goquery 解析 HTML 内容
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("无法解析HTML内容: %v", err)
	}

	// 提取文本
	text := doc.Text()
	return text, nil
}

func compactText(text string) string {
	// 移除多余的空白字符（包括空行）
	lines := strings.Split(text, "\n")
	var compactLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			compactLines = append(compactLines, trimmed)
		}
	}

	// 合并行，使用单个空格分隔段落
	compactText := strings.Join(compactLines, " ")
	return compactText
}
