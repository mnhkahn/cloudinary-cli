package uploader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type PathType int8

const (
	Unknown PathType = iota
	URLPath
	FilePath
)

func CheckStringType(s string) PathType {
	u, err := url.Parse(s)
	if err == nil && u.Scheme != "" {
		return URLPath
	}
	if _, err := os.Stat(s); err == nil {
		return FilePath
	}
	return Unknown
}

func IsImage(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp",
		".bmp", ".tiff", ".svg", ".ico", ".heic", ".heif", ".avif":
		return true
	}
	return false
}

func Upload(ctx context.Context, cloud, key, secret, directory, fileName string, data []byte, compress bool) (string, error) {
	cld, err := cloudinary.NewFromParams(cloud, key, secret)
	if err != nil {
		return "", fmt.Errorf("init cloudinary: %w", err)
	}

	ext := filepath.Ext(fileName)
	publicID := strings.TrimSuffix(fileName, ext)

	params := uploader.UploadParams{
		PublicID:  publicID,
		Folder:    directory,
		Overwrite: api.Bool(true),
	}

	if compress && IsImage(fileName) {
		params.Transformation = "q_auto"
	}

	resp, err := cld.Upload.Upload(ctx, bytes.NewReader(data), params)
	if err != nil {
		return "", err
	}
	return resp.SecureURL, nil
}

func ReadFileData(filePathStr string) ([]byte, string, error) {
	var data []byte
	var err error
	var fileName string

	pt := CheckStringType(filePathStr)
	switch pt {
	case URLPath:
		u, err := url.Parse(filePathStr)
		if err != nil {
			return nil, "", fmt.Errorf("parse url: %w", err)
		}
		fileName = filepath.Base(u.Path)
		resp, err := http.Get(filePathStr)
		if err != nil {
			return nil, "", fmt.Errorf("download: %w", err)
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", fmt.Errorf("read body: %w", err)
		}
	case FilePath:
		fileName = filepath.Base(filePathStr)
		data, err = os.ReadFile(filePathStr)
		if err != nil {
			return nil, "", fmt.Errorf("read file: %w", err)
		}
	default:
		return nil, "", fmt.Errorf("invalid path: %s", filePathStr)
	}

	return data, fileName, nil
}
