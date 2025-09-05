# Technical Stack Design

## Infrastructure Overview

### Orchestrator (Host Machine)
- **Hardware**: 2x A5000 GPUs, sufficient CPU/RAM for model serving
- **OS**: Ubuntu LTS (containerized services)
- **Role**: Central control, model serving, secret management, traffic routing

### Workers (4 Virtual Machines)
- **Specs**: 64 CPU cores, 8GB RAM, 128GB disk space
- **OS**: Ubuntu LTS (latest)
- **Role**: Code execution, web browsing, task completion
- **Network**: All traffic routed through orchestrator

## Model Serving Stack

### vLLM Configuration
The vLLM configuration defines model serving parameters for different AI models. The Qwen 32B Coder model uses tensor parallelism across 2 GPUs with 85% memory utilization and specified model path. The DeepSeek R1 32B model follows similar configuration with 2-GPU tensor parallelism and high memory utilization. Additional models like Qwen 14B Coder can use single GPU allocation, while smaller models share GPU memory resources efficiently.

### Ray Serve Deployment
The Ray Serve deployment implements model load balancing across 2 A5000 GPUs using deployment decorators with 2 replicas and single GPU allocation per replica. The ModelService class initializes with a specified model name and creates a vLLM instance. The service provides an asynchronous generate method that accepts prompts and maximum token parameters, returning generated text responses through the underlying vLLM model.

## Orchestration Layer

### Nomad Configuration
The Nomad job configuration defines the autocode-worker service deployment across datacenter dc1 with service type scheduling. The agent group specifies 4 instances (one per VM) running coding-agent tasks with exec driver. Each task executes the worker binary with agent-type arguments from metadata variables. Environment variables include daily budget, project ID, and orchestrator URL configuration. Resource allocation provides 4 CPU cores and 2GB RAM per agent instance for optimal performance.

### SystemD Service Management
The SystemD service configuration manages the Autocode Worker Agent as a system service. The unit depends on network availability and runs as a simple service type under the autocode user. The service executes from the opt/autocode directory with configuration file parameter, automatic restart capability with 10-second intervals, and orchestrator URL environment variable. The service installs as a multi-user target dependency for system-wide availability.

## Network Architecture

### Traffic Routing
The network traffic flow routes from Worker VMs through iptables REDIRECT rules to the Orchestrator's mitmproxy service, which then forwards traffic to the Internet. All traffic is simultaneously logged to Elasticsearch for monitoring and analysis purposes, creating a comprehensive audit trail of all network activity.

### mitmproxy Configuration
The mitmproxy configuration includes a custom logging addon that captures all HTTP traffic details. The AutocodeLogger class initializes an Elasticsearch connection to localhost on port 9200. The request handler method creates comprehensive log entries including timestamps, worker identification from client IP, HTTP methods, URLs, headers, and content sizes. All traffic data is indexed into the autocode-traffic Elasticsearch index for searchable traffic analysis and monitoring.

### Certificate Management
Certificate management for HTTPS interception uses mitmproxy in transparent mode with host display enabled. The configuration specifies a custom certificate directory and certificate authority certificate path for proper SSL/TLS handling. The traffic logger script is automatically loaded to capture and process all intercepted traffic for analysis and logging purposes.

## Secret Management

### Vault Configuration
The Vault configuration uses file-based storage with data stored in the autocode vault-data directory. The TCP listener binds to all interfaces on port 8200 with TLS encryption using specified certificate and key files. API and cluster addresses point to the orchestrator hostname on ports 8200 and 8201 respectively, with the web UI enabled for management access.

### Secret Access Pattern
The secret access pattern uses the hvac library for Vault integration. The SecretManager class initializes a Vault client connection to the orchestrator with authentication via environment token. Payment card retrieval accesses the KV v2 secrets engine at the payments/card_primary path and returns the secret data. Crypto wallet retrieval dynamically constructs paths based on currency type (defaulting to TON) and accesses secrets from the crypto path namespace, returning wallet credentials for specified currencies.

## Logging & Monitoring

### Elasticsearch Configuration
The Elasticsearch configuration establishes the autocode-cluster with node name autocode-node-1. Data and logs are stored in dedicated autocode directories. The network configuration binds to all interfaces on port 9200 for broad accessibility. Security features are disabled for MVP simplicity, while monitoring collection is enabled to track cluster performance and health metrics.

### Log Retention Policies
The logging system uses Loguru for structured log management with dual output handlers. Console output provides INFO-level logging with timestamp, level, and message formatting. File output creates daily rotated logs in the autocode logs directory with enhanced formatting including extra context fields. Log files rotate daily with 30-day retention, gzip compression, and JSON serialization for efficient storage and analysis.

### Filebeat Configuration
The Filebeat configuration collects log files from the autocode logs directory with service and environment field tagging. Output streams to Elasticsearch on localhost with daily-indexed naming pattern including year, month, and day. Processing includes host metadata addition for logs not already tagged as forwarded, providing comprehensive log context and traceability.

## Agent Architecture

### Swappable Agent Interface
The swappable agent architecture uses an abstract base class defining common methods for all coding agents. The CodingAgent class initializes with configuration including daily budget and project ID, and defines abstract methods for task execution, spending reports, and budget checking. Specific implementations like ClaudeCodeAgent and AiderAgent inherit the base class and implement their respective integration logic. ClaudeCodeAgent configures the Claude binary path, while AiderAgent sets up the Aider binary path for their specific toolchain requirements.

### Agent Factory
The Agent Factory pattern provides centralized agent creation through a class method factory. The AGENTS dictionary maps agent type strings to their corresponding classes including claude-code, aider, qwen-coder, and cursor-cli implementations. The create_agent method validates the requested agent type against available options and instantiates the appropriate agent class with provided configuration, raising a ValueError for unknown agent types.

## Data Storage

### Directory Structure
The data storage organization uses a structured directory hierarchy under /mnt/data/autocode/. The elasticsearch directory contains nodes and indices subdirectories for search data. Vault-data provides secure backend storage for secrets management. Model-weights stores ML model files including qwen-2.5-coder-32b and deepseek-r1-32b models. The logs directory organizes worker logs, orchestrator logs, and traffic logs separately. Financial data includes transactions, budgets, and reports subdirectories. Worker-artifacts stores agent outputs including projects, generated code, and results. Configuration files are organized by service type including nomad, vault, and agents configurations.

## Performance Considerations

### GPU Memory Management
- **Model sharding** across 2xA5000 for large models
- **Dynamic loading** of smaller models based on demand
- **Memory monitoring** to prevent OOM errors
- **Graceful degradation** when GPU memory is full

### Network Optimization
- **Connection pooling** for external API calls
- **Caching** for frequently accessed resources
- **Compression** for log data transmission
- **Rate limiting** to prevent service overload

### Storage Optimization
- **Log rotation** with compression
- **Index lifecycle management** for Elasticsearch
- **Cleanup policies** for temporary files
- **Backup strategies** for critical data