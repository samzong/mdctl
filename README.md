# mdctl

一个用于处理 Markdown 文件的命令行工具。目前支持自动下载远程图片到本地，并更新 Markdown 文件中的图片引用路径。

## 功能特点

- 支持处理单个 Markdown 文件或整个目录（包括子目录）
- 自动下载远程图片到本地指定目录
- 保持原有图片文件名（如果可用）
- 自动处理重名文件
- 使用相对路径更新 Markdown 文件中的图片引用
- 详细的处理日志输出

## 安装

```bash
go install github.com/samzong/mdctl@latest
```

## 使用方法

### 下载远程图片

处理单个文件：
```bash
mdctl download -f path/to/your/file.md
```

处理整个目录：
```bash
mdctl download -d path/to/your/directory
```

指定图片输出目录：
```bash
mdctl download -f path/to/your/file.md -o path/to/images
```

## 命令说明

### download 命令

下载并本地化 Markdown 文件中的远程图片。

参数：
- `-f, --file`: 指定要处理的 Markdown 文件
- `-d, --dir`: 指定要处理的目录（将递归处理所有 Markdown 文件）
- `-o, --output`: 指定图片保存的目录（可选）
  - 文件模式下默认保存在文件所在目录的 `images` 子目录中
  - 目录模式下默认保存在目录下的 `images` 子目录中

## 注意事项

1. 不能同时指定 `-f` 和 `-d` 参数
2. 如果不指定输出目录，工具会自动创建默认的 `images` 目录
3. 工具只会处理远程图片（http/https），本地图片引用不会被修改
4. 图片文件名会保持原有名称，如果发生重名，会自动添加 URL 的哈希值作为后缀 