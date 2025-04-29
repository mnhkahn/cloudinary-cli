package main

import (
	"bytes"
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ThinkInAIXYZ/go-mcp/pkg"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
	"github.com/mnhkahn/gogogo/logger"
)

func Upload(ctx context.Context, cloud, key, secret string, data []byte) (string, error) {
	cld, _ := cloudinary.NewFromParams(cloud, key, secret)

	fileName := uuid.New().String()
	resp, err := cld.Upload.Upload(ctx, bytes.NewReader(data),
		uploader.UploadParams{
			PublicID:  fileName,
			Overwrite: api.Bool(true)})
	if err != nil {
		return "", err
	}
	return resp.SecureURL, nil
}

type cloudinaryUploadRequest struct {
	FilePath string `json:"file_path" description:"file_path" required:"true"` // Use field tag to describe input schema
}

func handleCloudinaryUpload(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var uploadRequest cloudinaryUploadRequest
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &uploadRequest); err != nil {
		return nil, err
	}

	logger.Info("Received file path: %+v, %+v", uploadRequest)
	cloud, key, secret := "", "", ""
	for _, env := range os.Environ() {
		ks := strings.Split(env, "=")
		if len(ks) == 2 {
			k, v := ks[0], ks[1]
			if k == "cloud" {
				cloud = v
			} else if k == "key" {
				key = v
			} else if k == "secret" {
				secret = v
			}
		}
	}
	logger.Infof("Received cloud, key, secret: %+v, %+v, %+v", cloud, key, secret)

	file, err := os.ReadFile(uploadRequest.FilePath)
	if err != nil {
		return nil, err
	}
	res, err := Upload(ctx, cloud, key, secret, file)
	if err != nil {
		return nil, err
	}

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			protocol.TextContent{
				Type: "text",
				Text: res,
			},
		},
	}, nil
}

func main() {
	logger.SetJack("/tmp/cyeam.log", 300)
	mcpServer, err := server.NewServer(
		transport.NewStdioServerTransport(transport.WithStdioServerOptionLogger(pkg.DebugLogger)),
		server.WithServerInfo(protocol.Implementation{
			Name:    "current-time-v2-server",
			Version: "1.0.0",
		}),
	)
	if err != nil {
		logger.Errorf("Failed to create server: %v", err)
		return
	}
	// Register time query tool
	tool, err := protocol.NewTool("cloudinary", "Upload file to cloudinary", cloudinaryUploadRequest{})
	if err != nil {
		logger.Errorf("Failed to create tool: %v", err)
		return
	}
	mcpServer.RegisterTool(tool, handleCloudinaryUpload)
	errCh := make(chan error)
	go func() {
		errCh <- mcpServer.Run()
	}()

	if err = signalWaiter(errCh); err != nil {
		logger.Errorf("signal waiter: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := mcpServer.Shutdown(ctx); err != nil {
		logger.Errorf("Shutdown error: %v", err)
	}
}

func signalWaiter(errCh chan error) error {
	signalToNotify := []os.Signal{syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM}
	if signal.Ignored(syscall.SIGHUP) {
		signalToNotify = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, signalToNotify...)

	select {
	case sig := <-signals:
		switch sig {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
			logger.Infof("Received signal: %s\n", sig)
			// graceful shutdown
			return nil
		}
	case err := <-errCh:
		return err
	}

	return nil
}
