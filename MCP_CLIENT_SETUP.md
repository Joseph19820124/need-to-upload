# MCP 客户端配置指南 (Claude Desktop)

本文档说明如何配置 Claude Desktop 连接到部署在 Railway 上的 GitHub MCP Server。

## 概述

你的项目是一个 **MCP Server**，通过 HTTP/SSE 提供 GitHub 相关的工具和功能。要使用这些功能，需要配置 **MCP Client**（如 Claude Desktop）来连接到你的服务器。

## Claude Desktop 配置

### 1. 找到配置文件位置

Claude Desktop 的 MCP 配置文件位置：

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

### 2. 配置文件内容

在 `claude_desktop_config.json` 文件中添加以下配置：

```json
{
  "mcpServers": {
    "github-mcp-http": {
      "command": "node",
      "args": [
        "/path/to/mcp-client-http.js",
        "https://your-app.railway.app"
      ],
      "env": {
        "NODE_ENV": "production"
      }
    }
  }
}
```

### 3. HTTP MCP 客户端脚本

由于你的 MCP Server 使用 HTTP/SSE 传输，需要一个 HTTP MCP 客户端脚本。创建 `mcp-client-http.js`：

```javascript
#!/usr/bin/env node

const { SSEMCPClient } = require("@modelcontextprotocol/sdk/client/sse.js");
const { StdioClientTransport } = require("@modelcontextprotocol/sdk/client/stdio.js");

async function main() {
  const serverUrl = process.argv[2];
  if (!serverUrl) {
    console.error("Usage: node mcp-client-http.js <server-url>");
    process.exit(1);
  }

  try {
    // 创建 SSE 客户端连接到 HTTP MCP Server
    const client = new SSEMCPClient(serverUrl + "/sse");
    
    // 创建 stdio 传输用于与 Claude Desktop 通信
    const transport = new StdioClientTransport();
    
    await client.connect();
    await transport.start();
    
    // 转发消息
    transport.onmessage = async (message) => {
      const response = await client.request(message);
      transport.send(response);
    };
    
    client.onmessage = (message) => {
      transport.send(message);
    };
    
  } catch (error) {
    console.error("Failed to connect to MCP server:", error);
    process.exit(1);
  }
}

main().catch(console.error);
```

### 4. 简化配置（推荐）

如果你的 MCP Server 支持直接的 HTTP 连接，可以使用更简单的配置：

```json
{
  "mcpServers": {
    "github-mcp-railway": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "-H", "Content-Type: application/json",
        "-H", "Accept: text/event-stream",
        "--data-binary", "@-",
        "https://your-app.railway.app/mcp"
      ],
      "env": {}
    }
  }
}
```

## 部署后的配置步骤

### 第一步：获取 Railway 域名

1. 部署到 Railway 后，获取你的应用域名
2. 格式通常为：`https://your-app.railway.app`

### 第二步：验证 MCP Server 可访问性

```bash
# 检查健康状态
curl https://your-app.railway.app/api/v1/health

# 检查 MCP 端点
curl -X POST https://your-app.railway.app/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
```

### 第三步：更新 Claude Desktop 配置

1. 打开 `claude_desktop_config.json`
2. 替换 `your-app.railway.app` 为实际的 Railway 域名
3. 保存文件
4. 重启 Claude Desktop

### 示例完整配置

```json
{
  "mcpServers": {
    "github-mcp-production": {
      "command": "npx",
      "args": [
        "@modelcontextprotocol/client-http",
        "https://your-actual-railway-domain.railway.app"
      ],
      "env": {
        "NODE_ENV": "production",
        "MCP_CLIENT_TIMEOUT": "30000"
      }
    }
  },
  "globalShortcut": {
    "enabled": true,
    "key": "CommandOrControl+Shift+Space"
  }
}
```

## 验证连接

### 1. Claude Desktop 日志

检查 Claude Desktop 的日志输出：
- macOS: 在 Console.app 中搜索 "Claude"
- Windows: 检查事件查看器
- Linux: 检查系统日志

### 2. 测试 MCP 功能

在 Claude Desktop 中尝试使用 GitHub 相关功能：
- 询问仓库信息
- 请求创建 issue
- 搜索代码

### 3. 调试连接问题

如果连接失败，检查：

1. **Railway 应用状态**：确保应用正在运行
2. **网络连接**：确保可以访问 Railway 域名
3. **MCP 端点**：验证 `/mcp` 端点响应正常
4. **环境变量**：确保 `GITHUB_TOKEN` 已正确设置

## 高级配置

### 自定义传输配置

```json
{
  "mcpServers": {
    "github-mcp-custom": {
      "command": "node",
      "args": ["-e", "
        const { MCPClient } = require('@modelcontextprotocol/sdk/client');
        const client = new MCPClient('https://your-app.railway.app');
        client.connect().then(() => console.log('Connected'));
      "],
      "env": {
        "MCP_SERVER_URL": "https://your-app.railway.app",
        "MCP_TIMEOUT": "60000"
      }
    }
  }
}
```

### 多环境配置

```json
{
  "mcpServers": {
    "github-mcp-dev": {
      "command": "curl",
      "args": ["-X", "POST", "http://localhost:8080/mcp"],
      "env": {"NODE_ENV": "development"}
    },
    "github-mcp-prod": {
      "command": "curl", 
      "args": ["-X", "POST", "https://your-app.railway.app/mcp"],
      "env": {"NODE_ENV": "production"}
    }
  }
}
```

## 故障排除

### 常见问题

1. **"Server not responding"**
   - 检查 Railway 应用是否正在运行
   - 验证域名是否正确

2. **"Authentication failed"**
   - 确认 `GITHUB_TOKEN` 环境变量已设置
   - 验证 token 权限

3. **"Connection timeout"**
   - 增加超时设置
   - 检查网络连接

### 调试命令

```bash
# 测试 MCP 服务器
curl -v https://your-app.railway.app/mcp

# 检查 SSE 连接
curl -N -H "Accept: text/event-stream" https://your-app.railway.app/sse

# 验证 GitHub token
curl -H "Authorization: token YOUR_TOKEN" https://api.github.com/user
```

---

配置完成后，Claude Desktop 就可以通过你的 Railway MCP Server 使用 GitHub 功能了！🚀