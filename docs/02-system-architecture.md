# Autonomous Coding System - System Architecture

## Overview

Economic-first autonomous coding system that runs unattended for hours/days with complete traffic logging, financial tracking, and swappable intelligent machines. Configurable number of worker VMs.

## Core Principles

- **Economic-first design**: Every token, transaction, and resource tracked
- **Complete observability**: All traffic, inference, and spending logged
- **Machine autonomy**: Budget-aware intelligent machines making spending decisions
- **Maximum flexibility**: Swappable intelligent machines, payment methods, infrastructure

## Technology Stack

### Orchestrator (Docker + GPU)
- **vLLM** - Model serving for Qwen2.5-Coder + DeepSeek-R1 models
- **Ray Serve** - Load balancing across GPUs
- **Nomad** - Worker orchestration across VMs
- **HashiCorp Vault** - Secret management (cards, crypto keys, API tokens)
- **FastAPI** - Intelligent machine communication and control interface

### Workers (VMs, configurable specs)
- **Swappable intelligent machines**: Claude Code, Aider, Qwen Coder, Cursor CLI
- **SystemD services** - Process management
- **Network proxy routing** - All traffic through orchestrator
- **Budget-aware operation** - Daily spending limits via prompts

### Infrastructure Services
- **mitmproxy** - Transparent HTTPS interception + certificate generation
- **Elasticsearch** - Log storage
- **Filebeat** - Log shipping from workers
- **Kibana + Apache Superset** - Dashboards and visualization

## Model Configuration

### Model Pairs (Normal + Thinking)
- **32B**: Qwen2.5-Coder-32B-Instruct + DeepSeek-R1-32B
- **14B**: Qwen2.5-Coder-14B-Instruct + DeepSeek-R1-14B
- **9B**: Yi-Coder-9B-Chat + DeepSeek-R1-9B
- **7B**: Qwen2.5-Coder-7B-Instruct + DeepSeek-R1-7B
- **3B**: Qwen2.5-Coder-3B-Instruct + DeepSeek-R1-3B

### Selection Strategy
- **MVP**: Random assignment for configurable trials
- **Post-MVP**: ML-optimized selection based on task type and performance

## Economic System

### Budget Management
- **Project-level budgets** (configurable per trial)
- **Daily machine allowances** injected via prompts
- **Bi-weekly replenishment** tracking
- **Multi-currency support** (configurable primary currency, crypto preferences)

### Payment Strategy
1. **Direct card usage** via browser automation (no Stripe API)
2. **Cryptocurrency** preference (configurable, prompted)
3. **Other crypto** as fallback
4. **Free alternatives** always explored first

### Financial Tracking
- **Token cost monitoring** per model/request
- **Transaction logging** (successful + failed attempts)
- **Revenue tracking** from machine-generated income
- **Real-time dashboards** with profit/loss analysis

## Network Architecture

### Traffic Flow
All internet traffic flows through the mitmproxy service running on the orchestrator, which intercepts and logs all communications to/from the Worker VMs. This intercepted traffic data is then forwarded to Elasticsearch for storage and analysis.

### Security & Monitoring
- **Complete traffic interception** (HTTP/HTTPS)
- **Certificate pinning bypass** for mobile apps
- **Retention policies** configurable per field
- **Secret storage** in Vault for all sensitive data

## Data Storage Structure

Service-first organization with project subdirectories and date-based partitioning:

The data storage follows a service-first organization pattern with project subdirectories and date-based partitioning. The main data directory contains:

