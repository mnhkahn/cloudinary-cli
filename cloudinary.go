package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mnhkahn/gogogo/logger"
	"gitee.com/cyeam/cloudinary_mcp/pkg/uploader"
)



func handleCloudinaryUpload(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	filePathsStr, ok := arguments["file_paths"].(string)
	if !ok {
		return nil, errors.New("file_paths must be a string")
	}

	// Split comma-separated file paths
	filePaths := strings.Split(filePathsStr, ",")

	// Check if file count exceeds 50
	if len(filePaths) > 50 {
		return nil, errors.New("file count exceeds maximum limit of 50")
	}

	// Trim whitespace from each file path
	for i, path := range filePaths {
		filePaths[i] = strings.TrimSpace(path)
	}

	directory := ""
	if dir, ok := arguments["directory"].(string); ok {
		directory = dir
	}

	logger.Info("Received file paths: %+v, directory: %+v", filePaths, directory)
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

	// Upload each file and collect results
	results := make(map[string]string)
	errors := make(map[string]string)

	for _, filePathStr := range filePaths {
		if filePathStr == "" {
			continue
		}

		// Extract file name from file path
		fileName := ""
		if uploader.CheckStringType(filePathStr) == uploader.URLPath {
			u, err := url.Parse(filePathStr)
			if err == nil {
				fileName = filepath.Base(u.Path)
			}
		} else if uploader.CheckStringType(filePathStr) == uploader.FilePath {
			fileName = filepath.Base(filePathStr)
		}

		if fileName == "" {
			errors[filePathStr] = "failed to extract file name"
			continue
		}

		data, _, err := uploader.ReadFileData(filePathStr)
		if err != nil {
			errors[filePathStr] = err.Error()
			continue
		}

		res, err := uploader.Upload(context.Background(), cloud, key, secret, directory, fileName, data, false)
		if err != nil {
			errors[filePathStr] = err.Error()
			continue
		}

		results[filePathStr] = res
	}

	// Prepare response
	response := "Upload results:\n"
	if len(results) > 0 {
		response += "Successful uploads:\n"
		for path, url := range results {
			response += fmt.Sprintf("%s: %s\n", path, url)
		}
	}
	if len(errors) > 0 {
		response += "Failed uploads:\n"
		for path, err := range errors {
			response += fmt.Sprintf("%s: %s\n", path, err)
		}
	}

	return mcp.NewToolResultText(response), nil
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
		mcp.WithDescription("Upload files to cloudinary"),
		mcp.WithString("file_paths",
			mcp.Required(),
			mcp.Description("comma-separated list of file paths in local directory or file urls, max 50 files"),
		),
		mcp.WithString("directory",
			mcp.Description("directory to upload files to in cloudinary"),
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
