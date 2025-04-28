# Cloudinary MCP 项目

## 概述
本项目是一个基于Go语言开发的MCP（Multi - Cloud Processing）服务器，主要功能是将文件上传到Cloudinary云存储服务。服务器通过MCP协议接收文件路径请求，并将对应文件上传到Cloudinary，最后返回文件的安全访问链接。

## 环境变量
运行项目前，需要设置以下环境变量：
- `cloud`: Cloudinary的云名称。
- `key`: Cloudinary的API密钥。
- `secret`: Cloudinary的API密钥密码。

## 运行步骤
1. 确保Go环境已正确安装（版本1.23.1及以上）。
2. 设置所需的环境变量。
3. 在项目根目录下执行以下命令运行项目：
```bash
go run cloudinary.go
```

## 许可证
本项目采用[LICENSE](LICENSE)文件中指定的许可证。
