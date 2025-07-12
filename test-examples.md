# GitHub MCP HTTP Server - API Test Examples

## 基础测试命令

### 1. 健康检查
```bash
curl -w "\n" http://localhost:8080/api/v1/health
```

### 2. 建立连接
```bash
curl -X POST http://localhost:8080/api/v1/connect \
  -H "Content-Type: application/json" \
  -d '{
    "clientInfo": {
      "name": "test-client",
      "version": "1.0.0"
    }
  }'
```

### 3. 列出可用工具
```bash
SESSION_ID="your-session-id-here"

curl -X POST http://localhost:8080/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "id": 1
  }'
```

### 4. 列出可用资源
```bash
curl -X POST http://localhost:8080/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "method": "resources/list",
    "id": 2
  }'
```

### 5. 读取用户信息
```bash
curl -X POST http://localhost:8080/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "method": "resources/read",
    "params": {
      "uri": "github://user"
    },
    "id": 3
  }'
```

### 6. 列出仓库
```bash
curl -X POST http://localhost:8080/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "list_repositories",
      "arguments": {
        "type": "all",
        "sort": "updated"
      }
    },
    "id": 4
  }'
```

### 7. 获取特定仓库信息
```bash
curl -X POST http://localhost:8080/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "get_repository",
      "arguments": {
        "owner": "Joseph19820124",
        "repo": "github-mcp-http"
      }
    },
    "id": 5
  }'
```

### 8. 创建Issue（可选）
```bash
curl -X POST http://localhost:8080/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "create_issue",
      "arguments": {
        "owner": "Joseph19820124",
        "repo": "github-mcp-http",
        "title": "Test Issue from MCP Server",
        "body": "This is a test issue created via the GitHub MCP HTTP server."
      }
    },
    "id": 6
  }'
```

### 9. 测试Server-Sent Events
```bash
curl -N -H "X-Session-ID: $SESSION_ID" \
  http://localhost:8080/api/v1/events
```

### 10. 断开连接
```bash
curl -X POST http://localhost:8080/api/v1/disconnect \
  -H "X-Session-ID: $SESSION_ID"
```

## 完整的测试脚本

### 运行完整测试
```bash
./test-api.sh
```

### 运行快速测试
```bash
./quick-test.sh
```

## 一行式测试命令

### 获取会话ID并列出仓库
```bash
SESSION_ID=$(curl -s -X POST http://localhost:8080/api/v1/connect \
  -H "Content-Type: application/json" \
  -d '{"clientInfo": {"name": "test", "version": "1.0.0"}}' | \
  jq -r '.sessionId') && \
curl -s -X POST http://localhost:8080/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION_ID" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "list_repositories", "arguments": {"type": "all"}}, "id": 1}' | \
  jq '.result.content[0].text' | jq '.[].name'
```

### 获取用户信息
```bash
SESSION_ID=$(curl -s -X POST http://localhost:8080/api/v1/connect \
  -H "Content-Type: application/json" \
  -d '{"clientInfo": {"name": "test", "version": "1.0.0"}}' | \
  jq -r '.sessionId') && \
curl -s -X POST http://localhost:8080/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION_ID" \
  -d '{"jsonrpc": "2.0", "method": "resources/read", "params": {"uri": "github://user"}, "id": 1}' | \
  jq '.result.contents[0].text' | jq '.login'
```

## 错误测试

### 测试无效会话ID
```bash
curl -X POST http://localhost:8080/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: invalid-session" \
  -d '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}'
```

### 测试缺少会话ID
```bash
curl -X POST http://localhost:8080/api/v1/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}'
```

### 测试无效的RPC方法
```bash
SESSION_ID="your-session-id-here"

curl -X POST http://localhost:8080/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "method": "invalid/method",
    "id": 1
  }'
```

## 注意事项

1. 确保服务器在 `http://localhost:8080` 运行
2. 需要安装 `jq` 工具来格式化JSON输出：`brew install jq` (macOS)
3. 替换示例中的 `owner` 和 `repo` 为你的实际仓库
4. 会话ID有有效期限制（30分钟）
5. 创建Issue的测试请谨慎使用，避免在重要仓库创建不必要的Issue