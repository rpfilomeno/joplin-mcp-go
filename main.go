package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/pkg/browser"
)

const (
	defaultPort  = 41184
	defaultToken = "" // Set via config or UI
)

type Config struct {
	JoplinPort  int    `json:"joplin_port"`
	JoplinToken string `json:"joplin_token"`
	MCPPort     int    `json:"mcp_port"`
}

type JoplinClient struct {
	baseURL string
	token   string
	client  *http.Client
}

type MCPServer struct {
	joplin *JoplinClient
	config *Config
}

// MCP Protocol structures
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool parameter structures
type ListNotesParams struct {
	FolderID string `json:"folder_id,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Page     int    `json:"page,omitempty"`
}

type GetNoteParams struct {
	NoteID string `json:"note_id"`
}

type CreateNoteParams struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	FolderID string `json:"folder_id,omitempty"`
}

type UpdateNoteParams struct {
	NoteID string `json:"note_id"`
	Title  string `json:"title,omitempty"`
	Body   string `json:"body,omitempty"`
}

type SearchParams struct {
	Query string `json:"query"`
	Type  string `json:"type,omitempty"`
}

func NewJoplinClient(port int, token string) *JoplinClient {
	return &JoplinClient{
		baseURL: fmt.Sprintf("http://localhost:%d", port),
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (j *JoplinClient) makeRequest(method, endpoint string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", j.baseURL, endpoint)
	if j.token != "" {
		sep := "?"
		if strings.Contains(url, "?") {
			sep = "&"
		}
		url += fmt.Sprintf("%stoken=%s", sep, j.token)
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (j *JoplinClient) Ping() error {
	data, err := j.makeRequest("GET", "/ping", nil)
	if err != nil {
		return err
	}
	if string(data) != "JoplinClipperServer" {
		return fmt.Errorf("unexpected ping response: %s", string(data))
	}
	return nil
}

func NewMCPServer(config *Config) *MCPServer {
	return &MCPServer{
		joplin: NewJoplinClient(config.JoplinPort, config.JoplinToken),
		config: config,
	}
}

func (s *MCPServer) HandleRequest(req MCPRequest) MCPResponse {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    "joplin-mcp-server",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{},
			},
		}

	case "tools/list":
		resp.Result = map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "list_notes",
					"description": "List all notes or notes in a specific folder",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"folder_id": map[string]string{
								"type":        "string",
								"description": "Optional folder ID to filter notes",
							},
							"limit": map[string]interface{}{
								"type":        "number",
								"description": "Number of notes per page (max 100)",
								"default":     50,
							},
							"page": map[string]interface{}{
								"type":        "number",
								"description": "Page number (starts at 1)",
								"default":     1,
							},
						},
					},
				},
				{
					"name":        "get_note",
					"description": "Get a specific note by ID",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"note_id": map[string]string{
								"type":        "string",
								"description": "The ID of the note to retrieve",
							},
						},
						"required": []string{"note_id"},
					},
				},
				{
					"name":        "create_note",
					"description": "Create a new note",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"title": map[string]string{
								"type":        "string",
								"description": "The title of the note",
							},
							"body": map[string]string{
								"type":        "string",
								"description": "The body of the note in Markdown",
							},
							"folder_id": map[string]string{
								"type":        "string",
								"description": "Optional folder ID to create the note in",
							},
						},
						"required": []string{"title", "body"},
					},
				},
				{
					"name":        "update_note",
					"description": "Update an existing note",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"note_id": map[string]string{
								"type":        "string",
								"description": "The ID of the note to update",
							},
							"title": map[string]string{
								"type":        "string",
								"description": "New title for the note",
							},
							"body": map[string]string{
								"type":        "string",
								"description": "New body for the note in Markdown",
							},
						},
						"required": []string{"note_id"},
					},
				},
				{
					"name":        "delete_note",
					"description": "Delete a note (moves to trash by default)",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"note_id": map[string]string{
								"type":        "string",
								"description": "The ID of the note to delete",
							},
						},
						"required": []string{"note_id"},
					},
				},
				{
					"name":        "search_notes",
					"description": "Search for notes using Joplin's search syntax",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"query": map[string]string{
								"type":        "string",
								"description": "Search query",
							},
							"type": map[string]string{
								"type":        "string",
								"description": "Type of item to search (note, folder, tag)",
							},
						},
						"required": []string{"query"},
					},
				},
				{
					"name":        "list_folders",
					"description": "List all notebooks/folders",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
				{
					"name":        "list_tags",
					"description": "List all tags",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		}

	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &MCPError{Code: -32602, Message: "Invalid params"}
			return resp
		}

		result, err := s.executeTool(params.Name, params.Arguments)
		if err != nil {
			resp.Error = &MCPError{Code: -32603, Message: err.Error()}
			return resp
		}
		resp.Result = map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": result,
				},
			},
		}

	default:
		resp.Error = &MCPError{Code: -32601, Message: "Method not found"}
	}

	return resp
}

func (s *MCPServer) executeTool(name string, args json.RawMessage) (string, error) {
	switch name {
	case "list_notes":
		var params ListNotesParams
		json.Unmarshal(args, &params)
		endpoint := "/notes"
		if params.FolderID != "" {
			endpoint = fmt.Sprintf("/folders/%s/notes", params.FolderID)
		}
		if params.Limit > 0 {
			endpoint += fmt.Sprintf("?limit=%d", params.Limit)
		}
		if params.Page > 1 {
			endpoint += fmt.Sprintf("&page=%d", params.Page)
		}
		data, err := s.joplin.makeRequest("GET", endpoint, nil)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "get_note":
		var params GetNoteParams
		json.Unmarshal(args, &params)
		data, err := s.joplin.makeRequest("GET", fmt.Sprintf("/notes/%s?fields=id,parent_id,title,body,markup_language", params.NoteID), nil)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "create_note":
		var params CreateNoteParams
		json.Unmarshal(args, &params)
		data, err := s.joplin.makeRequest("POST", "/notes", params)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "update_note":
		var params UpdateNoteParams
		json.Unmarshal(args, &params)
		noteID := params.NoteID
		updateData := make(map[string]string)
		if params.Title != "" {
			updateData["title"] = params.Title
		}
		if params.Body != "" {
			updateData["body"] = params.Body
		}
		data, err := s.joplin.makeRequest("PUT", fmt.Sprintf("/notes/%s", noteID), updateData)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "delete_note":
		var params GetNoteParams
		json.Unmarshal(args, &params)
		_, err := s.joplin.makeRequest("DELETE", fmt.Sprintf("/notes/%s", params.NoteID), nil)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Note %s deleted successfully", params.NoteID), nil

	case "search_notes":
		var params SearchParams
		json.Unmarshal(args, &params)
		endpoint := fmt.Sprintf("/search?query=%s", params.Query)
		if params.Type != "" {
			endpoint += fmt.Sprintf("&type=%s", params.Type)
		}
		data, err := s.joplin.makeRequest("GET", endpoint, nil)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "list_folders":
		data, err := s.joplin.makeRequest("GET", "/folders", nil)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "list_tags":
		data, err := s.joplin.makeRequest("GET", "/tags", nil)
		if err != nil {
			return "", err
		}
		return string(data), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *MCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp := s.HandleRequest(req)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func loadConfig() *Config {
	config := &Config{
		JoplinPort: defaultPort,
		MCPPort:    3000,
	}

	data, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(data, config)
	}

	return config
}

func saveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("config.json", data, 0644)
}

func onReady() {
	systray.SetIcon(getIcon())
	systray.SetTitle("Joplin MCP")
	systray.SetTooltip("Joplin MCP Server")

	config := loadConfig()
	mcpServer := NewMCPServer(config)

	// Start HTTP server
	go func() {
		http.HandleFunc("/", mcpServer.ServeHTTP)
		addr := fmt.Sprintf(":%d", config.MCPPort)
		log.Printf("MCP Server listening on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Test Joplin connection
	go func() {
		time.Sleep(1 * time.Second)
		if err := mcpServer.joplin.Ping(); err != nil {
			log.Printf("Failed to connect to Joplin: %v", err)
		} else {
			log.Println("Connected to Joplin successfully")
		}
	}()

	mStatus := systray.AddMenuItem("Status: Running", "Server status")
	mStatus.Disable()

	systray.AddSeparator()

	mAbout := systray.AddMenuItem("About", "Visit project page")
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	go func() {
		for {
			select {
			case <-mAbout.ClickedCh:
				if err := browser.OpenURL("https://github.com/rpfilomeno/joplin-mcp-go"); err != nil {
					log.Fatalf("could not open URL in browser: %v", err)
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	log.Println("Shutting down Joplin MCP Server")
}

func getIcon() []byte {

	iconData, err := os.ReadFile("app.ico")
	if err != nil {
		log.Printf("Error loading icon: %v", err)
		return []byte{
			0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
			0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
			0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0xf3, 0xff, 0x61, 0x00, 0x00, 0x00,
			0x19, 0x74, 0x45, 0x58, 0x74, 0x53, 0x6f, 0x66, 0x74, 0x77, 0x61, 0x72,
			0x65, 0x00, 0x41, 0x64, 0x6f, 0x62, 0x65, 0x20, 0x49, 0x6d, 0x61, 0x67,
			0x65, 0x52, 0x65, 0x61, 0x64, 0x79, 0x71, 0xc9, 0x65, 0x3c, 0x00, 0x00,
			0x00, 0x18, 0x49, 0x44, 0x41, 0x54, 0x78, 0xda, 0x62, 0xfc, 0xff, 0xff,
			0x3f, 0x03, 0x25, 0x80, 0x89, 0x81, 0x81, 0x01, 0x00, 0x0b, 0x10, 0x03,
			0x00, 0x8e, 0x0d, 0xaa, 0x4e, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
			0x44, 0xae, 0x42, 0x60, 0x82,
		}
	} else {
		return iconData
	}
}

func main() {
	systray.Run(onReady, onExit)
}
