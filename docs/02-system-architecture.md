# Autonomous Coding System - System Architecture

## Overview

Economic-first autonomous coding system that runs unattended for hours/days with complete traffic logging, financial tracking, and swappable coding agents. Configurable number of worker VMs.

## Core Principles

- **Economic-first design**: Every token, transaction, and resource tracked
- **Complete observability**: All traffic, inference, and spending logged
- **Agent autonomy**: Budget-aware agents making spending decisions
- **Maximum flexibility**: Swappable agents, payment methods, infrastructure

## Technology Stack

### Orchestrator (Docker + GPU)
- **vLLM** - Model serving for Qwen2.5-Coder + DeepSeek-R1 models
- **Ray Serve** - Load balancing across GPUs
- **Nomad** - Worker orchestration across VMs
- **HashiCorp Vault** - Secret management (cards, crypto keys, API tokens)
- **FastAPI** - Agent communication and control interface

### Workers (VMs, configurable specs)
- **Swappable coding agents**: Claude Code, Aider, Qwen Coder, Cursor CLI
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
- **Daily agent allowances** injected via prompts
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
- **Revenue tracking** from agent-generated income
- **Real-time dashboards** with profit/loss analysis

## Network Architecture

### Traffic Flow
```
Internet ↔ mitmproxy (orchestrator) ↔ Worker VMs
              ↓
         Elasticsearch Logging
```

### Security & Monitoring
- **Complete traffic interception** (HTTP/HTTPS)
- **Certificate pinning bypass** for mobile apps
- **Retention policies** configurable per field
- **Secret storage** in Vault for all sensitive data

## Data Storage Structure

```
/mnt/data/autocode/
├── projects/
│   ├── project-1/
│   │   ├── financial/             # Project-specific transaction logs
│   │   ├── traffic-logs/          # Project-specific mitmproxy captures
│   │   ├── worker-artifacts/      # Generated code, apps, outputs
│   │   └── checkpoints/          # VM snapshots and evolution tree
│   ├── project-2/
│   │   └── ...
├── shared/
│   ├── elasticsearch/          # Centralized ES data directory
│   ├── model-weights/         # Qwen2.5-Coder + DeepSeek-R1 weights
│   ├── vault-secrets/         # Encrypted key storage
│   ├── configs/              # Nomad jobs, retention policies
│   └── postgres/             # Future PostgreSQL data
```

## Agent Design

### Core Capabilities
- **Budget awareness** via prompt injection
- **Payment decision-making** (buy vs free alternatives)
- **Multi-payment method support** (card, crypto, free)
- **Economic optimization** learning
- **Confidence-based execution** (>0.7 threshold using DeepEval)

### Swappable Agent Support
- **Plugin architecture** for easy agent switching
- **Standardized interfaces** for file operations, web access, payments
- **Agent-agnostic logging** and monitoring
- **Performance comparison** across different agents

## Post-MVP Extensions

### Development Streaming
- **Terminal session streaming** to platforms
- **Desktop/browser recording** with real-time commentary
- **Multi-agent dashboard** streams
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
- **Response time**: Agent decision-making speed

### Economic
- **ROI tracking**: Revenue vs spending per agent
- **Budget adherence**: Staying within daily/project limits
- **Payment success rate**: Transaction completion
- **Cost optimization**: Trend toward efficiency

### Quality
- **Confidence scores**: SWE-bench style validation
- **Agent comparison**: Performance across different tools
- **Error rates**: Failed vs successful executions
- **Learning curves**: Improvement over time