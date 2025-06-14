# t - Task Runner

A flexible task runner with template support and action-based workflow.

## Features

- **Template Variables**: Access command line arguments, environment variables, and action outputs
- **Action-based Workflow**: Modular actions for file operations, command execution, and more
- **Configuration-driven**: YAML-based task definitions

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

```yaml
defaults:
  llm:
    provider: "openai"
    model: "gpt-4"
    temperature: 0.3

tasks:
  example:
    steps:
      - id: create_file
        action: file.temp
        args:
          content: "Hello {{ .arg0 }}!"

      - id: show_file_path
        action: stdout.write
        args:
          content: "Created file: {{ .output }}"

      - id: show_specific_output
        action: stdout.write
        args:
          content: "File from create_file: {{ .output.create_file }}"

      - id: show_env
        action: stdout.write
        args:
          content: "User: {{ .env.USER }}"
```

Run with:
```bash
go run . -c config.yaml example "World"
```

## Template Syntax Summary

**Recommended Syntax:**
- `{{ .output }}`: Output from the previous step
- `{{ .output.action_id }}`: Output from a specific action (when action has an ID)
- `{{ .arg0 }}`, `{{ .arg1 }}`, `{{ .arg2 }}`: Command line arguments
- `{{ .env.VAR_NAME }}`: Environment variables
- `{{ .data.key }}`: Metadata

**Alternative Syntax:**
- `{{ index .args 0 }}`: Alternative argument access
- `{{ index .args 1 }}`: Second argument, etc.

## Actions

### file.temp
Creates a temporary file with specified content.

**Parameters:**
- `content` (string): File content (supports templates)

**Output:** Path to the created temporary file

### stdout.write
Writes content to standard output.

**Parameters:**
- `content` (string): Content to write (supports templates)

**Output:** The written content

### command.run
Executes a shell command.

**Parameters:**
- `command` (string): Command to execute (supports templates)
- `args` ([]string): Command arguments (supports templates)

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
        args:
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
        args:
          key: value
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
        args:
          param1: value1
          param2: value2
      - action: another-action
        # args are optional
```

## Available Actions

### 1. File Operations

#### file.read - Read File

```yaml
- action: file.read
  args:
    path: "/path/to/file.txt"
```

#### file.write - Write File

```yaml
- action: file.write
  args:
    path: "/path/to/output.txt"
    content: "Content to write"
    mode: 0644  # File permissions (optional)
```

#### file.temp - Create Temporary File

```yaml
- action: file.temp
  args:
    content: "Temporary file content"
    prefix: "temp-"     # Filename prefix
    suffix: ".txt"      # Filename suffix
    dir: "/tmp"         # Creation directory (optional)
```

#### file.delete - Delete File

```yaml
- action: file.delete
  args:
    path: "/path/to/file.txt"
```

### 2. Standard Output

#### stdout.write - Write to Standard Output

```yaml
- action: stdout.write
  args:
    content: "Content to display"
```

When args are omitted, displays the output from the previous step:

```yaml
- action: stdout.write
```

### 3. Command Execution

#### command.run - Execute System Command

```yaml
- action: command.run
  args:
    cmd: "echo 'Hello World'"
    dir: "/path/to/workdir"  # Working directory (optional)
```

### 4. Clipboard Operations

#### clipboard.read - Read from Clipboard

```yaml
- action: clipboard.read
```

#### clipboard.write - Write to Clipboard

```yaml
- action: clipboard.write
  args:
    content: "Content to write to clipboard"
```

### 5. LLM Operations

#### llm.generate - Generate Text

```yaml
- action: llm.generate
  args:
    system: "You are a helpful assistant."
    prompt: "Say hello in Japanese"
    model: gpt-4         # Override default settings
    temperature: 0.7     # Override default settings
```

#### llm.session - Session Management

```yaml
- action: llm.session
  args:
    session_id: "my-session"
    system: "System prompt for continuous conversation"
    prompt: "First message"
    model: gpt-4         # Override default settings (optional)
```

## Practical Usage Examples

### 1. Basic Hello World

```yaml
tasks:
  hello:
    steps:
      - action: stdout.write
        args:
          content: "Hello, World!"
```

### 2. File Processing Pipeline

```yaml
tasks:
  file-processing:
    steps:
      - action: file.read
        args:
          path: "input.txt"
      - action: llm.generate
        args:
          system: "Please summarize the text."
          prompt: "Please summarize the following text:\n{{ .output }}"
      - action: file.write
        args:
          path: "summary.txt"
          content: "{{ .output }}"
      - action: stdout.write
        args:
          content: "Summary saved to summary.txt"
```

### 3. Clipboard and LLM Integration

```yaml
tasks:
  clipboard-translate:
    steps:
      - action: clipboard.read
      - action: llm.generate
        args:
          system: "You are a translator."
          prompt: "Please translate the following text to Japanese:\n{{ .output }}"
      - action: clipboard.write
        args:
          content: "{{ .output }}"
      - action: stdout.write
        args:
          content: "Translation result copied to clipboard"
```

### 4. Command Execution and File Saving

```yaml
tasks:
  system-info:
    steps:
      - action: command.run
        args:
          cmd: "uname -a && date && whoami"
      - action: file.write
        args:
          path: "system-info.txt"
          content: "{{ .output }}"
      - action: stdout.write
        args:
          content: "System information saved to system-info.txt"
```

### 5. Processing with Temporary Files

```yaml
tasks:
  temp-processing:
    steps:
      - id: create_temp
        action: file.temp
        args:
          content: "Data to process"
          prefix: "process-"
          suffix: ".txt"
      - action: command.run
        args:
          cmd: "wc -l {{ .output }}"
      - action: stdout.write
        args:
          content: "Line count: {{ .output }}"
      - action: file.delete
        args:
          path: "{{ .output.create_temp }}"

  read-file-from-args:
    steps:
      - action: file.read
        args:
          path: "{{ .arg0 }}"
      - action: stdout.write
        args:
          content: "File {{ .arg0 }} contains:\n{{ .output }}"
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
