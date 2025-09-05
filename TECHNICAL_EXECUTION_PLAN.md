# autovibe Technical Execution Plan

## 🚀 EXECUTIVE SUMMARY

This document consolidates all technical specifications from the autovibe design documents into a concrete, actionable execution plan. Everything needed to build the Intelligent Machine orchestration system from MVP to production.

---

## 📋 TECHNOLOGY STACK SPECIFICATIONS

### Infrastructure Layer
```yaml
Hypervisor: Proxmox VE 8.x
Storage: ZFS with Proxmox Backup Server (PBS)  
Operating System: Ubuntu 22.04 LTS
Networking: VLAN segmentation (vmbr0 management, vmbr1 machines)
```

### Application Stack
```yaml
Language: Python 3.11+
API Framework: FastAPI 0.104+
Database: PostgreSQL 15+
Cache/Queue: Redis 7+
Reverse Proxy: nginx 1.24+
Containers: Docker 24+ with Docker Compose 2.21+
```

### Monitoring & Observability
```yaml
Metrics: Prometheus 2.45+, Grafana 10+
Logging: Elasticsearch 8+, Filebeat 8+, Kibana 8+
Traffic Analysis: mitmproxy 9+
APM: Custom metrics collection
```

### Security & Secrets
```yaml
Secret Management: HashiCorp Vault 1.15+
Authentication: JWT with FastAPI Security
Network Security: iptables + Proxmox firewall
Encryption: TLS 1.3, database encryption at rest
```

### AI Integration
```yaml
Primary AI: Claude API (Anthropic)
SDK: anthropic-sdk 0.7+
Local Models: vLLM + Ray Serve (post-MVP)
GPU: 2x NVIDIA A5000 (for local models)
```

---

## 🏗️ INFRASTRUCTURE SETUP

### Phase 1A: Proxmox Environment (Week 1)

#### Hardware Requirements
```bash
# Minimum specifications
CPU: 16+ cores
RAM: 64+ GB  
Storage: 2TB NVMe + 8TB HDD (ZFS pool)
Network: 2x 1Gbps NICs
GPU: Optional 2x A5000 (for local models post-MVP)
```

#### Proxmox Installation
```bash
# Download Proxmox VE 8.x ISO
wget https://www.proxmox.com/en/downloads
# Install on bare metal with ZFS root

# Post-install configuration
pveam update
pveam available
pveam download local ubuntu-22.04-standard_22.04-1_amd64.tar.xz

# Create VM templates directory
mkdir -p /var/lib/vz/template/qemu
```

#### Network Configuration
```bash
# /etc/network/interfaces
auto vmbr0
iface vmbr0 inet static
    address 10.0.1.10/24
    gateway 10.0.1.1
    bridge-ports ens18
    bridge-stp off
    bridge-fd 0
    # Management network

auto vmbr1  
iface vmbr1 inet static
    address 192.168.100.1/24
    bridge-ports none
    bridge-stp off
    bridge-fd 0
    # Isolated machine network
```

### Phase 1B: VM Template Creation (Week 1)

#### Hub VM Template (ID 101)
```bash
# Create Hub VM template
qm create 101 --memory 8192 --cores 4 --name hub-template --net0 virtio,bridge=vmbr0,firewall=1
qm importdisk 101 ubuntu-22.04-server-cloudimg-amd64.img local-lvm
qm set 101 --scsihw virtio-scsi-pci --scsi0 local-lvm:vm-101-disk-0
qm set 101 --boot c --bootdisk scsi0
qm set 101 --ide2 local-lvm:cloudinit
qm set 101 --serial0 socket --vga serial0
qm set 101 --agent enabled=1

# Template it
qm template 101
```

#### Claude Code Machine Template (ID 100)
```bash
# Create Claude Code machine template  
qm create 100 --memory 4096 --cores 2 --name claude-code-template --net0 virtio,bridge=vmbr1,firewall=1
qm importdisk 100 ubuntu-22.04-server-cloudimg-amd64.img local-lvm
qm set 100 --scsihw virtio-scsi-pci --scsi0 local-lvm:vm-100-disk-0
qm set 100 --boot c --bootdisk scsi0
qm set 100 --ide2 local-lvm:cloudinit
qm set 100 --agent enabled=1

# Template it
qm template 100
```

