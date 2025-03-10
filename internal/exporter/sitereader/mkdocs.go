package sitereader

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// MkDocsReader 实现 MkDocs 站点的读取
type MkDocsReader struct {
	Logger *log.Logger
}

// MkDocsConfig 表示 MkDocs 配置文件结构
type MkDocsConfig struct {
	Docs    []string `yaml:"nav"`
	DocsDir string   `yaml:"docs_dir"`
	Inherit string   `yaml:"INHERIT"`
}

// Detect 检测给定目录是否为 MkDocs 站点
func (r *MkDocsReader) Detect(dir string) bool {
	// 设置日志记录器
	if r.Logger == nil {
		r.Logger = log.New(io.Discard, "", 0)
	}

	// 检查是否存在 mkdocs.yml 文件
	mkdocsPath := filepath.Join(dir, "mkdocs.yml")
	if _, err := os.Stat(mkdocsPath); os.IsNotExist(err) {
		// 尝试 mkdocs.yaml
		mkdocsPath = filepath.Join(dir, "mkdocs.yaml")
		if _, err := os.Stat(mkdocsPath); os.IsNotExist(err) {
			r.Logger.Printf("No mkdocs.yml or mkdocs.yaml found in %s", dir)
			return false
		}
	}

	r.Logger.Printf("Found MkDocs configuration file: %s", mkdocsPath)
	return true
}

// ReadStructure 读取 MkDocs 站点结构
func (r *MkDocsReader) ReadStructure(dir string, configPath string, navPath string) ([]string, error) {
	// 设置日志记录器
	if r.Logger == nil {
		r.Logger = log.New(io.Discard, "", 0)
	}

	r.Logger.Printf("Reading MkDocs site structure from: %s", dir)
	if navPath != "" {
		r.Logger.Printf("Filtering by navigation path: %s", navPath)
	}

	// 查找配置文件
	if configPath == "" {
		configNames := []string{"mkdocs.yml", "mkdocs.yaml"}
		var err error
		configPath, err = FindConfigFile(dir, configNames)
		if err != nil {
			r.Logger.Printf("Failed to find MkDocs config file: %s", err)
			return nil, fmt.Errorf("failed to find MkDocs config file: %s", err)
		}
	}
	r.Logger.Printf("Using config file: %s", configPath)

	// 读取并解析配置文件，包括处理 INHERIT
	config, err := r.readAndMergeConfig(configPath, dir)
	if err != nil {
		r.Logger.Printf("Failed to read config file: %s", err)
		return nil, fmt.Errorf("failed to read config file: %s", err)
	}

	// 获取文档目录
	docsDir := "docs"
	if docsDirValue, ok := config["docs_dir"]; ok {
		if docsDirStr, ok := docsDirValue.(string); ok {
			docsDir = docsDirStr
		}
	}
	docsDir = filepath.Join(dir, docsDir)
	r.Logger.Printf("Using docs directory: %s", docsDir)

	// 解析导航结构
	var nav interface{}
	if navValue, ok := config["nav"]; ok {
		nav = navValue
	} else {
		// 如果没有导航配置，尝试查找所有 Markdown 文件
		r.Logger.Println("No navigation configuration found, searching for all markdown files")
		return getAllMarkdownFiles(docsDir)
	}

	// 解析导航结构，获取文件列表
	files, err := parseNavigation(nav, docsDir, navPath)
	if err != nil {
		r.Logger.Printf("Failed to parse navigation: %s", err)
		return nil, fmt.Errorf("failed to parse navigation: %s", err)
	}

	r.Logger.Printf("Found %d files in navigation", len(files))
	return files, nil
}

