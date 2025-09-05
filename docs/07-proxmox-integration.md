# Proxmox Integration Architecture

## Overview

Proxmox serves as the foundational infrastructure layer for autovibe, providing VM orchestration, snapshot management, and resource control. The Hub communicates with Proxmox APIs to manage the entire Intelligent Machine lifecycle.

## Proxmox API Integration

### Authentication and Permissions
The ProxmoxClient class handles authentication and API communication with Proxmox servers. It initializes with host credentials, authenticates by sending username and password to the Proxmox API ticket endpoint, extracts the API token from successful responses, and establishes an authenticated session. Authentication failures raise ProxmoxAuthenticationError exceptions.

### Required Proxmox Permissions
Hub VM user needs specific permissions:
- **VM.Allocate**: Create new VMs
- **VM.Config**: Modify VM configuration
- **VM.Console**: Access VM console (for debugging)
- **VM.PowerMgmt**: Start, stop, restart VMs
- **VM.Snapshot**: Create, delete, restore snapshots
- **Datastore.AllocateSpace**: Allocate storage for VMs and snapshots
- **Datastore.Audit**: Read storage usage information

## VM Lifecycle Management

### VM Creation from Templates
VM creation from templates involves a two-step process: First, clone a VM from the specified template using Proxmox API, generating a new VM ID and specifying target node, storage location, and full clone parameters. Second, configure the new VM's resources including CPU cores, memory allocation, network interface with virtio bridge and firewall, and storage disk configuration. The function returns the newly created VM ID.

### VM Template Management
VM template configurations define resource specifications and software packages for different machine types:

**Claude Code Template (ID 100)**: 2 CPU cores, 4GB memory, 20GB disk with API access network configuration. Includes Ubuntu 22.04 base, Docker, Python 3.11, Node.js, Claude API client, Git, and autovibe-agent.

**Hub Template (ID 101)**: 4 CPU cores, 8GB memory, 50GB disk with management access network configuration. Includes Ubuntu 22.04 server, Docker with Docker Compose, PostgreSQL client, Nginx, Proxmox VE client, and autovibe-hub software.

## Snapshot and Checkpoint Integration

### Snapshot Creation
VM snapshot creation initiates a Proxmox API call with snapshot parameters including name, description, and VM state inclusion (with RAM for full checkpoints). The system monitors snapshot creation progress using the returned task ID, waits for completion, verifies the snapshot exists in the VM's snapshot list, and returns snapshot metadata including ID, VM ID, creation timestamp, and size. Failed snapshot creation raises a SnapshotCreationError.

### Snapshot Restoration
VM snapshot restoration operates in two modes: If a new VM ID is provided, the system creates a new VM by cloning from the snapshot with parameters including target node and full clone settings (branching). If no new VM ID is provided, the system restores the existing VM to the snapshot state using Proxmox rollback API, monitors task completion, and returns the appropriate VM ID.

### Snapshot Management
Snapshot cleanup applies retention policies by iterating through all autovibe VMs and their snapshots. The retention policy evaluation considers snapshot age and naming conventions: recent snapshots are kept based on daily retention days, weekly snapshots (ending with '-weekly') are kept for the weekly retention period, monthly snapshots (ending with '-monthly') are kept for the monthly retention period, and manually marked important snapshots are always preserved. Snapshots meeting deletion criteria are removed and logged.

## Resource Monitoring and Control

### VM Resource Monitoring
VM resource monitoring retrieves real-time usage statistics via Proxmox API including CPU usage percentage, memory consumption, disk I/O statistics, network traffic, uptime, and VM status. Resource limit monitoring compares current usage against configured limits (default 90% for CPU and memory, 24 hours for maximum runtime) and returns a list of violations including CPU exceeded, memory exceeded, or maximum runtime exceeded.

### Resource Quota Enforcement
Resource quota enforcement operates at the project level by monitoring cumulative usage across all project VMs. The system calculates total resource consumption including CPU hours, memory gigabyte-hours, storage usage, and network traffic by aggregating data from all VMs associated with a project. Resource calculations include uptime adjustments for CPU and memory usage. The enforcement function compares actual usage against project quotas and triggers appropriate responses based on violation severity. For violations under 10%, the system logs warnings. For 10-50% violations, it throttles new VM creation and alerts project administrators. For severe violations over 50%, it performs emergency shutdown of non-critical VMs and notifies system administrators.

## Network Security and Isolation

### Network Configuration
Network security uses profile-based configuration for different machine types. The API access profile provides isolated access through a dedicated bridge with firewall enabled, allowing connections only to specific API endpoints (Anthropic, OpenAI, GitHub) and the Hub API while blocking all other destinations. The management access profile uses the main management bridge with firewall protection, permitting access to Proxmox management, database connections, API access for the Hub, and internal network ranges. VM network configuration applies the appropriate profile settings including bridge assignment, firewall activation, and security rule implementation. The system configures the VM's network interface with virtio driver and applies firewall rules based on the selected profile's allowed and blocked destinations.

This Proxmox integration provides the foundation for secure, scalable VM orchestration with comprehensive snapshot management and resource control.