# Joplin MCP Server

A Windows system tray application that exposes Joplin notes through the Model Context Protocol (MCP), enabling AI assistants like Claude to read, create, update, and search your Joplin notes.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.21%2B-blue.svg)
![Platform](https://img.shields.io/badge/platform-windows-lightgrey.svg)

## 🌟 Features

- **🔔 System Tray Integration** - Lightweight, runs quietly in your Windows system tray
- **🤖 MCP Protocol** - Full implementation of Model Context Protocol 2024-11-05
- **📝 Complete Joplin Access** - All CRUD operations for notes, notebooks, and tags
- **🔍 Advanced Search** - Leverage Joplin's powerful search syntax
- **⚡ Fast & Efficient** - Native Go application with minimal resource usage
- **🔐 Secure** - Uses Joplin's built-in API token authentication
- **🎯 Easy Setup** - Simple JSON configuration

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Available Tools](#available-tools)
- [Integration with Claude Desktop](#integration-with-claude-desktop)
- [API Reference](#api-reference)
- [Troubleshooting](#troubleshooting)
- [Development](#development)
- [License](#license)

## 🔧 Prerequisites

Before you begin, ensure you have:

- **Go 1.21 or higher** - [Download Go](https://golang.org/dl/)
- **Joplin Desktop** - [Download Joplin](https://joplinapp.org/)
- **Windows OS** - Windows 10 or later (for system tray support)

## 📦 Installation

### Quick Start

Download a precompiled binary from the [Releasea Page](https://github.com/rpfilomeno/joplin-mcp-go/releases)

### Build From Source

1. **Clone or download the project:**
```bash
git clone <repository-url>
cd joplin-mcp-server
```

2. **Download dependencies:**
```bash
go mod download
```

3. **Build the application:**
```bash
# Standard build
go build -o joplin-mcp-server.exe

# Or build without console window (recommended)
go build -ldflags="-H windowsgui" -o joplin-mcp-server.exe
```

### Building for Production

For a smaller, optimized executable:
```bash
go build -ldflags="-H windowsgui -s -w" -o joplin-mcp-server.exe
```

This removes debug symbols and creates a windowless executable.

## ⚙️ Configuration

### Step 1: Enable Joplin Web Clipper

1. Open **Joplin Desktop**
2. Navigate to **Tools → Options → Web Clipper**
3. Enable **"Enable Web Clipper Service"**
4. Note the **port number** (usually 41184)
5. Copy the **Authorization token**

### Step 2: Create Configuration File

Create a `config.json` file in the same directory as the executable:

```json
{
  "joplin_port": 41184,
  "joplin_token": "your_token_from_joplin_here",
  "mcp_port": 3000
}
```

#### Configuration Options

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `joplin_port` | Port where Joplin Web Clipper runs | 41184 | No |
| `joplin_token` | Your Joplin API authorization token | - | **Yes** |
| `mcp_port` | Port for the MCP server to listen on | 3000 | No |

### Example Configuration

```json
{
  "joplin_port": 41184,
  "joplin_token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6.....",
  "mcp_port": 3000
}
```

## 🚀 Usage

### Starting the Server

Simply double-click `joplin-mcp-server.exe` or run from command line:

```bash
./joplin-mcp-server.exe
```

The application will:
1. ✅ Start and appear in your system tray
2. ✅ Launch the MCP server on the configured port (default: 3000)
3. ✅ Test the connection to Joplin
4. ✅ Log status messages (if console window is visible)

### System Tray Menu

Right-click the tray icon to access:

- **Status: Running** - Current server status (non-clickable)
- **Configure** - Configuration settings (future feature)
- **Quit** - Stop the server and exit

### Verifying the Connection

Test that everything is working:

```bash
# Test Joplin connection
curl http://localhost:41184/ping?token=YOUR_TOKEN

# Test MCP server
curl -X POST http://localhost:3000/ \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize"}'
```

## 🛠️ Available Tools

The MCP server provides 8 tools for interacting with Joplin:

### 1. **list_notes**
List all notes or filter by folder with pagination support.

```json
{
  "folder_id": "abc123",  // Optional: filter by folder
  "limit": 50,            // Optional: results per page (max 100)
  "page": 1               // Optional: page number
}
```

### 2. **get_note**
Retrieve a complete note by its ID.

```json
{
  "note_id": "def456"     // Required: note ID
}
```

### 3. **create_note**
Create a new note in Joplin.

```json
{
  "title": "My New Note",           // Required: note title
  "body": "# Hello\n\nContent...",  // Required: markdown content
  "folder_id": "abc123"             // Optional: target folder
}
```

### 4. **update_note**
Update an existing note's title or content.

```json
{
  "note_id": "def456",              // Required: note ID
  "title": "Updated Title",         // Optional: new title
  "body": "Updated content..."      // Optional: new content
}
```

### 5. **delete_note**
Delete a note (moves to trash by default).

```json
{
  "note_id": "def456"     // Required: note ID to delete
}
```

### 6. **search_notes**
Search notes using Joplin's powerful search syntax.

```json
{
  "query": "meeting notes",  // Required: search query
  "type": "note"            // Optional: note, folder, or tag
}
```

**Search Examples:**
- `"notebook:Work urgent"` - Search in specific notebook
- `"tag:important"` - Search by tag
- `"created:20241101"` - Notes created on date
- `"todo:1"` - Find todo items

### 7. **list_folders**
List all notebooks/folders in Joplin.

```json
{}  // No parameters required
```

### 8. **list_tags**
List all tags in Joplin.

```json
{}  // No parameters required
```

## 🤝 Integration with Claude Desktop

Add to your Claude Desktop configuration file:

**Location:** `%APPDATA%\Claude\claude_desktop_config.json`


```json
{
  "mcpServers": {
    "joplin": {
      "url": "http://localhost:3000",
      "transport": "http"
    }
  }
}
```

### Restart Claude Desktop

After adding the configuration:
1. Completely quit Claude Desktop
2. Restart the application
3. The Joplin MCP server tools should now be available

### Example Claude Interactions

Once configured, you can ask Claude to:

- *"Show me all my notes from the Work notebook"*
- *"Create a new note titled 'Meeting Notes' with today's agenda"*
- *"Search my notes for references to the Q4 budget"*
- *"Update my grocery list note with milk and eggs"*
- *"What tags do I have in Joplin?"*

## 📚 API Reference

### MCP Endpoint

The server exposes a single JSON-RPC endpoint:

```
POST http://localhost:3000/
Content-Type: application/json
```

### Request Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "tool_name",
    "arguments": {
      // Tool-specific arguments
    }
  }
}
```

### Response Format

**Success:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "... response data ..."
      }
    ]
  }
}
```

**Error:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32603,
    "message": "Error description"
  }
}
```

### MCP Methods

| Method | Description |
|--------|-------------|
| `initialize` | Initialize MCP connection |
| `tools/list` | Get list of available tools |
| `tools/call` | Execute a specific tool |

## 🔍 Troubleshooting

### Server Won't Start

**Problem:** Application crashes or doesn't appear in tray

**Solutions:**
- ✅ Verify `config.json` exists and is valid JSON
- ✅ Check that port 3000 is not in use by another application
- ✅ Ensure Joplin Desktop is running
- ✅ Try running from command line to see error messages

### Can't Connect to Joplin

**Problem:** "Failed to connect to Joplin" error

**Solutions:**
- ✅ Verify Joplin Web Clipper service is enabled
- ✅ Check the API token is correct (no extra spaces)
- ✅ Test manually: `curl http://localhost:41184/ping?token=YOUR_TOKEN`
- ✅ Try a different port if 41184 is in use

### Tools Return Errors

**Problem:** MCP tools fail or return unexpected data

**Solutions:**
- ✅ Verify note/folder IDs are correct (32-character hex strings)
- ✅ Check Joplin database isn't corrupted
- ✅ Ensure you have the latest version of Joplin
- ✅ Look at console logs for detailed error messages

### Claude Desktop Can't Find Server

**Problem:** Tools don't appear in Claude Desktop

**Solutions:**
- ✅ Verify path in `claude_desktop_config.json` is correct
- ✅ Use absolute paths (e.g., `C:\\Users\\...`)
- ✅ Restart Claude Desktop completely
- ✅ Check server is running in system tray

### Permission Issues

**Problem:** Access denied or file errors

**Solutions:**
- ✅ Run as administrator (right-click → Run as administrator)
- ✅ Check antivirus isn't blocking the application
- ✅ Ensure config.json is in the same directory as executable

## 💻 Development

### Project Structure

```
joplin-mcp-server/
├── main.go              # Main application code
├── go.mod               # Go module definition
├── app.ico              # app icon
├── config.json          # Configuration file (user-created)
├── README.md            # This file
└── joplin-mcp-server.exe # Compiled executable
```

### Key Components

**JoplinClient** - HTTP client for Joplin REST API
- Handles authentication with token
- Makes HTTP requests (GET, POST, PUT, DELETE)
- Parses JSON responses

**MCPServer** - Model Context Protocol server
- Implements JSON-RPC 2.0
- Handles MCP protocol methods
- Routes tool calls to Joplin client

**System Tray** - Windows integration
- Uses `getlantern/systray` library
- Provides status menu
- Runs in background

### Adding New Tools

To add a new Joplin tool:

1. Add tool definition in `tools/list` response
2. Add case in `executeTool` switch statement
3. Implement Joplin API call
4. Test with MCP client

### Building from Source

```bash
# Install dependencies
go mod tidy

# Build for development
go build -o joplin-mcp-server.exe

# Build for production
go build -ldflags="-H windowsgui -s -w" -o joplin-mcp-server.exe

# Cross-compile for Linux (experimental)
GOOS=linux GOARCH=amd64 go build -o joplin-mcp-server
```

### Dependencies

- `github.com/getlantern/systray` - System tray support
- Go standard library (net/http, encoding/json, etc.)

## 🗺️ Roadmap

Future enhancements planned:

- [ ] Configuration dialog UI
- [ ] Auto-detect Joplin port
- [ ] Support for attachments/resources
- [ ] Cross-platform support (Linux, macOS)
- [ ] WebSocket transport option
- [ ] Better error messages and logging


## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Areas for Contribution

- 🐛 Bug fixes
- 📝 Documentation improvements
- ✨ New features
- 🧪 Test coverage


## 📄 License

This project is licensed under the MIT License. See below for details:

```
MIT License

Copyright (c) 2024

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## 🔗 Resources

- [Joplin Official Website](https://joplinapp.org/)
- [Joplin REST API Documentation](https://joplinapp.org/api/references/rest_api/)
- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [Go Programming Language](https://golang.org/)
- [Systray Library](https://github.com/getlantern/systray)

## 📞 Support

If you encounter issues:

1. Check the [Troubleshooting](#troubleshooting) section
2. Review [closed issues](../../issues?q=is%3Aissue+is%3Aclosed)
3. Open a [new issue](../../issues/new) with:
   - Your Go version
   - Your Joplin version
   - Windows version
   - Error messages/logs
   - Steps to reproduce

## 🙏 Acknowledgments

- Joplin team for the excellent note-taking application
- Anthropic for the Model Context Protocol specification
- Go community for amazing libraries and tools

---

**Made with ❤️ for the Joplin and AI assistant communities**

*Star ⭐ this repository if you find it helpful!*
