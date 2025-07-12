# Railway 部署指南

本文档提供了将 GitHub MCP HTTP 服务部署到 Railway 的详细步骤。

## 前提条件

1. **GitHub 账户** - 代码需要推送到 GitHub 仓库
2. **Railway 账户** - 访问 [railway.app](https://railway.app) 注册
3. **GitHub Personal Access Token** - 用于访问 GitHub API

## 第一步：准备 GitHub Token

1. 访问 [GitHub Settings > Developer settings > Personal access tokens](https://github.com/settings/tokens)
2. 点击 "Generate new token (classic)"
3. 设置以下权限：
   - `repo` - 访问仓库
   - `read:user` - 读取用户信息
4. 复制生成的 token（格式：`ghp_xxxxxxxxxxxx`）

## 第二步：推送代码到 GitHub

确保你的代码已推送到 GitHub 仓库：

```bash
git add .
git commit -m "Add Railway deployment configuration"
git push origin main
```

## 第三步：在 Railway 创建项目

1. 访问 [Railway Dashboard](https://railway.app/dashboard)
2. 点击 "New Project"
3. 选择 "Deploy from GitHub repo"
4. 选择你的 GitHub 仓库 `need-to-upload`
5. Railway 会自动检测到 Dockerfile 并开始构建

## 第四步：配置环境变量

在 Railway 项目中设置以下环境变量：

### 必需的环境变量

| 变量名 | 值 | 说明 |
|--------|------|------|
| `GITHUB_TOKEN` | `ghp_your_actual_token_here` | 你的 GitHub Personal Access Token |

### 可选的环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `GITHUB_MCP_HOST` | `0.0.0.0` | 服务监听地址 |
| `GITHUB_MCP_GITHUB_READ_ONLY` | `false` | 是否启用只读模式 |

**注意**：不要设置 `GITHUB_MCP_PORT`，Railway 会自动通过 `PORT` 环境变量分配端口。

### 设置环境变量的步骤

1. 在 Railway 项目页面，点击项目名称
2. 点击 "Variables" 标签页
3. 点击 "New Variable"
4. 添加变量名和值
5. 点击 "Add" 保存

## 第五步：配置域名（可选）

1. 在项目页面点击 "Settings"
2. 找到 "Networking" 部分
3. 点击 "Generate Domain" 获取公共域名
4. 或者添加自定义域名

## 第六步：监控部署

1. 在 "Deployments" 标签页查看构建日志
2. 等待构建完成（通常需要 2-5 分钟）
3. 查看 "Logs" 标签页确认服务正常启动

## 验证部署

部署成功后，你可以通过以下方式验证：

1. **健康检查端点**：
   ```bash
   curl https://your-app.railway.app/api/v1/health
   ```
   
   应该返回：
   ```json
   {"status":"healthy","time":"2025-07-12T06:00:00Z","version":"1.0.0"}
   ```

2. **检查日志**：
   在 Railway 的 "Logs" 标签页应该看到：
   ```
   Starting HTTP/SSE server on 0.0.0.0:PORT
   ```

## 故障排除

### 常见问题

1. **"GitHub token is required" 错误**
   - 确认 `GITHUB_TOKEN` 环境变量已正确设置
   - 验证 token 仍然有效且有正确的权限

2. **构建失败**
   - 检查 Dockerfile 语法
   - 查看构建日志中的错误信息

3. **应用无法启动**
   - 检查应用日志
   - 确认所有必需的环境变量已设置

4. **端口问题**
   - 不要手动设置 `PORT` 环境变量
   - Railway 会自动分配端口

### 获取帮助

- 查看 Railway [官方文档](https://docs.railway.app/)
- 检查项目的 "Logs" 标签页获取详细错误信息
- 查看 GitHub 仓库的 Issues

## 环境变量优先级

应用按以下优先级读取配置：

1. `PORT` 环境变量（Railway 自动设置）
2. `GITHUB_TOKEN` 环境变量
3. `GITHUB_MCP_*` 环境变量
4. 命令行参数
5. 配置文件默认值

## 自动部署

Railway 支持自动部署，当你推送代码到 GitHub 主分支时，会自动触发重新部署。

---

**部署完成！** 🚀

你的 GitHub MCP HTTP 服务现在已在 Railway 上运行，可以通过生成的域名访问。