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

Service-first organization with project subdirectories and date-based partitioning:

```
/mnt/data/autovibe/
├── postgres/                   # PostgreSQL data from docker compose
│   └── data/                   # Standard postgres data directory
├── traffic-logs/               # mitmproxy captures with date partitioning
│   ├── project-0-autovibe/
│   │   ├── 2025-01-01/        # Daily partitions for easy filtering
│   │   ├── 2025-01-02/        # Before loading to Elasticsearch
│   │   └── 2025-01-03/
│   ├── project-2-crypto-profit/
│   │   ├── 2025-01-01/
│   │   └── 2025-01-02/
│   └── project-4-cryptopulse/
├── financial/                  # Transaction logs and beancount ledgers
│   ├── project-0-autovibe/
│   │   ├── transactions.json
│   │   ├── ledger.beancount
│   │   └── receipts/
│   ├── project-2-crypto-profit/
│   └── project-4-cryptopulse/
├── worker-artifacts/           # Generated code, apps, outputs
│   ├── project-0-autovibe/
│   ├── project-2-crypto-profit/
│   └── project-4-cryptopulse/
├── checkpoints/                # VM snapshots and evolution tree
│   ├── project-0-autovibe/
│   │   ├── vm-snapshots/
│   │   ├── file-manifests/
│   │   └── metadata/
│   ├── project-2-crypto-profit/
│   └── project-4-cryptopulse/
├── elasticsearch/              # ES data directory
│   └── data/                   # Standard elasticsearch data
├── model-weights/              # Shared AI model weights
│   ├── qwen-coder/
│   ├── deepseek-r1/
│   └── yi-coder/
├── vault-secrets/              # HashiCorp Vault encrypted storage
└── configs/                    # Nomad jobs, retention policies
    ├── nomad/
    ├── retention-policies.yaml
    └── logrotate.conf
```

### Filesystem Retention Management

**Directory-based retention with automatic cleanup**:

```bash
# Filesystem retention using tmpreaper, find, and systemd timers
/mnt/data/autovibe/
├── traffic-logs/               # Auto-cleanup with tmpreaper
│   ├── project-*/2025-*-*/     # tmpreaper removes dirs older than 30d
├── financial/                  # Manual retention (7y legal requirement)
├── worker-artifacts/           # find + cron cleanup (90d retention)
├── checkpoints/                # Unlimited retention with LZ4 compression
```

**Retention Implementation (filesystem level)**:

1. **tmpreaper** - Efficient directory tree cleanup
```bash
# /etc/tmpreaper.conf  
TMPREAPER_TIME=30d
TMPREAPER_DIRS="/mnt/data/autovibe/traffic-logs"
TMPREAPER_PROTECT_EXTRA='*.lock'
```

2. **logrotate** - File-based rotation
```bash
# /etc/logrotate.d/autovibe-traffic
/mnt/data/autovibe/traffic-logs/*/2025-*-*/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    sharedscripts
}
```

3. **systemd timer** - Custom cleanup service
```ini
# /etc/systemd/system/autovibe-cleanup.timer
[Unit]
Description=AutoVibe filesystem cleanup
Requires=autovibe-cleanup.service

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
```

4. **find + xargs** - Bulk operations
```bash
# Remove worker artifacts older than 90 days
find /mnt/data/autovibe/worker-artifacts -type f -mtime +90 -print0 | xargs -0 rm -f

# Compress checkpoints older than 7 days
find /mnt/data/autovibe/checkpoints -name "*.qcow2" -mtime +7 -exec lz4 {} \;
```

**Common filesystem retention tools**:
- **tmpreaper**: Removes files/dirs by age (Debian/Ubuntu standard)
- **tmpwatch**: RedHat/CentOS equivalent  
- **bleachbit**: Cross-platform system cleaner
- **ncdu**: Disk usage analyzer for manual cleanup
- **duf**: Modern disk usage utility

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