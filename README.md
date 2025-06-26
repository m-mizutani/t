# t - Task Runner with LLM

A flexible task runner that integrates Large Language Models with automation workflows. Build pipelines that combine AI text processing with file operations, command execution, and system automation.

## Features

- **LLM Integration**: Built-in support for OpenAI GPT, Anthropic Claude, and Google Gemini
- **AI Workflows**: Chain LLM responses with file operations and command execution
- **Template Variables**: Access command line arguments, environment variables, and action outputs
- **Action-based Architecture**: Modular actions for file operations, clipboard management, and more
- **Configuration-driven**: YAML-based task definitions with consistent parameter syntax
- **Session Management**: Maintain conversation context across multiple LLM interactions
- **Type Safety**: Built-in parameter validation at configuration load time
- **Unified Syntax**: All action parameters use consistent top-level YAML structure

## Quick Start

Create workflows that combine AI capabilities with system automation:

```yaml
defaults:
  llm:
    provider: "openai"
    model: "gpt-4"
    temperature: 0.3

tasks:
  ai-summary:
    steps:
      - action: file.read
        path: "document.txt"

      - action: llm.generate
        system: "You are a skilled summarizer."
        prompt: "Please summarize this document:\n{{ .output }}"

      - action: file.write
        path: "summary.txt"
        content: "{{ .output }}"

      - action: stdout.write
        content: "Summary saved to summary.txt"
```

Run with:
```bash
t ai-summary
```

## Template Variables

The following template variables are available in action parameters:

### Arguments
- `{{ .arg0 }}`, `{{ .arg1 }}`, `{{ .arg2 }}`, etc. - Command line arguments (zero-indexed)
- Alternative: `{{ index .args 0 }}`, `{{ index .args 1 }}`, etc.

### Outputs
- `{{ .output }}` - Output from the previous step
- `{{ .output.action_id }}` - Output from a specific action by its ID

### Environment Variables
- `{{ .env.VAR_NAME }}` - Environment variable value

### Metadata
- `{{ .data.key }}` - Metadata value

## Example Usage

### Clipboard Processing
Transform clipboard content with AI:

```yaml
tasks:
  clipboard-enhance:
    steps:
      - action: clipboard.read
      
      - action: llm.generate
        system: "You are a professional writer. Improve clarity and readability."
        prompt: "Please enhance this text:\n{{ .output }}"
      
      - action: clipboard.write
        content: "{{ .output }}"
      
      - action: stdout.write
        content: "Enhanced text copied to clipboard"
```

### Document Translation
Translate documentation files:

```yaml
tasks:
  translate-docs:
    steps:
      - action: file.read
        path: "{{ .arg0 }}"
      
      - action: llm.generate
        system: "You are a professional translator."
        prompt: "Translate this to {{ .arg1 }}:\n{{ .output }}"
      
      - action: file.write
        path: "{{ .arg0 }}.{{ .arg1 }}"
        content: "{{ .output }}"
```

Run with:
```bash
t translate-docs README.md japanese
```

### Interactive AI Sessions
Maintain conversation context:

```yaml
tasks:
  chat:
    steps:
      - action: llm.session
        session_id: "coding-assistant"
        system: "You are a helpful coding assistant."
        prompt: "{{ .arg0 }}"
      
      - action: stdout.write
        content: "{{ .output }}"
```

## Template Syntax Summary

**Recommended Syntax:**
- `{{ .output }}`: Output from the previous step
- `{{ .output.action_id }}`: Output from a specific action (when action has an ID)
- `{{ .arg0 }}`, `{{ .arg1 }}`, `{{ .arg2 }}`: Command line arguments
- `{{ .env.VAR_NAME }}`: Environment variables
- `{{ .data.key }}`: Metadata

## Actions

### file.temp
Creates a temporary file with specified content.

**Parameters:**
- `id` (optional): Step identifier for referencing output
- `content` (optional): File content (supports templates)

