# bcp

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/cowdogmoo/bcp)](https://goreportcard.com/report/github.com/cowdogmoo/bcp)
[![Pre-Commit](https://github.com/CowDogMoo/bcp/actions/workflows/pre-commit.yaml/badge.svg)](https://github.com/CowDogMoo/bcp/actions/workflows/pre-commit.yaml)
[![Tests](https://github.com/CowDogMoo/bcp/actions/workflows/tests.yaml/badge.svg)](https://github.com/CowDogMoo/bcp/actions/workflows/tests.yaml)
[![🚨 Semgrep Analysis](https://github.com/CowDogMoo/bcp/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/CowDogMoo/bcp/actions/workflows/semgrep.yaml)
[![Coverage Status](https://coveralls.io/repos/github/CowDogMoo/bcp/badge.svg?branch=main)](https://coveralls.io/github/CowDogMoo/bcp?branch=main)
[![Latest Release](https://img.shields.io/github/v/release/cowdogmoo/bcp)](https://github.com/CowDogMoo/bcp/releases/latest)
[![Renovate](https://github.com/CowDogMoo/bcp/actions/workflows/renovate.yaml/badge.svg)](https://github.com/CowDogMoo/bcp/actions/workflows/renovate.yaml)

`bcp` (Blob Copy) provides SCP-like functionality for copying files
to and from EC2 instances through S3 and AWS Systems Manager (SSM).

## Features

- **Bidirectional Transfer**: Copy files both TO and FROM EC2 instances
- **Resource Discovery**: List available S3 buckets and SSM-managed EC2 instances
- **Robust Error Handling**: Retry logic with exponential backoff for transient failures
- **Input Validation**: Comprehensive validation of paths, instance IDs,
  and bucket names
- **Structured Logging**: Configurable logging with multiple levels
  (debug, info, warn, error)
- **Configuration File Support**: YAML-based configuration with sensible defaults
- **Progress Reporting**: Clear progress messages during uploads and
  downloads
- **Automatic Cleanup**: S3 objects are automatically cleaned up after
  successful transfers

## Usage

### Copy Files

**Copy TO remote instance:**

```shell
bcp [local_source] [ssm_instance_id:remote_destination] --bucket BUCKET_NAME
```

**Copy FROM remote instance:**

```shell
bcp [ssm_instance_id:remote_source] [local_destination] --bucket BUCKET_NAME
```

**Arguments:**

- `local_source`: Local directory or file path to upload (when copying TO
  remote)
- `local_destination`: Local directory or file path to download to (when
  copying FROM remote)
- `ssm_instance_id:remote_path`: SSM instance ID and remote path (format:
  `i-xxxxxxxxx:/path/to/file`)

### List Resources

Discover available AWS resources before copying:

```shell
# List all S3 buckets
bcp list buckets

# List SSM-managed EC2 instances
bcp list instances

# List all EC2 instances (including non-SSM)
bcp list instances --all

# List instances in a specific region
bcp list instances --region us-west-2
```

### Global Flags

- `-b, --bucket`: S3 bucket name for transfer (required if not set in config)
- `-c, --config`: Path to config file (default: `$HOME/.bcp/config.yaml`)
- `-v, --verbose`: Enable verbose output (debug level)
- `-q, --quiet`: Suppress all output except errors
- `-h, --help`: Display help information

## Prerequisites

- **Local Machine**:
  - Go 1.20+ (for building from source)
  - AWS CLI installed and configured with appropriate credentials
  - IAM permissions for S3 and SSM

- **Remote Instance**:
  - AWS CLI installed
  - SSM Agent running
  - IAM instance profile with S3 read permissions

## Installation

### From Source

1. Clone the repository:

    ```shell
    git clone https://github.com/cowdogmoo/bcp.git
    cd bcp
    ```

2. Build the project:

    ```shell
    go build -o bcp
    ```

3. (Optional) Move to PATH:

    ```shell
    sudo mv bcp /usr/local/bin/
    ```

### Configuration

Create a configuration file at `$HOME/.bcp/config.yaml`:

```yaml
---
log:
  format: text  # text, json, or color
  level: info   # debug, info, warn, error

aws:
  region: us-east-1
  profile: default
  bucket: my-default-bucket  # Optional: default S3 bucket

transfer:
  max_retries: 3    # Maximum retry attempts
  retry_delay: 2    # Base delay in seconds (exponential backoff)
```

See `cmd/config/config.yaml` for a complete example.

### Shell Completion

Enable autocomplete for bucket names, instance IDs, and common paths:

**Bash:**

```bash
# One-time setup (Linux)
bcp completion bash | sudo tee /etc/bash_completion.d/bcp

# One-time setup (macOS with Homebrew)
bcp completion bash > $(brew --prefix)/etc/bash_completion.d/bcp

# Or source directly in current shell
source <(bcp completion bash)
```

**Zsh:**

```bash
# Enable completion system (if not already enabled)
echo "autoload -U compinit; compinit" >> ~/.zshrc

# One-time setup
bcp completion zsh > "${fpath[1]}/_bcp"

# Restart shell or reload
source ~/.zshrc
```

**Fish:**

```bash
# One-time setup
bcp completion fish > ~/.config/fish/completions/bcp.fish

# Reload completions
source ~/.config/fish/completions/bcp.fish
```

**Completion Features:**

- `bcp --bucket <TAB>` - Autocomplete S3 bucket names
- `bcp file.txt <TAB>` - Autocomplete SSM instance IDs
- `bcp file.txt i-xxx:<TAB>` - Suggest common destination paths
  (/tmp/, /home/ec2-user/, /opt/, etc.)

## Examples

### Discover Resources

```shell
# List available S3 buckets
bcp list buckets

# List SSM-managed instances (shows instance ID, status, platform, IP, and name)
bcp list instances

# List all instances including non-SSM ones
bcp list instances --all

# List instances in a specific region
bcp list instances --region us-east-1
```

### Basic File Transfer

```shell
# Copy a directory TO an EC2 instance
bcp ./my-files i-1234567890abcdef0:/home/ec2-user/files --bucket my-bucket

# Copy a single file TO an EC2 instance
bcp ~/app/binary i-1234567890abcdef0:/usr/local/bin/myapp --bucket my-bucket

# Copy a directory FROM an EC2 instance
bcp i-1234567890abcdef0:/home/ec2-user/files ./my-files --bucket my-bucket

# Copy a single file FROM an EC2 instance
bcp i-1234567890abcdef0:/var/log/app.log ./logs/app.log --bucket my-bucket
```

### With Verbose Logging

```shell
# Enable debug logging to see detailed progress
bcp ./my-files i-1234567890abcdef0:/home/ec2-user/files --bucket my-bucket --verbose
```

### Using Configuration File

```shell
# Use a custom config file
bcp ./my-files i-1234567890abcdef0:/home/ec2-user/files --config ./my-config.yaml
```

### Quiet Mode

```shell
# Only show errors
bcp ./my-files i-1234567890abcdef0:/home/ec2-user/files --bucket my-bucket --quiet
```

### Workflow Example

```shell
# 1. Discover available buckets and instances
bcp list buckets
bcp list instances

# 2. Copy files TO remote using discovered resources
bcp ./deployment i-0080bcec99ef6fbf2:/opt/app --bucket dread-infra-alpha-operator-range-dev-us-west-2

# 3. Copy files FROM remote
bcp i-0080bcec99ef6fbf2:/opt/app/logs ./logs --bucket dread-infra-alpha-operator-range-dev-us-west-2
```

## Development

### Running Tests

```shell
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests verbosely
go test ./... -v
```

### Building

```shell
# Build for current platform
go build -o bcp

# Build with version info
go build -ldflags "-X main.Version=1.0.0" -o bcp
```

## Troubleshooting

### Common Issues

1. **"bucket name is required"**: Set bucket via `--bucket` flag or in config file
2. **"invalid SSM instance ID"**: Ensure instance ID format is `i-xxxxxxxxx`
3. **"AWS CLI is not installed on instance"**: Install AWS CLI on the remote instance
4. **"operation failed after N retries"**: Check network connectivity and AWS credentials

### Enable Debug Logging

```shell
bcp ./files i-xxx:/tmp/files --bucket my-bucket --verbose
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request
