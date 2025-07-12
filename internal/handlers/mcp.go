package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type MCPHandler struct {
	client   *github.Client
	clientV4 *githubv4.Client
	readOnly bool
}

type InitializeResult struct {
	ServerInfo   ServerInfo   `json:"serverInfo"`
	Capabilities Capabilities `json:"capabilities"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Capabilities struct {
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe"`
	ListChanged bool `json:"listChanged"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type PromptsCapability struct {
	ListChanged bool `json:"listChanged"`
}

func NewMCPHandler(token string, readOnly bool) (*MCPHandler, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	client := github.NewClient(tc)
	clientV4 := githubv4.NewClient(tc)

	return &MCPHandler{
		client:   client,
		clientV4: clientV4,
		readOnly: readOnly,
	}, nil
}

func (h *MCPHandler) Initialize(ctx context.Context, clientName, clientVersion string) (*InitializeResult, error) {
	capabilities := Capabilities{
		Resources: &ResourcesCapability{
			Subscribe:   true,
			ListChanged: true,
		},
		Tools: &ToolsCapability{
			ListChanged: true,
		},
		Prompts: &PromptsCapability{
			ListChanged: true,
		},
	}

	serverInfo := ServerInfo{
		Name:    "github-mcp-http",
		Version: "1.0.0",
	}

	return &InitializeResult{
		ServerInfo:   serverInfo,
		Capabilities: capabilities,
	}, nil
}

func (h *MCPHandler) ProcessRPC(ctx context.Context, request json.RawMessage) (interface{}, error) {
	var rpcReq struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
		ID      interface{}     `json:"id"`
	}

	if err := json.Unmarshal(request, &rpcReq); err != nil {
		return nil, fmt.Errorf("invalid JSON-RPC request: %w", err)
	}

	switch rpcReq.Method {
	case "resources/list":
		return h.handleListResources(ctx, rpcReq.ID)
	case "resources/read":
		return h.handleReadResource(ctx, rpcReq.Params, rpcReq.ID)
	case "tools/list":
		return h.handleListTools(ctx, rpcReq.ID)
	case "tools/call":
		return h.handleCallTool(ctx, rpcReq.Params, rpcReq.ID)
	case "prompts/list":
		return h.handleListPrompts(ctx, rpcReq.ID)
	case "prompts/get":
		return h.handleGetPrompt(ctx, rpcReq.Params, rpcReq.ID)
	default:
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32601,
				"message": "Method not found",
			},
			"id": rpcReq.ID,
		}, nil
	}
}

func (h *MCPHandler) handleListResources(ctx context.Context, id interface{}) (interface{}, error) {
	resources := []map[string]interface{}{
		{
			"uri":         "github://repositories",
			"name":        "repositories",
			"description": "List of user repositories",
			"mimeType":    "application/json",
		},
		{
			"uri":         "github://user",
			"name":        "user",
			"description": "Current user information",
			"mimeType":    "application/json",
		},
	}

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
			"resources": resources,
		},
		"id": id,
	}, nil
}

func (h *MCPHandler) handleReadResource(ctx context.Context, params json.RawMessage, id interface{}) (interface{}, error) {
	var req struct {
		URI string `json:"uri"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid read resource request: %w", err)
	}

	switch req.URI {
	case "github://repositories":
		repos, _, err := h.client.Repositories.List(ctx, "", &github.RepositoryListOptions{
			Type:        "all",
			ListOptions: github.ListOptions{PerPage: 50},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		data, err := json.Marshal(repos)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal repositories: %w", err)
		}

		return map[string]interface{}{
			"jsonrpc": "2.0",
			"result": map[string]interface{}{
				"contents": []map[string]interface{}{
					{
						"uri":      req.URI,
						"mimeType": "application/json",
						"text":     string(data),
					},
				},
			},
			"id": id,
		}, nil

	case "github://user":
		user, _, err := h.client.Users.Get(ctx, "")
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		data, err := json.Marshal(user)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal user: %w", err)
		}

		return map[string]interface{}{
			"jsonrpc": "2.0",
			"result": map[string]interface{}{
				"contents": []map[string]interface{}{
					{
						"uri":      req.URI,
						"mimeType": "application/json",
						"text":     string(data),
					},
				},
			},
			"id": id,
		}, nil

	default:
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32602,
				"message": fmt.Sprintf("Unknown resource URI: %s", req.URI),
			},
			"id": id,
		}, nil
	}
}

