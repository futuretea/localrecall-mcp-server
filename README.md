# LocalRecall MCP Server

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server for LocalRecall knowledge base management.

## Features

- **Multiple Modes**: Supports stdio, HTTP, and SSE transport modes
- **Knowledge Management**: Full CRUD operations for LocalRecall collections and documents
- **Search Capabilities**: Semantic search across your knowledge base
- **Flexible Configuration**: Command-line flags, environment variables, or configuration files
- **Collection Isolation**: Lock the server to a single collection for security
- **Multiple Output Formats**: JSON, YAML output formats
- **Cross-platform**: Native binaries for Linux, macOS, and Windows

## Comparison with MCPs/localrecall

| Feature | localrecall-mcp-server | MCPs/localrecall |
|---------|------------------------|------------------|
| Transport Modes | stdio, HTTP, SSE | stdio only |
| Configuration | Multiple sources (flags, env, file) | Environment variables only |
| Collection Isolation | Enforced (hides cross-collection tools) | Default only (can be overridden) |
| Output Formats | JSON, YAML | JSON only |
| Tool Control | Enable/disable specific tools | All tools enabled |
| Architecture | Modular, extensible | Simple, monolithic |

## Installation

### From Source

```bash
git clone https://github.com/futuretea/localrecall-mcp-server.git
cd localrecall-mcp-server
make build
```

### Using Docker

```bash
docker build -t localrecall-mcp-server .
```

## Quick Start

### Stdio Mode (for MCP clients)

```bash
./localrecall-mcp-server \
  --localrecall-url http://localhost:8080 \
  --localrecall-collection my-collection
```

### HTTP/SSE Mode (network access)

```bash
./localrecall-mcp-server \
  --port 8080 \
  --localrecall-url http://localhost:8080 \
  --localrecall-api-key your-api-key
```

## Configuration

Configuration can be set via CLI flags, environment variables, or a config file.

### Priority (highest to lowest)
1. Command-line flags
2. Environment variables (prefix: `LOCALRECALL_MCP_`)
3. Configuration file
4. Default values

### CLI Options

```bash
./localrecall-mcp-server --help
```

| Option | Description | Default |
|--------|-------------|---------|
| `--config` | Config file path (YAML) | |
| `--port` | Port for HTTP/SSE mode (0 = stdio mode) | `0` |
| `--sse-base-url` | Public base URL for SSE endpoint | |
| `--log-level` | Log level (0-9) | `5` |
| `--localrecall-url` | LocalRecall API URL | `http://localhost:8080` |
| `--localrecall-api-key` | LocalRecall API key | |
| `--localrecall-collection` | Collection isolation (locks to this collection) | |
| `--list-output` | Output format (json, yaml) | `json` |
| `--output-filters` | Fields to filter from output | |
| `--enabled-tools` | Tools to enable | |
| `--disabled-tools` | Tools to disable | |

### Configuration File

Create `config.yaml`:

```yaml
port: 0  # 0 for stdio, or set a port like 8080 for HTTP/SSE

log_level: 5

localrecall_url: http://localhost:8080
localrecall_api_key: your-api-key
localrecall_collection: my-collection  # locks server to this collection

list_output: json

enabled_tools: []
disabled_tools: []
```

### Environment Variables

Use `LOCALRECALL_MCP_` prefix with underscores:

```bash
LOCALRECALL_MCP_PORT=8080
LOCALRECALL_MCP_LOCALRECALL_URL=http://localhost:8080
LOCALRECALL_MCP_LOCALRECALL_API_KEY=your-api-key
```

## MCP Client Integration

### Claude Code

```bash
claude mcp add localrecall -- /path/to/localrecall-mcp-server \
  --localrecall-url http://localhost:8080 \
  --localrecall-collection my-collection
```

### VS Code / Cursor

Add to `.vscode/mcp.json` or `~/.cursor/mcp.json`:

```json
{
  "servers": {
    "localrecall": {
      "command": "/path/to/localrecall-mcp-server",
      "args": [
        "--localrecall-url",
        "http://localhost:8080",
        "--localrecall-collection",
        "my-collection"
      ]
    }
  }
}
```

## Available Tools

### search
Search content in a LocalRecall collection.

**Parameters:**
- `query` (string, required): The search query
- `max_results` (number, optional): Maximum number of results (default: 5)
- `collection_name` (string, required*): The collection to search

### add_document
Add a document to a LocalRecall collection.

**Parameters:**
- `filename` (string, required): The filename for the document
- `file_path` (string, optional): Path to file to upload
- `file_content` (string, optional): File content as string
- `collection_name` (string, required*): The collection to add to

### create_collection
Create a new collection in LocalRecall. **Hidden when collection isolation is active.**

**Parameters:**
- `name` (string, required): The name of the collection to create

### reset_collection
Reset (clear) a collection in LocalRecall. **Hidden when collection isolation is active.**

**Parameters:**
- `name` (string, required): The name of the collection to reset

### list_collections
List all collections in LocalRecall. **Hidden when collection isolation is active.**

**Parameters:** None

### list_files
List files in a LocalRecall collection.

**Parameters:**
- `collection_name` (string, required*): The collection to list

### delete_entry
Delete an entry from a LocalRecall collection.

**Parameters:**
- `entry` (string, required): The filename of the entry to delete
- `collection_name` (string, required*): The collection to delete from

> **\*** When `--localrecall-collection` is set, `collection_name` is removed from all tool schemas and automatically enforced. The parameter is only required in multi-collection mode.

## HTTP/SSE Mode

When running with a port number, the server exposes these endpoints:

- `/healthz` - Health check
- `/mcp` - Streamable HTTP endpoint
- `/sse` - Server-Sent Events endpoint
- `/message` - Message endpoint for SSE clients

Example:

```bash
./localrecall-mcp-server --port 8080 \
  --localrecall-url http://localhost:8080 \
  --localrecall-api-key your-api-key
```

Access at: `http://localhost:8080`

## Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Format Code

```bash
make format
```

### Test with MCP Inspector

```bash
npx @modelcontextprotocol/inspector@latest $(pwd)/localrecall-mcp-server \
  --localrecall-url http://localhost:8080
```

## Architecture

```
localrecall-mcp-server/
├── cmd/                        # Application entry points
├── internal/cmd/               # Command-line interface
├── pkg/
│   ├── client/                 # LocalRecall API client
│   ├── core/                   # Core utilities (config, logging, version)
│   ├── server/                 # MCP and HTTP servers
│   └── toolset/                # Tool implementations
```

## License

Apache-2.0

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

- GitHub Issues: https://github.com/futuretea/localrecall-mcp-server/issues
- LocalRecall Documentation: https://github.com/mudler/LocalRecall
