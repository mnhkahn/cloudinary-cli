package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mnhkahn/gogogo/logger"
)

func Upload(ctx context.Context, cloud, key, secret, directory string, data []byte) (string, error) {
	cld, _ := cloudinary.NewFromParams(cloud, key, secret)

	fileName := uuid.New().String()
	publicID := fileName
	if directory != "" {
		publicID = directory + "/" + fileName
	}
	resp, err := cld.Upload.Upload(ctx, bytes.NewReader(data),
		uploader.UploadParams{
			PublicID:  publicID,
			Overwrite: api.Bool(true)})
	if err != nil {
		return "", err
	}
	return resp.SecureURL, nil
}

func handleCloudinaryUpload(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	filePathStr, ok := arguments["file_path"].(string)
	if !ok {
		return nil, errors.New("file_path must be a string")
	}

	directory := ""
	if dir, ok := arguments["directory"].(string); ok {
		directory = dir
	}

	logger.Info("Received file path: %+v, directory: %+v", filePathStr, directory)
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

	var data []byte
	var err error
	if checkStringType(filePathStr) == urlPath {
		resp, err := http.Get(filePathStr)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else if checkStringType(filePathStr) == filePath {
		data, err = os.ReadFile(filePathStr)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("file path is invalid")
	}
	res, err := Upload(context.Background(), cloud, key, secret, directory, data)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(res), nil
}

type pathType int8

const (
	unknown pathType = iota
	urlPath
	filePath
)

func checkStringType(s string) pathType {
	// 检查是否为URL
	u, err := url.Parse(s)
	if err == nil && u.Scheme != "" {
		return urlPath
	}

	// 检查是否为文件路径（存在性校验）
	if _, err := os.Stat(s); err == nil {
		return filePath
	}

	return unknown
}
func main() {
	logger.SetJack("/tmp/cyeam.log", 300)
	mcpServer := server.NewMCPServer(
		"cloudinary-mcp-server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	// Register cloudinary upload tool
	tool := mcp.NewTool("cloudinary",
		mcp.WithDescription("Upload file to cloudinary"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("file path in local directory or file url"),
		),
		mcp.WithString("directory",
			mcp.Description("directory to upload file to in cloudinary"),
		),
	)
	mcpServer.AddTool(tool, handleCloudinaryUpload)

	errCh := make(chan error)
	go func() {
		errCh <- server.ServeStdio(mcpServer)
	}()

	if err := signalWaiter(errCh); err != nil {
		logger.Errorf("signal waiter: %v", err)
		return
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