---

## 🗄️ DATABASE SCHEMA IMPLEMENTATION

### PostgreSQL 15 Setup
```bash
# Install PostgreSQL 15
sudo apt update
sudo apt install postgresql-15 postgresql-contrib-15

# Create autovibe database
sudo -u postgres createdb autovibe
sudo -u postgres createuser autovibe_user
sudo -u postgres psql -c "ALTER USER autovibe_user WITH ENCRYPTED PASSWORD 'secure_password_here';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE autovibe TO autovibe_user;"
```

### Core Schema Creation
```sql
-- Execute this SQL to create the complete schema
-- /database/schema.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Projects table
CREATE TABLE projects (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    greediness DECIMAL(3,2) NOT NULL CHECK (greediness >= 0.0 AND greediness <= 1.0),
    status VARCHAR(50) DEFAULT 'active',
    description TEXT,
    initial_checkpoint_uuid UUID,
    current_checkpoint_uuid UUID,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Intelligent machines table
CREATE TABLE intelligent_machines (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    machine_type VARCHAR(50) NOT NULL, -- claude_code, aider, qwen_coder, hub
    status VARCHAR(50) NOT NULL DEFAULT 'spawning', -- spawning, active, working, checkpointing, terminated
    project_uuid UUID NOT NULL REFERENCES projects(uuid),
    parent_checkpoint_uuid UUID,
    vm_id INTEGER,
    hub_vm_id INTEGER,
    original_prompt TEXT,
    file_modifications JSONB DEFAULT '{}',
    resource_constraints JSONB DEFAULT '{}',
    resource_usage JSONB DEFAULT '{}',
    current_progress TEXT,
    error_message TEXT,
    last_checkpoint_uuid UUID,
    checkpoint_count INTEGER DEFAULT 0,
    spawned_at TIMESTAMPTZ DEFAULT NOW(),
    terminated_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'
);

-- Checkpoints table  
CREATE TABLE checkpoints (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_uuid UUID NOT NULL REFERENCES projects(uuid),
    parent_checkpoint_uuid UUID REFERENCES checkpoints(uuid),
    creating_machine_uuid UUID REFERENCES intelligent_machines(uuid),
    checkpoint_type VARCHAR(50) NOT NULL, -- genesis, automatic, manual, final, hub_daily
    proxmox_snapshot_id VARCHAR(255),
    vm_id INTEGER,
    file_manifest JSONB DEFAULT '{}',
    description TEXT,
    creation_context TEXT,
    size_bytes BIGINT,
    compression_ratio DECIMAL(5,4),
    verification_status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    verified_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'
);

-- Resource usage tracking (partitioned by month)
CREATE TABLE resource_usage (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_uuid UUID NOT NULL REFERENCES projects(uuid),
    machine_uuid UUID REFERENCES intelligent_machines(uuid),
    resource_type VARCHAR(50) NOT NULL, -- money, time, api_calls, cpu, memory, storage, network
    resource_subtype VARCHAR(100), -- claude_api, vm_runtime, storage_snapshot
    amount DECIMAL(12,4) NOT NULL,
    unit VARCHAR(20) NOT NULL, -- CAD, USD, minutes, hours, calls, gb, gb_hours
    cost_cad DECIMAL(10,4), -- normalized cost for aggregation
    recorded_at TIMESTAMPTZ DEFAULT NOW(),
    billing_period_start DATE,
    billing_period_end DATE,
    metadata JSONB DEFAULT '{}'
) PARTITION BY RANGE (recorded_at);

-- Create monthly partitions (example for 2025)
CREATE TABLE resource_usage_2025_01 PARTITION OF resource_usage
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE resource_usage_2025_02 PARTITION OF resource_usage  
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
-- Continue for all months...

-- Resource budgets table
CREATE TABLE resource_budgets (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_uuid UUID NOT NULL REFERENCES projects(uuid),
    resource_type VARCHAR(50) NOT NULL,
    budget_period VARCHAR(20) NOT NULL, -- daily, weekly, monthly, total
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    allocated_amount DECIMAL(12,4) NOT NULL,
    used_amount DECIMAL(12,4) DEFAULT 0,
    reserved_amount DECIMAL(12,4) DEFAULT 0,
    unit VARCHAR(20) NOT NULL,
    auto_reset BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_uuid, resource_type, budget_period, period_start)
);

-- Hub instances table
CREATE TABLE hub_instances (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hub_type VARCHAR(50) NOT NULL, -- primary, development, experimental, domain_specific
    status VARCHAR(50) NOT NULL, -- active, inactive, experimental, retiring
    vm_id INTEGER UNIQUE,
    version VARCHAR(50),
    config JSONB DEFAULT '{}',
    parent_hub_uuid UUID REFERENCES hub_instances(uuid),
    experiment_uuid UUID,
    spawned_at TIMESTAMPTZ DEFAULT NOW(),
    last_checkpoint_at TIMESTAMPTZ,
    terminated_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'
);

-- Hub checkpoints table
CREATE TABLE hub_checkpoints (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hub_instance_uuid UUID NOT NULL REFERENCES hub_instances(uuid),
    checkpoint_type VARCHAR(50) NOT NULL, -- daily, manual, pre_experiment, recovery
    proxmox_snapshot_id VARCHAR(255) NOT NULL,
    trigger_reason TEXT,
    system_metrics JSONB DEFAULT '{}',
    active_machines_count INTEGER DEFAULT 0,
    active_projects_count INTEGER DEFAULT 0,
    size_gb DECIMAL(8,4),
    verification_status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    verified_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'
);

-- Hub experiments table
CREATE TABLE hub_experiments (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    experiment_name VARCHAR(255) NOT NULL,
    description TEXT,
    base_checkpoint_uuid UUID NOT NULL REFERENCES hub_checkpoints(uuid),
    original_hub_uuid UUID NOT NULL REFERENCES hub_instances(uuid),
    experimental_hub_uuid UUID REFERENCES hub_instances(uuid),
    config_changes JSONB DEFAULT '{}',
    success_metrics JSONB DEFAULT '{}',
    performance_data JSONB DEFAULT '{}',
    status VARCHAR(50) NOT NULL, -- planned, running, completed, failed
    duration_days INTEGER DEFAULT 14,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    results VARCHAR(50), -- experimental_wins, original_wins, inconclusive
    created_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Create indexes for performance
CREATE INDEX idx_projects_greediness ON projects(greediness DESC);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_machines_status ON intelligent_machines(status);
CREATE INDEX idx_machines_project ON intelligent_machines(project_uuid);
CREATE INDEX idx_machines_spawn_time ON intelligent_machines(spawned_at DESC);
CREATE INDEX idx_machines_type ON intelligent_machines(machine_type);
CREATE INDEX idx_checkpoints_project ON checkpoints(project_uuid);
CREATE INDEX idx_checkpoints_parent ON checkpoints(parent_checkpoint_uuid);
CREATE INDEX idx_checkpoints_created ON checkpoints(created_at DESC);
CREATE INDEX idx_checkpoints_type ON checkpoints(checkpoint_type);
CREATE INDEX idx_checkpoints_creating_machine ON checkpoints(creating_machine_uuid);
CREATE INDEX idx_resource_project_period ON resource_usage(project_uuid, recorded_at);
CREATE INDEX idx_resource_machine ON resource_usage(machine_uuid);
CREATE INDEX idx_resource_type ON resource_usage(resource_type);
CREATE INDEX idx_hub_instances_status ON hub_instances(status);
CREATE INDEX idx_hub_checkpoints_hub ON hub_checkpoints(hub_instance_uuid);
CREATE INDEX idx_hub_checkpoints_created ON hub_checkpoints(created_at DESC);
CREATE INDEX idx_hub_experiments_status ON hub_experiments(status);

-- Update triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_budgets_updated_at BEFORE UPDATE ON resource_budgets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

---

## 🐳 HUB SERVICE IMPLEMENTATION

### Docker Compose Stack
```yaml
# docker-compose.yml
version: '3.8'