// readAndMergeConfig 读取并合并 MkDocs 配置文件，处理 INHERIT 指令
func (r *MkDocsReader) readAndMergeConfig(configPath string, baseDir string) (map[string]interface{}, error) {
	r.Logger.Printf("Reading and merging config file: %s", configPath)

	// 读取主配置文件
	configData, err := os.ReadFile(configPath)
	if err != nil {
		r.Logger.Printf("Failed to read MkDocs config file: %s", err)
		return nil, fmt.Errorf("failed to read MkDocs config file: %s", err)
	}

	// 解析配置文件
	var config map[string]interface{}
	if err := yaml.Unmarshal(configData, &config); err != nil {
		r.Logger.Printf("Failed to parse MkDocs config file: %s", err)
		return nil, fmt.Errorf("failed to parse MkDocs config file: %s", err)
	}

	// 检查是否有 INHERIT 指令
	inheritValue, hasInherit := config["INHERIT"]
	if !hasInherit {
		// 没有继承，直接返回当前配置
		return config, nil
	}

	// 处理 INHERIT 指令
	inheritPath, ok := inheritValue.(string)
	if !ok {
		r.Logger.Printf("Invalid INHERIT value, expected string but got: %T", inheritValue)
		return nil, fmt.Errorf("invalid INHERIT value, expected string")
	}

	r.Logger.Printf("Found INHERIT directive pointing to: %s", inheritPath)

	// 解析继承路径，可能是相对于当前配置文件的路径
	configDir := filepath.Dir(configPath)
	inheritFullPath := filepath.Join(configDir, inheritPath)

	// 读取继承的配置文件
	inheritConfig, err := r.readAndMergeConfig(inheritFullPath, baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read inherited config file %s: %s", inheritFullPath, err)
	}

	// 合并配置，当前配置优先
	mergedConfig := make(map[string]interface{})

	// 先复制继承的配置
	for k, v := range inheritConfig {
		mergedConfig[k] = v
	}

	// 再覆盖当前配置
	for k, v := range config {
		if k != "INHERIT" { // 不复制 INHERIT 指令
			mergedConfig[k] = v
		}
	}

	r.Logger.Printf("Successfully merged config with inherited file")
	return mergedConfig, nil
}

// preprocessMarkdownFile 预处理 Markdown 文件，移除可能导致问题的 YAML front matter
func preprocessMarkdownFile(filePath string) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 检查是否有 YAML front matter
	contentStr := string(content)
	yamlFrontMatterRegex := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)

	// 如果有 YAML front matter，移除它
	if yamlFrontMatterRegex.MatchString(contentStr) {
		// 创建临时文件
		tempFile, err := os.CreateTemp("", "mdctl-*.md")
		if err != nil {
			return err
		}
		tempFilePath := tempFile.Name()
		tempFile.Close()

		// 移除 YAML front matter
		processedContent := yamlFrontMatterRegex.ReplaceAllString(contentStr, "")

		// 写入处理后的内容到临时文件
		if err := os.WriteFile(tempFilePath, []byte(processedContent), 0644); err != nil {
			os.Remove(tempFilePath)
			return err
		}

		// 替换原始文件
		if err := os.Rename(tempFilePath, filePath); err != nil {
			os.Remove(tempFilePath)
			return err
		}
	}

	return nil
}

// parseNavigation 解析 MkDocs 导航结构
func parseNavigation(nav interface{}, docsDir string, navPath string) ([]string, error) {
	var files []string

	switch v := nav.(type) {
	case []interface{}:
		// 导航是一个列表
		for _, item := range v {
			itemFiles, err := parseNavigation(item, docsDir, navPath)
			if err != nil {
				return nil, err
			}
			files = append(files, itemFiles...)
		}
	case map[string]interface{}:
		// 导航是一个映射
		for title, value := range v {
			// 如果指定了导航路径，检查当前节点标题是否匹配
			if navPath != "" {
				// 支持简单的路径匹配，例如 "Section1/Subsection2"
				navParts := strings.Split(navPath, "/")
				if strings.TrimSpace(title) == strings.TrimSpace(navParts[0]) {
					// 如果是多级路径，继续匹配下一级
					if len(navParts) > 1 {
						subNavPath := strings.Join(navParts[1:], "/")
						itemFiles, err := parseNavigation(value, docsDir, subNavPath)
						if err != nil {
							return nil, err
						}
						files = append(files, itemFiles...)
						continue
					} else {
						// 如果是单级路径且匹配，只处理这个节点
						itemFiles, err := parseNavigation(value, docsDir, "")
						if err != nil {
							return nil, err
						}
						files = append(files, itemFiles...)
						continue
					}
				} else {
					// 标题不匹配，跳过这个节点
					continue
				}
			}

			// 如果没有指定导航路径或已经匹配到了路径，正常处理
			itemFiles, err := parseNavigation(value, docsDir, "")
			if err != nil {
				return nil, err
			}
			files = append(files, itemFiles...)
		}
	case string:
		// 导航项是一个文件路径
		if strings.HasSuffix(v, ".md") {
			filePath := filepath.Join(docsDir, v)
			if _, err := os.Stat(filePath); err == nil {
				// 如果没有指定导航路径或已经在导航路径过滤中处理过，添加文件
				if navPath == "" {
					files = append(files, filePath)
				}
			}
		}
	}

	return files, nil
}

// getAllMarkdownFiles 获取目录中的所有 Markdown 文件
func getAllMarkdownFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".md" || ext == ".markdown" {
				files = append(files, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %s", dir, err)
	}

	return files, nil
}