**Example:**
```yaml
- action: file.temp
  id: temp-file
  content: "Temporary content"
```

**Output:** Path to the created temporary file

---

### file.delete
Deletes a file from the filesystem.

**Parameters:**
- `id` (optional): Step identifier for referencing output
- `path` (required): Path to the file to delete

**Example:**
```yaml
- action: file.delete
  path: "/tmp/unwanted-file.txt"
```

**Output:** Path of the deleted file

### stdout.write
Writes content to standard output.

**Parameters:**
- `id` (optional): Step identifier for referencing output
- `content` (optional): Content to write (supports templates)
- `no_newline` (optional): If true, don't add automatic newline (default: false)

**Output:** The written content

### command.run
Executes a shell command.

**Parameters:**
- `id` (optional): Step identifier for referencing output
- `cmd` (required): Command to execute (supports templates)
- `dir` (optional): Working directory (supports templates)

**Output:** Command output (stdout)

## Configuration

Create a configuration file (default: `~/.config/t/task.yaml`):

```yaml
defaults:
  llm:
    provider: "openai"
    model: "gpt-4"
    temperature: 0.3

tasks:
  my_task:
    steps:
      - action: stdout.write
        content: "Hello {{ .arg0 }}!"
```

## Installation

### From Binary

```bash
# Download from releases page (coming soon)
```

### From Source

```bash
git clone https://github.com/m-mizutani/t.git
cd t
go build -o t .
```

## Basic Usage

```bash
# Run a task
t <task-name>

# Use custom configuration file
t --config custom.yaml <task-name>

# Pass arguments
t echo "Hello World"

# Show help
t --help
```

## Configuration File (task.yaml) Syntax

### Basic Structure

```yaml
# Default settings
defaults:
  llm:
    provider: openai    # LLM provider
    model: gpt-4       # Model to use
    temperature: 0.3   # Generation creativity

# Task definitions
tasks:
  task-name:
    steps:
      - action: action-name
        param1: value1
        param2: value2
```

### Default Settings (defaults)

#### LLM Configuration

```yaml
defaults:
  llm:
    provider: openai           # LLM provider (required)
    model: gpt-4              # Model name (required)
    temperature: 0.3          # Generation creativity (0.0-1.0)
    max_tokens: 1000          # Maximum output tokens
    api_key: ""              # API key (use environment variables)
```

**Supported Providers:**

| Provider | Description | Environment Variables |
|----------|-------------|---------------------|
| `openai` | OpenAI GPT models (GPT-4, GPT-3.5, etc.) | `OPENAI_API_KEY` |
| `claude` | Anthropic Claude models (Claude 3, Claude 3.5, etc.) | `ANTHROPIC_API_KEY` |
| `gemini` | Google Vertex AI models (Gemini, PaLM, etc.) | `GOOGLE_CLOUD_PROJECT`, `GOOGLE_CLOUD_LOCATION` (optional) |

**Available Configuration Options:**

| Option | Type | Default | Range/Values | Description |
|--------|------|---------|--------------|-------------|
| `provider` | string | `openai` | `openai`, `claude`, `gemini` | LLM provider to use |
| `model` | string | `gpt-4` | Provider-specific | Model name (see URLs below) |
| `temperature` | float64 | `0.3` | 0.0-1.0 | Controls randomness (0.0=deterministic, 1.0=creative) |
| `max_tokens` | int | `1000` | 1-∞ | Maximum number of tokens to generate |
| `api_key` | string | `""` | - | API key (use environment variables) |

**Note**: Additional options like `top_p`, `top_k`, `frequency_penalty`, `presence_penalty`, and `timeout` are defined in config but currently used for documentation purposes. The actual implementation uses the core options above.

**Model Lists by Provider:**

- **OpenAI Models**: https://platform.openai.com/docs/models
- **Claude Models**: https://docs.anthropic.com/en/docs/about-claude/models
- **Gemini Models**: https://cloud.google.com/vertex-ai/generative-ai/docs/models

