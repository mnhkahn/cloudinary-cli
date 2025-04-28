# Cloudinary MCP Project

## Overview
This project is an MCP (Multi - Cloud Processing) server developed in Go. Its main function is to upload files to the Cloudinary cloud storage service. The server receives file path requests via the MCP protocol, uploads the corresponding files to Cloudinary, and finally returns the secure access links for the files.

## Environment Variables
Before running the project, you need to set the following environment variables:
- `cloud`: Cloudinary's cloud name.
- `key`: Cloudinary's API key.
- `secret`: Cloudinary's API secret.

You can obtain the keys from [here](https://console.cloudinary.com/settings/api-keys).

## Steps to Run
1. Ensure that the Go environment is correctly installed (version 1.23.1 or higher).
2. Set the required environment variables.
3. Execute the following command in the project root directory to run the project:
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

## Debug Code
```go
package main

import (
	"context"
	"github.com/ThinkInAIXYZ/go-mcp/pkg"
	"log"

	"github.com/ThinkInAIXYZ/go-mcp/client"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
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

## License
This project is licensed under the [LICENSE](LICENSE).