package sitereader

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// SiteReader 定义站点读取器接口
type SiteReader interface {
	// 检测给定目录是否为此类型的站点
	Detect(dir string) bool

	// 读取站点结构，返回按顺序排列的文件列表
	// navPath 参数用于指定要导出的导航路径，为空时导出全部
	ReadStructure(dir string, configPath string, navPath string) ([]string, error)
}

// GetSiteReader 根据站点类型返回相应的读取器
func GetSiteReader(siteType string, verbose bool, logger *log.Logger) (SiteReader, error) {
	// 如果没有提供日志记录器，创建一个默认的
	if logger == nil {
		if verbose {
			logger = log.New(os.Stdout, "[SITE-READER] ", log.LstdFlags)
		} else {
			logger = log.New(io.Discard, "", 0)
		}
	}

	logger.Printf("Creating site reader for type: %s", siteType)

	switch siteType {
	case "mkdocs":
		logger.Println("Using MkDocs site reader")
		return &MkDocsReader{Logger: logger}, nil
	case "hugo":
		logger.Println("Hugo site type is not yet implemented")
		return nil, fmt.Errorf("hugo site type is not yet implemented")
	case "docusaurus":
		logger.Println("Docusaurus site type is not yet implemented")
		return nil, fmt.Errorf("docusaurus site type is not yet implemented")
	default:
		logger.Printf("Unsupported site type: %s", siteType)
		return nil, fmt.Errorf("unsupported site type: %s", siteType)
	}
}

// FindConfigFile 在给定目录中查找配置文件
func FindConfigFile(dir string, configNames []string) (string, error) {
	// 如果没有提供配置文件名，使用默认值
	if len(configNames) == 0 {
		configNames = []string{"config.yml", "config.yaml"}
	}

	// 查找配置文件
	for _, name := range configNames {
		configPath := filepath.Join(dir, name)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("no config file found in %s", dir)
}