- **postgres/**: PostgreSQL database files from the docker compose stack
- **traffic-logs/**: mitmproxy traffic captures organized by project UUID and daily partitions for efficient filtering before Elasticsearch ingestion
- **financial/**: Transaction logs, beancount ledger files, and receipt storage organized by project
- **worker-artifacts/**: Generated code, applications, and other outputs from intelligent machines
- **checkpoints/**: VM snapshots, file manifests, and metadata for the evolution tracking system
- **elasticsearch/**: Standard Elasticsearch data directory for log storage
- **model-weights/**: Shared storage for AI model weights (Qwen-Coder, DeepSeek-R1, Yi-Coder)
- **vault-secrets/**: HashiCorp Vault encrypted storage for sensitive data
- **configs/**: Configuration files for Nomad jobs, retention policies, and log rotation

### Filesystem Retention Management

**Configurable retention policies with industry-standard tools**:

The retention configuration system uses a YAML-based policy file that defines different retention periods for various data types:

- **Traffic logs**: 14-day default retention (industry standard), reduced to 7 days for high-volume projects, 30 days for security events, with compression after 1 day
- **Worker artifacts**: Generated code retained for 30 days, build artifacts cleaned weekly, debug logs cleaned after 3 days, valuable outputs require manual review before deletion
- **Financial records**: 90-day retention for basic tracking, archived to cold storage after 30 days with daily backup frequency
- **Checkpoints**: Unlimited retention for active projects, failed snapshots cleaned after 7 days, LZ4 compression for frequent access, orphaned metadata cleanup within 24 hours

**Modern systemd-tmpfiles approach** (preferred over legacy tmpreaper):

The system uses modern systemd-tmpfiles configuration (preferred over legacy tmpreaper) to implement automated cleanup with differentiated retention periods:

- **Traffic logs**: Directory creation with configurable retention per UUID (14d standard, 7d for high volume projects, 30d for security events)
- **Worker artifacts**: Differentiated cleanup policies - code files retained 30 days, build artifacts 7 days, debug files 3 days, with exclusions for valuable outputs
- **Checkpoints**: Selective cleanup for failed checkpoints (7 days) and orphaned metadata (1 day), while preserving active project data
- **Temporary processing**: Fast cleanup of temporary processing directories (4 hours) to prevent disk space accumulation

**Dynamic configuration management**:

The dynamic configuration management system includes a bash script that reads the YAML configuration file and generates systemd-tmpfiles rules automatically. The script parses retention policies for traffic logs and other data types, generates appropriate cleanup rules with the correct paths and retention periods, then reloads the systemd-tmpfiles configuration to apply changes.

**Enterprise automation integration**:

Enterprise automation integration uses Ansible playbooks for retention management. The playbooks deploy templated retention configuration files with environment-specific variables (production environments use 2x retention multipliers, development uses 0.5x), set security log retention to 30 days, and automatically generate systemd-tmpfiles configuration through the retention manager script.

**Monitoring and alerting**:

The monitoring and alerting system includes a disk usage monitoring script with configurable thresholds that warns at 80% disk usage, triggers emergency cleanup procedures at 90% capacity, reports retention policy violations, and integrates with Prometheus/Grafana for real-time dashboard visualization and alerting.

**Research-based retention standards**:
- **Security logs**: 30d (basic operational needs)
- **Application logs**: 14d typical, 30-60d safer (industry standard)
- **Debug/trace logs**: 3-7d (high volume, low retention value)
- **Financial records**: 90d (basic financial tracking)
- **Build artifacts**: 7-30d (depending on CI/CD frequency)
- **Backup retention**: 3-2-1 rule (3 copies, 2 media types, 1 offsite)

## Intelligent Machine Design

### Core Capabilities
- **Budget awareness** via prompt injection
- **Payment decision-making** (buy vs free alternatives)
- **Multi-payment method support** (card, crypto, free)
- **Economic optimization** learning
- **Confidence-based execution** (>0.7 threshold using DeepEval)

### Swappable Intelligent Machine Support
- **Plugin architecture** for easy machine switching
- **Standardized interfaces** for file operations, web access, payments
- **Machine-agnostic logging** and monitoring
- **Performance comparison** across different intelligent machines

## Post-MVP Extensions

### Development Streaming
- **Terminal session streaming** to platforms
- **Desktop/browser recording** with real-time commentary
- **Multi-machine dashboard** streams
- **Financial overlay** showing live spending/earnings

### Advanced Controls
- **Microsoft Magentic-UI** integration for web automation
- **VNC/RDP access** to all worker VMs
- **Interactive task planning** with human oversight
- **Action guards** for sensitive operations

### VM Templates (Jinja2-based)
- **Automated VM provisioning**
- **Standard configurations** (64CPU/8GB/128GB Ubuntu LTS)
- **Template inheritance** and composition
- **Environment-specific customization**

### Enhanced Infrastructure
- **VLAN isolation** for network security
- **Advanced monitoring** with Grafana
- **ML-optimized scheduling** replacing random selection
- **High availability** orchestrator setup

## Success Metrics

### Technical
- **Uptime**: >99% worker availability
- **Throughput**: Successful task completion rate
- **Cost efficiency**: CAD per successful task
- **Response time**: Machine decision-making speed

### Economic
- **ROI tracking**: Revenue vs spending per machine
- **Budget adherence**: Staying within daily/project limits
- **Payment success rate**: Transaction completion
- **Cost optimization**: Trend toward efficiency

### Quality
- **Confidence scores**: SWE-bench style validation
- **Machine comparison**: Performance across different tools
- **Error rates**: Failed vs successful executions
- **Learning curves**: Improvement over time