**Popular Models by Provider:**

- **OpenAI**: `gpt-4o`, `gpt-4.5`, `o1`, `o3-mini`, `gpt-4.1`
- **Claude**: `claude-opus-4`, `claude-sonnet-4`, `claude-3.7-sonnet`, `claude-3.5-haiku`  
- **Gemini**: `gemini-2.5-pro`, `gemini-2.5-flash`, `gemini-2.0-flash`, `gemini-2.0-flash-lite`

### Task Definitions (tasks)

Each task is defined as an array of `steps`. Steps are executed sequentially, with the output of the previous step passed as input to the next step.

```yaml
tasks:
  my-task:
    steps:
      - action: action-name
        param1: value1
        param2: value2
      - action: another-action
        # parameters are at top level
```

## Available Actions

### 1. File Operations

#### file.read - Read File

Read content from a file.

**Parameters:**
- `id` (optional): Step identifier for referencing output
- `path` (required): Path to the file to read

**Example:**
```yaml
- action: file.read
  id: read-config
  path: "/path/to/file.txt"
```

**Output:** File content as string

---

#### file.write - Write File

Write content to a file. Creates parent directories if they don't exist.

**Parameters:**
- `id` (optional): Step identifier for referencing output
- `path` (required): Path to the output file
- `content` (optional): Content to write (uses input from previous step if not specified)
- `mode` (optional): File permissions (e.g., "0644", defaults to 0644)

**Example:**
```yaml
- action: file.write
  id: save-result
  path: "/path/to/output.txt"
  content: "Content to write"
  mode: "0644"
```

**Output:** Written file path

---

### 2. Standard Output

#### stdout.write - Write to Standard Output

Write content to stdout with optional newline control.

**Parameters:**
- `id` (optional): Step identifier for referencing output
- `content` (optional): Content to display (uses input from previous step if not specified)
- `no_newline` (optional): If true, don't add automatic newline (default: false)

**Example:**
```yaml
- action: stdout.write
  content: "Content to display"
  no_newline: false
```

When content is omitted, displays the output from the previous step:

```yaml
- action: stdout.write
```

**Output:** Written content as string

---

### 3. Command Execution

#### command.run - Execute System Command

Execute a shell command with optional working directory.

**Parameters:**
- `id` (optional): Step identifier for referencing output
- `cmd` (required): Shell command to execute
- `dir` (optional): Working directory for command execution

**Example:**
```yaml
- action: command.run
  id: list-files
  cmd: "echo 'Hello World'"
  dir: "/path/to/workdir"
```

**Output:** Command stdout as string

**Metadata:**
- `command`: Executed command
- `workdir`: Working directory
- `stdout`: Command stdout
- `stderr`: Command stderr
- `success`: Whether command succeeded
- `exit_code`: Exit code (if failed)

---

### 4. Clipboard Operations

#### clipboard.read - Read from Clipboard

Read content from the system clipboard.

**Parameters:**
- `id` (optional): Step identifier for referencing output

**Example:**
```yaml
- action: clipboard.read
  id: clipboard-content
```

**Output:** Clipboard content as string

---

#### clipboard.write - Write to Clipboard

Write content to the system clipboard.

**Parameters:**
- `id` (optional): Step identifier for referencing output
- `content` (optional): Content to write to clipboard (uses input from previous step if not specified)

**Example:**
```yaml
- action: clipboard.write
  content: "Content to write to clipboard"
```

**Output:** Written content as string

### 5. LLM Operations

#### llm.generate - Generate Text

Generate text using a Large Language Model.

**Parameters:**
- `id` (optional): Step identifier for referencing output
- `system` (optional): System prompt to set context/role
- `prompt` (required): Text prompt for generation
- `model` (optional): Override default model
- `temperature` (optional): Override default temperature
- `max_tokens` (optional): Override default max tokens

