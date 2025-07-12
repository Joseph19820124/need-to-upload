# 使用 Supergateway 连接 Claude Desktop

## 🎯 问题解决方案

你的 GitHub MCP Server 使用 **HTTP/SSE 传输**，但 Claude Desktop 只支持 **stdio 协议**。Supergateway 解决了这个问题，它将你的 SSE 服务器转换为 Claude Desktop 可以理解的 stdio 接口。

## 📋 架构说明

```
Claude Desktop ←→ Supergateway ←→ 你的 Railway MCP Server
   (stdio)         (SSE ↔ stdio)       (HTTP/SSE)
```

## 🚀 Claude Desktop 配置

### 方法 1: 使用 npx (推荐)

编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "github-railway": {
      "command": "npx",
      "args": [
        "-y",
        "supergateway",
        "--sse",
        "https://your-railway-app.railway.app/api/v1/events"
      ],
      "env": {
        "NODE_ENV": "production"
      }
    }
  }
}
```

### 方法 2: 使用 Docker

```json
{
  "mcpServers": {
    "github-railway-docker": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "supercorp/supergateway",
        "--sse",
        "https://your-railway-app.railway.app/api/v1/events"
      ],
      "env": {}
    }
  }
}
```

### 方法 3: 全功能配置 (带认证)

如果你的服务器需要认证或特殊配置：

```json
{
  "mcpServers": {
    "github-railway-auth": {
      "command": "npx",
      "args": [
        "-y", 
        "supergateway",
        "--sse",
        "https://your-railway-app.railway.app/api/v1/events",
        "--header",
        "Authorization: Bearer YOUR_AUTH_TOKEN",
        "--header",
        "X-Client-Name: claude-desktop"
      ],
      "env": {
        "DEBUG": "*"
      }
    }
  }
}
```

## 🔧 部署后配置步骤

### 第一步：确认你的 Railway 部署

1. 确保应用已部署到 Railway
2. 获取你的应用域名（例如：`https://your-app-name.railway.app`）
3. 测试 SSE 端点：

```bash
# 测试健康检查
curl https://your-app.railway.app/api/v1/health

# 测试 SSE 端点
curl -N -H "Accept: text/event-stream" \
  https://your-app.railway.app/api/v1/events
```

### 第二步：安装 Supergateway (可选)

你可以全局安装以获得更好的性能：

```bash
npm install -g supergateway
```

然后配置中使用：

```json
{
  "mcpServers": {
    "github-railway": {
      "command": "supergateway",
      "args": [
        "--sse",
        "https://your-railway-app.railway.app/api/v1/events"
      ]
    }
  }
}
```

### 第三步：测试本地连接

在配置 Claude Desktop 之前，先测试 supergateway：

```bash
# 测试 supergateway 连接
npx -y supergateway --sse "https://your-railway-app.railway.app/api/v1/events"
```

如果连接成功，你应该看到类似输出：
```
Connected to SSE endpoint: https://your-railway-app.railway.app/api/v1/events
Waiting for stdio input...
```

### 第四步：更新 Claude Desktop 配置

1. 替换配置中的 `your-railway-app.railway.app` 为你的实际 Railway 域名
2. 保存配置文件
3. **重启 Claude Desktop**

## 🧪 验证连接

### 检查 Claude Desktop 日志

重启 Claude Desktop 后，检查连接状态：

- **macOS**: 打开 Console.app，搜索 "Claude"
- **终端调试**: 使用 DEBUG 环境变量

```json
{
  "mcpServers": {
    "github-railway": {
      "command": "npx",
      "args": ["-y", "supergateway", "--sse", "https://your-app.railway.app/api/v1/events"],
      "env": {
        "DEBUG": "supergateway:*"
      }
    }
  }
}
```

### 测试 MCP 功能

在 Claude Desktop 中尝试：

1. **询问 GitHub 相关问题**："帮我查看我的 GitHub 仓库"
2. **请求创建 issue**："在我的仓库中创建一个 issue"
3. **搜索代码**："搜索我仓库中的特定代码"

如果配置正确，Claude 应该能够识别并使用你的 GitHub MCP 工具。

## 🔍 故障排除

### 常见问题

1. **"Failed to connect to SSE endpoint"**
   ```bash
   # 检查 Railway 应用状态
   curl https://your-app.railway.app/api/v1/health
   
   # 检查 SSE 端点
   curl -N https://your-app.railway.app/api/v1/events
   ```

2. **"GitHub token is required"**
   - 确认 Railway 中设置了 `GITHUB_TOKEN` 环境变量
   - 验证 token 有效性

3. **Claude Desktop 无法识别服务器**
   - 确认配置文件语法正确
   - 重启 Claude Desktop
   - 检查 supergateway 版本：`npx supergateway --version`

### 调试命令

```bash
# 测试 supergateway 详细输出
DEBUG=* npx -y supergateway --sse "https://your-app.railway.app/api/v1/events"

# 验证你的 MCP 服务器端点
curl -X POST https://your-app.railway.app/api/v1/connect \
  -H "Content-Type: application/json" \
  -d '{"clientInfo":{"name":"test","version":"1.0.0"}}'
```

### 高级配置

如果需要更多控制：

```json
{
  "mcpServers": {
    "github-railway-advanced": {
      "command": "npx",
      "args": [
        "-y",
        "supergateway", 
        "--sse",
        "https://your-app.railway.app/api/v1/events",
        "--timeout",
        "30000",
        "--reconnect",
        "true",
        "--header",
        "User-Agent: Claude-Desktop-Supergateway"
      ],
      "env": {
        "NODE_ENV": "production",
        "DEBUG": "supergateway:sse"
      }
    }
  }
}
```

## 📚 Supergateway 参数说明

| 参数 | 描述 | 示例 |
|------|------|------|
| `--sse` | SSE 端点 URL | `--sse "https://example.com/events"` |
| `--header` | 添加 HTTP 头 | `--header "Authorization: Bearer token"` |
| `--timeout` | 连接超时 (ms) | `--timeout 30000` |
| `--reconnect` | 自动重连 | `--reconnect true` |

---

配置完成后，Claude Desktop 就可以通过 Supergateway 使用你部署在 Railway 上的 GitHub MCP Server 了！🎉