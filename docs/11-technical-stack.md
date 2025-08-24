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
```yaml
models:
  qwen-32b-coder:
    model_path: "/mnt/data/autocode/model-weights/Qwen2.5-Coder-32B-Instruct-AWQ"
    tensor_parallel_size: 2
    gpu_memory_utilization: 0.85
    
  deepseek-r1-32b:
    model_path: "/mnt/data/autocode/model-weights/DeepSeek-R1-32B"
    tensor_parallel_size: 2
    gpu_memory_utilization: 0.85

  # Additional models with single GPU allocation
  qwen-14b-coder:
    tensor_parallel_size: 1
    
  # Smaller models can share GPU memory
```

### Ray Serve Deployment
```python
# Model load balancing across 2xA5000
@serve.deployment(num_replicas=2, ray_actor_options={"num_gpus": 1})
class ModelService:
    def __init__(self, model_name: str):
        self.model = vLLM(model=model_name)
    
    async def generate(self, prompt: str, max_tokens: int = 2048):
        return await self.model.generate(prompt, max_tokens=max_tokens)
```

## Orchestration Layer

### Nomad Configuration
```hcl
# nomad/worker-agent.nomad
job "autocode-worker" {
  datacenters = ["dc1"]
  type = "service"
  
  group "agent" {
    count = 4  # One per VM
    
    task "coding-agent" {
      driver = "exec"
      
      config {
        command = "/opt/autocode/worker"
        args = ["--agent-type", "${NOMAD_META_AGENT_TYPE}"]
      }
      
      env {
        DAILY_BUDGET = "${NOMAD_META_DAILY_BUDGET}"
        PROJECT_ID = "${NOMAD_META_PROJECT_ID}"
        ORCHESTRATOR_URL = "https://orchestrator.local:8443"
      }
      
      resources {
        cpu = 4000  # 4 cores per agent
        memory = 2048  # 2GB RAM per agent
      }
    }
  }
}
```

### SystemD Service Management
```ini
# /etc/systemd/system/autocode-worker.service
[Unit]
Description=Autocode Worker Agent
After=network.target
Wants=network.target

[Service]
Type=simple
User=autocode
WorkingDirectory=/opt/autocode
ExecStart=/opt/autocode/worker --config /etc/autocode/worker.conf
Restart=always
RestartSec=10
Environment=ORCHESTRATOR_URL=https://orchestrator.local:8443

[Install]
WantedBy=multi-user.target
```

## Network Architecture

### Traffic Routing
```
Worker VM → iptables REDIRECT → Orchestrator mitmproxy → Internet
           ↓
    Elasticsearch Logging
```

### mitmproxy Configuration
```python
# mitmproxy addon for logging
from mitmproxy import http
import json
import elasticsearch

class AutocodeLogger:
    def __init__(self):
        self.es = elasticsearch.Elasticsearch(['localhost:9200'])
    
    def request(self, flow: http.HTTPFlow):
        log_entry = {
            "timestamp": flow.request.timestamp_start,
            "worker_id": self.get_worker_from_ip(flow.client_conn.ip),
            "method": flow.request.method,
            "url": flow.request.pretty_url,
            "headers": dict(flow.request.headers),
            "size": len(flow.request.content)
        }
        self.es.index(index="autocode-traffic", body=log_entry)
```

### Certificate Management
```bash
# Automatic certificate generation for HTTPS interception
mitmproxy \
  --mode transparent \
  --showhost \
  --set confdir=/mnt/data/autocode/certs \
  --set certificate_authority_cert=/mnt/data/autocode/certs/mitmproxy-ca.pem \
  --script /opt/autocode/traffic-logger.py
```

## Secret Management

### Vault Configuration
```hcl
# vault/config.hcl
storage "file" {
  path = "/mnt/data/autocode/vault-data"
}

listener "tcp" {
  address = "0.0.0.0:8200"
  tls_cert_file = "/etc/vault/tls/vault.crt"
  tls_key_file = "/etc/vault/tls/vault.key"
}

api_addr = "https://orchestrator.local:8200"
cluster_addr = "https://orchestrator.local:8201"
ui = true
```