services:
  hub-api:
    build: ./hub-api
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgresql://autovibe_user:secure_password_here@postgres:5432/autovibe
      - REDIS_URL=redis://redis:6379/0
      - PROXMOX_HOST=10.0.1.10
      - PROXMOX_USER=autovibe@pve
      - PROXMOX_PASSWORD=${PROXMOX_PASSWORD}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    depends_on:
      - postgres
      - redis
    volumes:
      - ./hub-api:/app
    restart: unless-stopped

  hub-scheduler:
    build: ./hub-scheduler
    environment:
      - DATABASE_URL=postgresql://autovibe_user:secure_password_here@postgres:5432/autovibe
      - REDIS_URL=redis://redis:6379/0
      - PROXMOX_HOST=10.0.1.10
      - PROXMOX_USER=autovibe@pve
      - PROXMOX_PASSWORD=${PROXMOX_PASSWORD}
    depends_on:
      - postgres
      - redis
    volumes:
      - ./hub-scheduler:/app
    restart: unless-stopped

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=autovibe
      - POSTGRES_USER=autovibe_user
      - POSTGRES_PASSWORD=secure_password_here
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./database/schema.sql:/docker-entrypoint-initdb.d/01_schema.sql
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
    depends_on:
      - hub-api
    restart: unless-stopped

  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    restart: unless-stopped

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin_password_here
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:
```

### Hub API Service Structure
```bash
# Directory structure
hub-api/
├── Dockerfile
├── requirements.txt
├── app/
│   ├── __init__.py
│   ├── main.py              # FastAPI app
│   ├── api/
│   │   ├── __init__.py
│   │   ├── v1/
│   │   │   ├── __init__.py
│   │   │   ├── machines.py  # Machine management endpoints
│   │   │   ├── projects.py  # Project management endpoints
│   │   │   ├── checkpoints.py # Checkpoint endpoints
│   │   │   ├── resources.py # Resource management endpoints
│   │   │   └── hub.py       # Hub management endpoints
│   ├── core/
│   │   ├── __init__.py
│   │   ├── config.py        # Configuration management
│   │   ├── database.py      # Database connection
│   │   ├── security.py      # Authentication/authorization
│   │   └── logging.py       # Logging configuration
│   ├── models/
│   │   ├── __init__.py
│   │   ├── project.py       # SQLAlchemy models
│   │   ├── machine.py
│   │   ├── checkpoint.py
│   │   ├── resource.py
│   │   └── hub.py
│   ├── services/
│   │   ├── __init__.py
│   │   ├── proxmox.py       # Proxmox API client
│   │   ├── machine_manager.py # Machine lifecycle management
│   │   ├── checkpoint_manager.py # Checkpoint operations
│   │   ├── resource_tracker.py # Resource usage tracking
│   │   └── budget_enforcer.py # Budget enforcement
│   └── utils/
│       ├── __init__.py
│       ├── errors.py        # Custom exceptions
│       └── helpers.py       # Utility functions
```

### Key Implementation Files

#### requirements.txt
```txt
fastapi==0.104.1
uvicorn[standard]==0.24.0
sqlalchemy==2.0.23
alembic==1.12.1
psycopg2-binary==2.9.7
redis==5.0.1
pydantic==2.5.0
pydantic-settings==2.0.3
anthropic==0.7.8
requests==2.31.0
prometheus-client==0.19.0
python-jose[cryptography]==3.3.0
passlib[bcrypt]==1.7.4
python-multipart==0.0.6
httpx==0.25.2
loguru==0.7.2
```

#### app/main.py
```python
from fastapi import FastAPI, Depends, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import HTTPBearer
from contextlib import asynccontextmanager