func (h *MCPHandler) handleListTools(ctx context.Context, id interface{}) (interface{}, error) {
	tools := []map[string]interface{}{
		{
			"name":        "list_repositories",
			"description": "List user repositories",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Repository type (all, owner, member)",
						"default":     "all",
					},
					"sort": map[string]interface{}{
						"type":        "string",
						"description": "Sort order (created, updated, pushed, full_name)",
						"default":     "updated",
					},
				},
			},
		},
		{
			"name":        "get_repository",
			"description": "Get repository information",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Repository owner",
					},
					"repo": map[string]interface{}{
						"type":        "string",
						"description": "Repository name",
					},
				},
				"required": []string{"owner", "repo"},
			},
		},
	}

	if !h.readOnly {
		tools = append(tools, map[string]interface{}{
			"name":        "create_issue",
			"description": "Create a new issue",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Repository owner",
					},
					"repo": map[string]interface{}{
						"type":        "string",
						"description": "Repository name",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Issue title",
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "Issue body",
					},
				},
				"required": []string{"owner", "repo", "title"},
			},
		})
	}

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
			"tools": tools,
		},
		"id": id,
	}, nil
}

func (h *MCPHandler) handleCallTool(ctx context.Context, params json.RawMessage, id interface{}) (interface{}, error) {
	var req struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid call tool request: %w", err)
	}

	switch req.Name {
	case "list_repositories":
		return h.listRepositories(ctx, req.Arguments, id)
	case "get_repository":
		return h.getRepository(ctx, req.Arguments, id)
	case "create_issue":
		if h.readOnly {
			return map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    -32602,
					"message": "Tool not available in read-only mode",
				},
				"id": id,
			}, nil
		}
		return h.createIssue(ctx, req.Arguments, id)
	default:
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32602,
				"message": fmt.Sprintf("Unknown tool: %s", req.Name),
			},
			"id": id,
		}, nil
	}
}

func (h *MCPHandler) listRepositories(ctx context.Context, args map[string]interface{}, id interface{}) (interface{}, error) {
	repoType := "all"
	sort := "updated"

	if t, ok := args["type"].(string); ok {
		repoType = t
	}
	if s, ok := args["sort"].(string); ok {
		sort = s
	}

	repos, _, err := h.client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		Type:        repoType,
		Sort:        sort,
		ListOptions: github.ListOptions{PerPage: 50},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	data, err := json.Marshal(repos)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal repositories: %w", err)
	}

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": string(data),
				},
			},
		},
		"id": id,
	}, nil
}

func (h *MCPHandler) getRepository(ctx context.Context, args map[string]interface{}, id interface{}) (interface{}, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32602,
				"message": "owner is required",
			},
			"id": id,
		}, nil
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32602,
				"message": "repo is required",
			},
			"id": id,
		}, nil
	}

	repository, _, err := h.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	data, err := json.Marshal(repository)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal repository: %w", err)
	}

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": string(data),
				},
			},
		},
		"id": id,
	}, nil
}

func (h *MCPHandler) createIssue(ctx context.Context, args map[string]interface{}, id interface{}) (interface{}, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32602,
				"message": "owner is required",
			},
			"id": id,
		}, nil
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32602,
				"message": "repo is required",
			},
			"id": id,
		}, nil
	}

	title, ok := args["title"].(string)
	if !ok {
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32602,
				"message": "title is required",
			},
			"id": id,
		}, nil
	}

	body, _ := args["body"].(string)

	issue := &github.IssueRequest{
		Title: &title,
		Body:  &body,
	}

	createdIssue, _, err := h.client.Issues.Create(ctx, owner, repo, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	data, err := json.Marshal(createdIssue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal issue: %w", err)
	}

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": string(data),
				},
			},
		},
		"id": id,
	}, nil
}

func (h *MCPHandler) handleListPrompts(ctx context.Context, id interface{}) (interface{}, error) {
	prompts := []map[string]interface{}{
		{
			"name":        "analyze_repository",
			"description": "Analyze a GitHub repository for insights",
			"arguments": []map[string]interface{}{
				{
					"name":        "owner",
					"description": "Repository owner",
					"required":    true,
				},
				{
					"name":        "repo",
					"description": "Repository name",
					"required":    true,
				},
			},
		},
	}

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
			"prompts": prompts,
		},
		"id": id,
	}, nil
}

func (h *MCPHandler) handleGetPrompt(ctx context.Context, params json.RawMessage, id interface{}) (interface{}, error) {
	var req struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid get prompt request: %w", err)
	}

	switch req.Name {
	case "analyze_repository":
		owner := req.Arguments["owner"]
		repo := req.Arguments["repo"]

		prompt := fmt.Sprintf(`Analyze the GitHub repository %s/%s and provide insights on:

1. Repository overview and purpose
2. Code structure and organization  
3. Development activity and health
4. Key technologies and dependencies
5. Documentation quality
6. Community engagement

Please focus on actionable insights and recommendations.`, owner, repo)

		return map[string]interface{}{
			"jsonrpc": "2.0",
			"result": map[string]interface{}{
				"description": "Repository analysis prompt",
				"messages": []map[string]interface{}{
					{
						"role": "user",
						"content": map[string]interface{}{
							"type": "text",
							"text": prompt,
						},
					},
				},
			},
			"id": id,
		}, nil

	default:
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32602,
				"message": fmt.Sprintf("Unknown prompt: %s", req.Name),
			},
			"id": id,
		}, nil
	}
}