### Secret Access Pattern
```python
# Agent secret retrieval
import hvac

class SecretManager:
    def __init__(self):
        self.client = hvac.Client(url='https://orchestrator.local:8200')
        self.client.token = os.environ['VAULT_TOKEN']
    
    def get_payment_card(self):
        response = self.client.secrets.kv.v2.read_secret_version(
            path='payments/card_primary'
        )
        return response['data']['data']
    
    def get_crypto_wallet(self, currency='TON'):
        response = self.client.secrets.kv.v2.read_secret_version(
            path=f'crypto/{currency.lower()}_wallet'
        )
        return response['data']['data']
```

## Logging & Monitoring

### Elasticsearch Configuration
```yaml
# elasticsearch/elasticsearch.yml
cluster.name: autocode-cluster
node.name: autocode-node-1
path.data: /mnt/data/autocode/elasticsearch
path.logs: /mnt/data/autocode/logs/elasticsearch

network.host: 0.0.0.0
http.port: 9200

xpack.security.enabled: false  # For MVP
xpack.monitoring.collection.enabled: true
```

### Log Retention Policies
```python
# Loguru configuration with retention
from loguru import logger
import sys

logger.configure(
    handlers=[
        {
            "sink": sys.stdout,
            "format": "{time:YYYY-MM-DD HH:mm:ss} | {level} | {message}",
            "level": "INFO"
        },
        {
            "sink": "/mnt/data/autocode/logs/worker-{time:YYYY-MM-DD}.log",
            "format": "{time:YYYY-MM-DD HH:mm:ss} | {level} | {extra} | {message}",
            "rotation": "1 day",
            "retention": "30 days",
            "compression": "gz",
            "serialize": True  # JSON output
        }
    ]
)
```

### Filebeat Configuration
```yaml
# filebeat/filebeat.yml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /mnt/data/autocode/logs/*.log
  fields:
    service: autocode-worker
    environment: production

output.elasticsearch:
  hosts: ["localhost:9200"]
  index: "autocode-logs-%{+yyyy.MM.dd}"

processors:
- add_host_metadata:
    when.not.contains.tags: forwarded
```

## Agent Architecture

### Swappable Agent Interface
```python
from abc import ABC, abstractmethod
from typing import Dict, Any, List

class CodingAgent(ABC):
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.budget = config.get('daily_budget', 0)
        self.project_id = config.get('project_id')
    
    @abstractmethod
    async def execute_task(self, task: str) -> Dict[str, Any]:
        pass
    
    @abstractmethod
    async def get_spending_report(self) -> Dict[str, Any]:
        pass
    
    @abstractmethod
    async def check_budget(self) -> bool:
        pass

class ClaudeCodeAgent(CodingAgent):
    def __init__(self, config: Dict[str, Any]):
        super().__init__(config)
        self.claude_binary = "/usr/local/bin/claude"
    
    async def execute_task(self, task: str) -> Dict[str, Any]:
        # Implementation for Claude Code integration
        pass

class AiderAgent(CodingAgent):
    def __init__(self, config: Dict[str, Any]):
        super().__init__(config)
        self.aider_binary = "/usr/local/bin/aider"
    
    async def execute_task(self, task: str) -> Dict[str, Any]:
        # Implementation for Aider integration
        pass
```

### Agent Factory
```python
class AgentFactory:
    AGENTS = {
        'claude-code': ClaudeCodeAgent,
        'aider': AiderAgent,
        'qwen-coder': QwenCoderAgent,
        'cursor-cli': CursorCLIAgent
    }
    
    @classmethod
    def create_agent(cls, agent_type: str, config: Dict[str, Any]) -> CodingAgent:
        if agent_type not in cls.AGENTS:
            raise ValueError(f"Unknown agent type: {agent_type}")
        
        return cls.AGENTS[agent_type](config)
```

## Data Storage

### Directory Structure
```
/mnt/data/autocode/
├── elasticsearch/          # Elasticsearch data
│   ├── nodes/
│   └── indices/
├── vault-data/            # Vault backend storage
├── model-weights/         # ML model files
│   ├── qwen-2.5-coder-32b/
│   ├── deepseek-r1-32b/
│   └── ...
├── logs/                  # Application logs
│   ├── worker-logs/
│   ├── orchestrator/
│   └── traffic/
├── financial/             # Financial data
│   ├── transactions/
│   ├── budgets/
│   └── reports/
├── worker-artifacts/      # Agent outputs
│   ├── projects/
│   ├── generated-code/
│   └── results/
└── configs/              # Configuration files
    ├── nomad/
    ├── vault/
    └── agents/
```

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