**Example:**
```yaml
- action: llm.generate
  system: "You are a helpful assistant."
  prompt: "Say hello in Japanese"
  model: gpt-4         # Override default settings
  temperature: 0.7     # Override default settings
```

**Output:** Generated text as string

#### llm.session - Session Management

Manage persistent conversation sessions with LLMs.

**Parameters:**
- `id` (optional): Step identifier for referencing output
- `session_id` (required): Unique identifier for the session
- `system` (optional): System prompt for the session
- `prompt` (required): Message to send in the session
- `model` (optional): Override default model
- `temperature` (optional): Override default temperature

**Example:**
```yaml
- action: llm.session
  session_id: "my-session"
  system: "System prompt for continuous conversation"
  prompt: "First message"
  model: gpt-4         # Override default settings (optional)
```

**Output:** LLM response as string

## Practical Usage Examples

### 1. Basic Hello World

```yaml
tasks:
  hello:
    steps:
      - action: stdout.write
        content: "Hello, World!"
```

### 2. File Processing Pipeline

```yaml
tasks:
  file-processing:
    steps:
      - action: file.read
        id: input-content
        path: "input.txt"
      - action: file.write
        id: save-processed
        path: "output.txt"
        content: "Processed: {{ .output }}"
      - action: stdout.write
        content: "File processed and saved to output.txt"
```

### 3. Command Pipeline

```yaml
tasks:
  system-info:
    steps:
      - action: command.run
        id: get-info
        cmd: "uname -a && date && whoami"
      - action: file.write
        path: "system-info.txt"
        content: "{{ .output }}"
      - action: stdout.write
        content: "System information saved to system-info.txt"
```

### 4. Multi-step File Operations

```yaml
tasks:
  file-chain:
    steps:
      - action: file.write
        id: create-temp
        path: "/tmp/example.txt"
        content: "Hello from typed actions!"
      
      - action: file.read
        id: read-temp
        path: "/tmp/example.txt"
      
      - action: stdout.write
        content: "File content: {{ .output }}"
      
      - action: command.run
        cmd: "rm -f /tmp/example.txt"
```

### 5. Template Usage with Action References

```yaml
tasks:
  template-example:
    steps:
      - action: command.run
        id: current-date
        cmd: "date '+%Y-%m-%d'"
      
      - action: file.write
        path: "log-{{ .output_of.current-date }}.txt"
        content: "Log created on {{ .output_of.current-date }}"
      
      - action: stdout.write
        content: "Log file created with date: {{ .output_of.current-date }}"
```

## Configuration File Locations

Default configuration file locations:

- macOS/Linux: `~/.config/t/task.yaml`
- Windows: `%APPDATA%\t\task.yaml`

Specify custom file:

```bash
t --config /path/to/custom.yaml task-name
```

## Environment Variables

Manage sensitive information like LLM API keys using environment variables:

```bash
export OPENAI_API_KEY="your-api-key"
export ANTHROPIC_API_KEY="your-api-key"
export GOOGLE_CLOUD_PROJECT="your-project-id"
export GOOGLE_CLOUD_LOCATION="us-central1"  # Optional, defaults to us-central1
```

## Troubleshooting

### Common Issues

1. **LLM API key not configured**
   ```
   Error: API key not found
   ```
   → Set the API key using environment variables

2. **File not found**
   ```
   Error: file not found
   ```
   → Verify that the file path is correct

3. **Clipboard access error**
   ```
   Error: failed to initialize clipboard
   ```
   → Ensure you're running in a GUI environment

### Debugging

Display detailed logs:

```bash
t --log-level debug task-name
```

## License

Apache License 2.0

## Contributing

Pull requests and issue reports are welcome.

## Related Links

- [GitHub Repository](https://github.com/m-mizutani/t)
- [Issues](https://github.com/m-mizutani/t/issues)