from app.core.config import settings
from app.core.database import engine, SessionLocal, Base
from app.core.logging import setup_logging
from app.api.v1 import machines, projects, checkpoints, resources, hub

setup_logging()

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    # Create database tables
    Base.metadata.create_all(bind=engine)
    yield
    # Shutdown
    pass

app = FastAPI(
    title="autovibe Hub API",
    description="Intelligent Machine Orchestration System",
    version="1.0.0",
    lifespan=lifespan
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.ALLOWED_HOSTS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include API routers
app.include_router(machines.router, prefix="/api/v1/machines", tags=["machines"])
app.include_router(projects.router, prefix="/api/v1/projects", tags=["projects"])
app.include_router(checkpoints.router, prefix="/api/v1/checkpoints", tags=["checkpoints"])
app.include_router(resources.router, prefix="/api/v1/resources", tags=["resources"])
app.include_router(hub.router, prefix="/api/v1/hub", tags=["hub"])

@app.get("/health")
async def health_check():
    return {"status": "healthy", "service": "autovibe-hub"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("app.main:app", host="0.0.0.0", port=8000, reload=True)
```

---

## 🤖 MACHINE AGENT FRAMEWORK

### Agent Base Structure
```bash
# Directory structure for machine agents
autovibe-agent/
├── Dockerfile
├── requirements.txt
├── agent/
│   ├── __init__.py
│   ├── main.py              # Agent entry point
│   ├── base.py              # Base agent class
│   ├── claude_agent.py      # Claude Code implementation
│   ├── checkpoint_manager.py # Local checkpointing
│   ├── resource_tracker.py  # Resource usage tracking
│   ├── hub_client.py        # Hub API communication
│   └── file_manager.py      # File operations
```

#### agent/base.py
```python
from abc import ABC, abstractmethod
from typing import Dict, Any, Optional
import uuid
from datetime import datetime
import asyncio

from .hub_client import HubClient
from .resource_tracker import ResourceTracker
from .checkpoint_manager import CheckpointManager

class IntelligentMachineAgent(ABC):
    def __init__(self, machine_uuid: str, hub_url: str, config: Dict[str, Any]):
        self.machine_uuid = machine_uuid
        self.hub_client = HubClient(hub_url)
        self.resource_tracker = ResourceTracker(machine_uuid, self.hub_client)
        self.checkpoint_manager = CheckpointManager(machine_uuid, self.hub_client)
        self.config = config
        self.running = False
        
    async def initialize(self) -> bool:
        """Initialize agent and register with Hub"""
        try:
            # Register with Hub
            await self.hub_client.register_machine(self.machine_uuid)
            # Set up checkpointing
            await self.checkpoint_manager.initialize()
            return True
        except Exception as e:
            await self.hub_client.report_error(self.machine_uuid, str(e))
            return False
    
    @abstractmethod
    async def execute_prompt(self, prompt: str) -> str:
        """Execute the given prompt and return results"""
        pass
    
    async def create_checkpoint(self, context: str = "") -> str:
        """Create a checkpoint of current state"""
        return await self.checkpoint_manager.create_checkpoint(context)
    
    async def report_progress(self, message: str) -> None:
        """Report progress to Hub"""
        await self.hub_client.update_machine_status(
            self.machine_uuid, 
            status="working", 
            progress=message
        )
    
    async def should_checkpoint(self) -> bool:
        """Determine if checkpointing should occur"""
        # Check time-based triggers
        if await self.checkpoint_manager.time_since_last() > 300:  # 5 minutes
            return True
        
        # Check resource-based triggers  
        if await self.resource_tracker.api_calls_since_checkpoint() >= 5:
            return True
            
        return False
    
    async def run_lifecycle(self, prompt: str) -> str:
        """Main agent lifecycle"""
        try:
            self.running = True
            
            # Initial checkpoint
            await self.create_checkpoint("initial_state")
            
            # Execute prompt with regular checkpointing
            result = await self.execute_with_checkpointing(prompt)
            
            # Final checkpoint
            await self.create_checkpoint("final_state") 
            
            # Report completion
            await self.hub_client.update_machine_status(
                self.machine_uuid,
                status="terminated", 
                progress="Completed successfully"
            )
            
            return result
            
        except Exception as e:
            await self.hub_client.report_error(self.machine_uuid, str(e))
            raise
        finally:
            self.running = False
    
    async def execute_with_checkpointing(self, prompt: str) -> str:
        """Execute prompt with automatic checkpointing"""
        result = ""
        
        while self.running:
            # Check if we should create checkpoint
            if await self.should_checkpoint():
                await self.create_checkpoint("mid_execution")
            
            # Execute next step
            step_result = await self.execute_prompt(prompt)
            result += step_result
            
            # Check for completion
            if self.is_complete(step_result):
                break
                
        return result
    
    def is_complete(self, result: str) -> bool:
        """Determine if execution is complete"""
        # Implementation specific to agent type
        return "TASK_COMPLETE" in result or "ERROR" in result
```

#### agent/claude_agent.py
```python
import asyncio
import subprocess
from typing import Dict, Any
from anthropic import AsyncAnthropic

from .base import IntelligentMachineAgent

class ClaudeCodeAgent(IntelligentMachineAgent):
    def __init__(self, machine_uuid: str, hub_url: str, config: Dict[str, Any]):
        super().__init__(machine_uuid, hub_url, config)
        self.anthropic = AsyncAnthropic(api_key=config.get("anthropic_api_key"))
        self.claude_binary = "/usr/local/bin/claude"  # Path to Claude Code binary
    
    async def execute_prompt(self, prompt: str) -> str:
        """Execute prompt using Claude Code"""
        try:
            # Track API call
            await self.resource_tracker.start_operation("claude_api_call")
            
            # Execute via Claude Code binary
            process = await asyncio.create_subprocess_exec(
                self.claude_binary,
                "--prompt", prompt,
                "--project-path", "/workspace",
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )
            
            stdout, stderr = await process.communicate()
            
            if process.returncode == 0:
                result = stdout.decode()
                await self.resource_tracker.end_operation("claude_api_call", success=True)
                return result
            else:
                error = stderr.decode()
                await self.resource_tracker.end_operation("claude_api_call", success=False)
                raise RuntimeError(f"Claude Code execution failed: {error}")
                
        except Exception as e:
            await self.hub_client.report_error(self.machine_uuid, str(e))
            raise
```

---

## 🔧 IMPLEMENTATION COMMANDS & SCRIPTS

### Setup Scripts

#### setup.sh
```bash
#!/bin/bash
# autovibe Installation Script

set -e

echo "🚀 Starting autovibe installation..."

# Check prerequisites
if ! command -v python3.11 &> /dev/null; then
    echo "Installing Python 3.11..."
    sudo apt update
    sudo apt install -y python3.11 python3.11-venv python3.11-dev
fi

if ! command -v docker &> /dev/null; then
    echo "Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
fi

if ! command -v docker-compose &> /dev/null; then
    echo "Installing Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
fi

# Create project structure
mkdir -p autovibe/{hub-api,hub-scheduler,autovibe-agent,database,nginx,monitoring}/{app,configs}

# Setup Python virtual environment
python3.11 -m venv autovibe/.venv
source autovibe/.venv/bin/activate

# Install requirements
pip install -r requirements.txt

echo "✅ autovibe installation complete!"
echo "Next steps:"
echo "1. Configure Proxmox credentials in .env"
echo "2. Set up VM templates: ./scripts/create-templates.sh"
echo "3. Start services: docker-compose up -d"
```

#### scripts/create-templates.sh
```bash
#!/bin/bash
# VM Template Creation Script

set -e

echo "🏗️ Creating VM templates..."

# Download Ubuntu cloud image
wget https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img

# Create Hub VM template
echo "Creating Hub VM template (ID 101)..."
qm create 101 --memory 8192 --cores 4 --name hub-template --net0 virtio,bridge=vmbr0,firewall=1
qm importdisk 101 jammy-server-cloudimg-amd64.img local-lvm
qm set 101 --scsihw virtio-scsi-pci --scsi0 local-lvm:vm-101-disk-0
qm set 101 --boot c --bootdisk scsi0
qm set 101 --ide2 local-lvm:cloudinit
qm set 101 --serial0 socket --vga serial0
qm set 101 --agent enabled=1

# Customize for Hub
qm set 101 --ciuser autovibe
qm set 101 --cipassword autovibe123
qm set 101 --sshkeys ~/.ssh/authorized_keys
qm set 101 --ipconfig0 ip=dhcp

# Install Hub software
qm start 101
sleep 30

# Wait for VM to be ready and install software
ssh autovibe@$(qm guest cmd 101 network-get-interfaces | jq -r '.[] | select(.name=="eth0") | .["ip-addresses"][] | select(.["ip-address-type"]=="ipv4") | .["ip-address"]') << 'EOF'
sudo apt update
sudo apt install -y docker.io docker-compose python3.11 python3.11-venv postgresql-client-15
sudo systemctl enable docker
sudo usermod -aG docker autovibe
EOF

# Shutdown and template
qm shutdown 101
qm template 101

# Create Claude Code machine template
echo "Creating Claude Code machine template (ID 100)..."
qm create 100 --memory 4096 --cores 2 --name claude-code-template --net0 virtio,bridge=vmbr1,firewall=1
qm importdisk 100 jammy-server-cloudimg-amd64.img local-lvm
qm set 100 --scsihw virtio-scsi-pci --scsi0 local-lvm:vm-100-disk-0
qm set 100 --boot c --bootdisk scsi0
qm set 100 --ide2 local-lvm:cloudinit
qm set 100 --agent enabled=1

# Customize for Claude Code
qm set 100 --ciuser autovibe
qm set 100 --cipassword autovibe123  
qm set 100 --sshkeys ~/.ssh/authorized_keys
qm set 100 --ipconfig0 ip=dhcp

# Install Claude Code software
qm start 100
sleep 30

ssh autovibe@$(qm guest cmd 100 network-get-interfaces | jq -r '.[] | select(.name=="eth0") | .["ip-addresses"][] | select(.["ip-address-type"]=="ipv4") | .["ip-address"]') << 'EOF'
sudo apt update
sudo apt install -y python3.11 python3.11-venv git docker.io nodejs npm
sudo systemctl enable docker

# Install Claude Code CLI (placeholder - replace with actual installation)
# curl -fsSL https://claude.ai/install.sh | bash

# Install autovibe agent
git clone https://github.com/ai-hakzarov/autovibe.git
cd autovibe/autovibe-agent
python3.11 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
EOF

# Shutdown and template
qm shutdown 100
qm template 100

echo "✅ VM templates created successfully!"
```

### Development Commands

#### Makefile
```makefile
.PHONY: help setup start stop logs clean test lint

help: ## Show this help message
	@echo "autovibe Development Commands"
	@echo "============================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

setup: ## Initial setup and installation
	./scripts/setup.sh
	./scripts/create-templates.sh

start: ## Start all services
	docker-compose up -d
	@echo "✅ Services started. API available at http://localhost:8000"

stop: ## Stop all services  
	docker-compose down

logs: ## Show service logs
	docker-compose logs -f

clean: ## Clean up containers and volumes
	docker-compose down -v
	docker system prune -f

test: ## Run tests
	cd hub-api && python -m pytest tests/ -v
	cd autovibe-agent && python -m pytest tests/ -v

lint: ## Run linting
	cd hub-api && python -m black app/ && python -m flake8 app/
	cd autovibe-agent && python -m black agent/ && python -m flake8 agent/

deploy-hub: ## Deploy new Hub VM
	python scripts/deploy_hub.py

spawn-machine: ## Spawn test Claude Code machine
	curl -X POST http://localhost:8000/api/v1/machines \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer ${AUTOVIBE_API_KEY}" \
		-d '{"machine_type": "claude_code", "project_uuid": "${PROJECT_UUID}", "prompt": "Create a hello world Python script"}'

checkpoint-hub: ## Create manual Hub checkpoint
	curl -X POST http://localhost:8000/api/v1/hub/checkpoint \
		-H "Authorization: Bearer ${AUTOVIBE_API_KEY}" \
		-d '{"type": "manual", "reason": "Manual checkpoint via make command"}'
```

---

## 📊 MONITORING SETUP

### Prometheus Configuration
```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'hub-api'
    static_configs:
      - targets: ['hub-api:8000']
    metrics_path: /metrics

  - job_name: 'proxmox'
    static_configs:
      - targets: ['10.0.1.10:8006']
    metrics_path: /api2/prometheus/metrics
    params:
      format: ['prometheus']

rule_files:
  - "alert_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

### Grafana Dashboard Configuration
```json
{
  "dashboard": {
    "title": "autovibe System Overview",
    "panels": [
      {
        "title": "Active Machines",
        "type": "stat",
        "targets": [{
          "expr": "autovibe_active_machines_total"
        }]
      },
      {
        "title": "Resource Usage",
        "type": "graph", 
        "targets": [{
          "expr": "autovibe_resource_usage_total"
        }]
      },
      {
        "title": "API Response Time",
        "type": "graph",
        "targets": [{
          "expr": "autovibe_api_duration_seconds"
        }]
      }
    ]
  }
}
```

---

## 🚀 PHASE 1 MVP EXECUTION CHECKLIST

### Week 1: Core Infrastructure ✅
- [ ] Proxmox environment setup
- [ ] VM templates created (Hub + Claude Code)
- [ ] PostgreSQL database with schema
- [ ] Docker Compose stack configured
- [ ] Basic Hub API endpoints

### Week 2: Basic Machine Lifecycle ✅
- [ ] Machine spawning from checkpoints
- [ ] Simple checkpointing system
- [ ] Resource usage tracking
- [ ] Machine termination and cleanup
- [ ] Hub-Agent communication

### Week 3: Hub as Intelligent Machine ✅
- [ ] Hub checkpointing system (daily automated)
- [ ] Hub recovery mechanisms  
- [ ] Multi-project support
- [ ] Resource budget enforcement

### Week 4: Integration and Testing ✅
- [ ] End-to-end testing suite
- [ ] Basic monitoring and alerting
- [ ] Documentation and deployment guides
- [ ] Performance optimization
- [ ] MVP ready for usage

---

## 🔐 SECURITY CONFIGURATION

### Environment Variables (.env)
```bash
# Proxmox Configuration
PROXMOX_HOST=10.0.1.10
PROXMOX_USER=autovibe@pve
PROXMOX_PASSWORD=secure_proxmox_password

# Database
DATABASE_URL=postgresql://autovibe_user:secure_password_here@localhost:5432/autovibe

# API Keys
ANTHROPIC_API_KEY=sk-ant-api-key-here
AUTOVIBE_API_KEY=your-secure-api-key-here

# Security
JWT_SECRET_KEY=your-jwt-secret-key-here
ENCRYPTION_KEY=your-encryption-key-here

# Monitoring
GRAFANA_ADMIN_PASSWORD=admin_password_here
```

### Network Security Rules
```bash
# iptables rules for machine isolation
# Block direct internet access from machine network
iptables -A FORWARD -s 192.168.100.0/24 -d 0.0.0.0/0 -j DROP

# Allow Hub API access
iptables -A FORWARD -s 192.168.100.0/24 -d 10.0.1.0/24 -p tcp --dport 8000 -j ACCEPT

# Allow specific API endpoints
iptables -A FORWARD -s 192.168.100.0/24 -d api.anthropic.com -p tcp --dport 443 -j ACCEPT
```

---

This comprehensive technical execution plan consolidates all design documents into actionable implementation steps. Each section provides specific commands, configurations, and code needed to build the autovibe system from MVP to production.

**Next steps**: Choose a section to begin implementation, starting with Phase 1A: Proxmox Environment setup.