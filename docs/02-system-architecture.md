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
│   ├── 123e4567-e89b-12d3-a456-426614174000/
│   │   ├── 2025-01-01/        # Daily partitions for easy filtering
│   │   ├── 2025-01-02/        # Before loading to Elasticsearch
│   │   └── 2025-01-03/
│   ├── 987fcdeb-51c2-43d1-a983-123456789abc/
│   │   ├── 2025-01-01/
│   │   └── 2025-01-02/
│   └── 456e7890-a12b-34c5-d678-901234567890/
├── financial/                  # Transaction logs and beancount ledgers
│   ├── 123e4567-e89b-12d3-a456-426614174000/
│   │   ├── transactions.json
│   │   ├── ledger.beancount
│   │   └── receipts/
│   ├── 987fcdeb-51c2-43d1-a983-123456789abc/
│   └── 456e7890-a12b-34c5-d678-901234567890/
├── worker-artifacts/           # Generated code, apps, outputs
│   ├── 123e4567-e89b-12d3-a456-426614174000/
│   ├── 987fcdeb-51c2-43d1-a983-123456789abc/
│   └── 456e7890-a12b-34c5-d678-901234567890/
├── checkpoints/                # VM snapshots and evolution tree
│   ├── 123e4567-e89b-12d3-a456-426614174000/
│   │   ├── vm-snapshots/
│   │   ├── file-manifests/
│   │   └── metadata/
│   ├── 987fcdeb-51c2-43d1-a983-123456789abc/
│   └── 456e7890-a12b-34c5-d678-901234567890/
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

**Configurable retention policies with industry-standard tools**:

```yaml
# /etc/autovibe/retention-config.yaml
retention_policies:
  traffic_logs:
    default_retention: "14d"      # Industry standard: 2 weeks typical, 1-2 months safer
    high_volume_projects: "7d"    # Reduce for high-traffic projects
    security_events: "30d"        # Security logs basic retention
    compression_after: "1d"       # Compress after 1 day
    
  worker_artifacts:
    code_outputs: "30d"           # Generated code kept 1 month
    build_artifacts: "7d"         # Build outputs cleaned weekly  
    debug_logs: "3d"              # Debug info cleaned quickly
    valuable_outputs: "manual"    # Require manual review before deletion
    
  financial_records:
    retention: "90d"              # Basic financial tracking
    archive_after: "30d"         # Move to cold storage after 30 days  
    backup_frequency: "daily"
    
  checkpoints:
    active_projects: "unlimited"  # Keep all project evolution
    failed_snapshots: "7d"       # Clean failed/corrupted quickly
    compression: "lz4"            # Fast compression for frequent access
    cleanup_orphaned: "24h"      # Remove orphaned metadata
```

**Modern systemd-tmpfiles approach** (preferred over legacy tmpreaper):

```ini
# /etc/tmpfiles.d/autovibe.conf
# Type Path                    Mode User Group Age      Argument

# Traffic logs with configurable retention per UUID
d /mnt/data/autovibe/traffic-logs    0755 autovibe autovibe -
d /mnt/data/autovibe/traffic-logs/123e4567-e89b-12d3-a456-426614174000 0755 autovibe autovibe -
Z /mnt/data/autovibe/traffic-logs/123e4567-e89b-12d3-a456-426614174000 - autovibe autovibe 14d  # Standard retention
Z /mnt/data/autovibe/traffic-logs/987fcdeb-51c2-43d1-a983-123456789abc - autovibe autovibe 7d   # High volume = shorter
e /mnt/data/autovibe/traffic-logs/*security* - - - 30d  # Security events basic retention

# Worker artifacts with differentiated cleanup
Z /mnt/data/autovibe/worker-artifacts/*/code 0644 autovibe autovibe 30d
Z /mnt/data/autovibe/worker-artifacts/*/builds 0644 autovibe autovibe 7d  
Z /mnt/data/autovibe/worker-artifacts/*/debug 0644 autovibe autovibe 3d
x /mnt/data/autovibe/worker-artifacts/*/valuable*  # Exclude valuable outputs

# Checkpoints with selective cleanup  
d /mnt/data/autovibe/checkpoints     0755 autovibe autovibe -
Z /mnt/data/autovibe/checkpoints/*/failed 0644 autovibe autovibe 7d
Z /mnt/data/autovibe/checkpoints/*/orphaned 0644 autovibe autovibe 1d

# Temporary processing directories
D /tmp/autovibe-processing 0755 autovibe autovibe 4h  # Clean processing temps quickly
```

**Dynamic configuration management**:

```bash
# /usr/local/bin/autovibe-retention-manager
#!/bin/bash
# Reads YAML config and generates systemd-tmpfiles rules dynamically

CONFIG="/etc/autovibe/retention-config.yaml"
TMPFILES_DIR="/etc/tmpfiles.d"

# Parse YAML and generate tmpfiles.d entries
yq eval '.retention_policies.traffic_logs' "$CONFIG" | while IFS= read -r line; do
    # Generate appropriate systemd-tmpfiles rules based on config
    echo "Z /mnt/data/autovibe/traffic-logs/$uuid - autovibe autovibe $retention"
done > "$TMPFILES_DIR/autovibe-dynamic.conf"

# Reload systemd-tmpfiles configuration
systemctl restart systemd-tmpfiles-clean
```

**Enterprise automation integration**:

```yaml
# Ansible playbook example for retention management
- name: Configure AutoVibe retention policies
  template:
    src: retention-config.yaml.j2
    dest: /etc/autovibe/retention-config.yaml
  vars:
    # Environment-specific retention periods
    retention_multiplier: "{{ '2' if env == 'production' else '0.5' }}"
    security_retention: "30d"
    
- name: Generate systemd-tmpfiles configuration  
  command: /usr/local/bin/autovibe-retention-manager
  notify: restart tmpfiles
```

**Monitoring and alerting**:

```bash
# Disk usage monitoring with configurable thresholds
/usr/local/bin/autovibe-storage-monitor:
- Warns at 80% disk usage
- Triggers emergency cleanup at 90%  
- Reports retention policy violations
- Integrates with Prometheus/Grafana for dashboards
```

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