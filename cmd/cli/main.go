package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gitee.com/cyeam/cloudinary_mcp/pkg/uploader"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file1> [file2] ...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nEnvironment variables (or .env file):\n")
		fmt.Fprintf(os.Stderr, "  CLOUDINARY_CLOUD     - Cloudinary cloud name\n")
		fmt.Fprintf(os.Stderr, "  CLOUDINARY_KEY       - Cloudinary API key\n")
		fmt.Fprintf(os.Stderr, "  CLOUDINARY_SECRET    - Cloudinary API secret\n")
		fmt.Fprintf(os.Stderr, "  CLOUDINARY_DIRECTORY - Upload directory (optional)\n")
		fmt.Fprintf(os.Stderr, "  CLOUDINARY_COMPRESS  - Auto compress images: true/false (default: true)\n")
		os.Exit(1)
	}

	_ = godotenv.Load()

	cloud := os.Getenv("CLOUDINARY_CLOUD")
	key := os.Getenv("CLOUDINARY_KEY")
	secret := os.Getenv("CLOUDINARY_SECRET")
	directory := os.Getenv("CLOUDINARY_DIRECTORY")
	compressStr := strings.ToLower(os.Getenv("CLOUDINARY_COMPRESS"))
	compress := compressStr != "false" && compressStr != "0" && compressStr != "no"

	if cloud == "" || key == "" || secret == "" {
		fmt.Fprintf(os.Stderr, "Error: CLOUDINARY_CLOUD, CLOUDINARY_KEY, CLOUDINARY_SECRET must be set\n")
		os.Exit(1)
	}

	results := make(map[string]string)
	errors := make(map[string]string)

	for _, filePathStr := range os.Args[1:] {
		if filePathStr == "" {
			continue
		}

		data, fileName, err := uploader.ReadFileData(filePathStr)
		if err != nil {
			errors[filePathStr] = err.Error()
			continue
		}

		res, err := uploader.Upload(context.Background(), cloud, key, secret, directory, fileName, data, compress)
		if err != nil {
			errors[filePathStr] = err.Error()
			continue
		}

		results[filePathStr] = res
	}

	if len(results) > 0 {
		fmt.Println("Successful uploads:")
		for path, url := range results {
			fmt.Printf("  %s -> %s\n", path, url)
		}
	}
	if len(errors) > 0 {
		fmt.Println("Failed uploads:")
		for path, err := range errors {
			fmt.Printf("  %s -> %s\n", path, err)
		}
	}

	if len(errors) > 0 {
		os.Exit(1)
	}
}
