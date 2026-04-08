# Cloudinary MCP 项目

## 概述
本项目是一个基于Go语言开发的MCP（Multi - Cloud Processing）服务器，主要功能是将文件上传到Cloudinary云存储服务。服务器通过MCP协议接收文件路径请求，并将对应文件上传到Cloudinary，最后返回文件的安全访问链接。

## 环境变量
运行项目前，需要设置以下环境变量：
- `cloud`: Cloudinary的云名称。
- `key`: Cloudinary的API密钥。
- `secret`: Cloudinary的API密钥密码。

密钥从[这里](https://console.cloudinary.com/settings/api-keys)获取。

## 运行步骤
1. 确保Go环境已正确安装（版本1.23.1及以上）。
2. 设置所需的环境变量。
3. 在项目根目录下执行以下命令运行项目：
```bash
go install gitee.com/cyeam/cloudinary_mcp@latest

{
  "mcpServers": {
    "image_upload": {
      "type": "stdio",
      "command": "cloudinary",
      "args": [],
      "env": {
        "cloud": "cyeam",
        "key": "key",
        "secret": "password"
      }
    }
  }
}
```

## 测试代码
```go
package main

import (
	"context"
	"github.com/mark3labs/mcp-go/pkg"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/protocol"
	"github.com/mark3labs/mcp-go/transport"
)

func main() {
	transportClient, err := transport.NewStdioClientTransport("cloudinary", nil,
		transport.WithStdioClientOptionLogger(pkg.DebugLogger),
		transport.WithStdioClientOptionEnv("cloud=cyeam", "key=key1", "secret=password"))
	if err != nil {
		log.Fatalf("Failed to create transport client: %v", err)
	}
	// Initialize MCP client
	mcpClient, err := client.NewClient(transportClient)
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}
	defer mcpClient.Close()

	// Get available tools
	ctx := context.Background()
	tools, err := mcpClient.ListTools(ctx)
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}
	for _, tool := range tools.Tools {
		log.Printf("Tool Name: %+v, Description: %s, Required: %+v", tool.Name, tool.Description, tool.InputSchema.Required)
		if tool.Name == "cloudinary" {
			req := &protocol.CallToolRequest{
				Name: tool.Name,
				Arguments: map[string]interface{}{
					"file_path": "/Users/cyeam/Downloads/abc.jpg",
				},
			}
			resp, err := mcpClient.CallTool(context.Background(), req)
			if err != nil {
				log.Fatalf("Failed to call tool: %v", err)
			} else {
				log.Printf("Tool Response: %+v", resp)
			}
		}
	}
}
```

## 许可证
本项目采用[LICENSE](LICENSE)文件中指定的许可证。
