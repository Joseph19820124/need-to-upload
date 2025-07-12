# Claude Desktop 快速配置指南

## 你的 MCP Server 端点

部署到 Railway 后，你的 MCP Server 提供以下端点：

- **连接端点**: `POST /api/v1/connect`
- **RPC 端点**: `POST /api/v1/rpc` 
- **SSE 事件流**: `GET /api/v1/events`
- **健康检查**: `GET /api/v1/health`
- **断开连接**: `POST /api/v1/disconnect`

## Claude Desktop 配置

### 📁 配置文件位置

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

### 📝 配置内容

将以下配置添加到你的 `claude_desktop_config.json` 文件中：

```json
{
  "mcpServers": {
    "github-railway": {
      "command": "node",
      "args": [
        "-e",
        "
        const http = require('http');
        const https = require('https');
        const url = require('url');
        
        const SERVER_URL = 'https://your-app.railway.app';
        
        // MCP 客户端实现
        class MCPClient {
          constructor(baseUrl) {
            this.baseUrl = baseUrl;
            this.sessionId = null;
          }
          
          async connect() {
            const response = await this.request('/api/v1/connect', {
              method: 'POST',
              body: JSON.stringify({
                clientInfo: {
                  name: 'claude-desktop',
                  version: '1.0.0'
                }
              })
            });
            this.sessionId = response.sessionId;
            return response;
          }
          
          async rpc(method, params) {
            return await this.request('/api/v1/rpc', {
              method: 'POST', 
              body: JSON.stringify({
                jsonrpc: '2.0',
                id: Date.now(),
                method: method,
                params: params || {}
              })
            });
          }
          
          async request(path, options = {}) {
            const fullUrl = this.baseUrl + path;
            const urlObj = new URL(fullUrl);
            const client = urlObj.protocol === 'https:' ? https : http;
            
            return new Promise((resolve, reject) => {
              const req = client.request(fullUrl, {
                method: options.method || 'GET',
                headers: {
                  'Content-Type': 'application/json',
                  'X-Session-ID': this.sessionId,
                  ...options.headers
                }
              }, (res) => {
                let data = '';
                res.on('data', chunk => data += chunk);
                res.on('end', () => {
                  try {
                    resolve(JSON.parse(data));
                  } catch (e) {
                    resolve(data);
                  }
                });
              });
              
              req.on('error', reject);
              if (options.body) req.write(options.body);
              req.end();
            });
          }
        }
        
        // 启动客户端
        async function main() {
          const client = new MCPClient(SERVER_URL);
          
          try {
            await client.connect();
            console.log('Connected to MCP server');
            
            // 处理 stdin 输入
            process.stdin.on('data', async (data) => {
              try {
                const message = JSON.parse(data.toString());
                const response = await client.rpc(message.method, message.params);
                console.log(JSON.stringify(response));
              } catch (error) {
                console.error('Error:', error.message);
              }
            });
            
          } catch (error) {
            console.error('Failed to connect:', error.message);
            process.exit(1);
          }
        }
        
        main();
        "
      ],
      "env": {
        "NODE_ENV": "production"
      }
    }
  }
}
```

### 🔧 简化配置（推荐）

如果你想要更简单的配置，创建一个单独的客户端脚本文件：

**第一步**: 创建 `~/mcp-github-client.js`：

```javascript
const http = require('http');
const https = require('https');

const SERVER_URL = process.env.MCP_SERVER_URL || 'https://your-app.railway.app';

class MCPClient {
  constructor(baseUrl) {
    this.baseUrl = baseUrl;
    this.sessionId = null;
  }
  
  async connect() {
    const response = await this.request('/api/v1/connect', {
      method: 'POST',
      body: JSON.stringify({
        clientInfo: { name: 'claude-desktop', version: '1.0.0' }
      })
    });
    this.sessionId = response.sessionId;
    return response;
  }
  
  async rpc(method, params) {
    return await this.request('/api/v1/rpc', {
      method: 'POST',
      body: JSON.stringify({
        jsonrpc: '2.0',
        id: Date.now(),
        method,
        params: params || {}
      })
    });
  }
  
  async request(path, options = {}) {
    const fullUrl = this.baseUrl + path;
    const client = fullUrl.startsWith('https') ? https : http;
    
    return new Promise((resolve, reject) => {
      const req = client.request(fullUrl, {
        method: options.method || 'GET',
        headers: {
          'Content-Type': 'application/json',
          'X-Session-ID': this.sessionId,
          ...options.headers
        }
      }, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
          try {
            resolve(JSON.parse(data));
          } catch (e) {
            resolve(data);
          }
        });
      });
      
      req.on('error', reject);
      if (options.body) req.write(options.body);
      req.end();
    });
  }
}

async function main() {
  const client = new MCPClient(SERVER_URL);
  
  try {
    await client.connect();
    console.log('✅ Connected to GitHub MCP server');
    
    process.stdin.on('data', async (data) => {
      try {
        const message = JSON.parse(data.toString());
        const response = await client.rpc(message.method, message.params);
        process.stdout.write(JSON.stringify(response) + '\n');
      } catch (error) {
        console.error('❌ Error:', error.message);
      }
    });
    
  } catch (error) {
    console.error('❌ Failed to connect:', error.message);
    process.exit(1);
  }
}

main().catch(console.error);
```

**第二步**: 更新 `claude_desktop_config.json`：

```json
{
  "mcpServers": {
    "github-railway": {
      "command": "node",
      "args": ["/Users/your-username/mcp-github-client.js"],
      "env": {
        "MCP_SERVER_URL": "https://your-actual-railway-domain.railway.app"
      }
    }
  }
}
```

## 🚀 部署后配置步骤

### 1. 获取 Railway 域名
部署完成后，从 Railway Dashboard 获取你的应用域名

### 2. 测试连接
```bash
# 测试健康检查
curl https://your-app.railway.app/api/v1/health

# 测试连接端点
curl -X POST https://your-app.railway.app/api/v1/connect \
  -H "Content-Type: application/json" \
  -d '{"clientInfo":{"name":"test","version":"1.0.0"}}'
```

### 3. 更新配置
将配置中的 `your-app.railway.app` 替换为实际域名

### 4. 重启 Claude Desktop
保存配置后重启 Claude Desktop

## 🔍 验证连接

成功配置后，在 Claude Desktop 中你应该能够：

- ✅ 使用 GitHub 相关工具
- ✅ 查询仓库信息  
- ✅ 创建和管理 Issues
- ✅ 搜索代码和用户

## 🐛 故障排除

### 连接失败
1. 检查 Railway 应用状态
2. 验证域名是否正确
3. 确认 `GITHUB_TOKEN` 环境变量已设置

### 权限错误
1. 检查 GitHub token 权限
2. 确认 token 未过期

### Claude Desktop 日志
检查 Claude Desktop 控制台输出查看错误信息

---

配置完成后，你就可以在 Claude Desktop 中使用你的 GitHub MCP Server 了！🎉