# autovibe REST API Design

## API Architecture

The Hub exposes a comprehensive REST API for managing Intelligent Machines, projects, checkpoints, and resources. All endpoints use UUID-based identification and support real-time resource tracking.

### Base URL
The API base URL is https://hub.autovibe.local/api/v1 for all endpoint requests.

### Authentication
All API requests require authentication using Bearer token authorization headers with the AUTOVIBE_API_KEY and Content-Type set to application/json when sending requests to any endpoint.

## Machine Management Endpoints

### Spawn Intelligent Machine
The POST /machines endpoint creates new Intelligent Machines with a request body containing machine type (claude_code), project UUID, checkpoint UUID, prompt text, file modifications as key-value pairs, and resource constraints including maximum money, time in minutes, and API calls. The response includes the assigned machine UUID, VM ID, spawning status, estimated resource usage breakdown, and spawning timestamp.

### Get Machine Status
The GET /machines/{machine_uuid} endpoint retrieves current machine status including UUID, type, status, project UUID, VM ID, and original prompt. The response provides detailed resource usage with estimated and actual values for money, time, API calls, and compute resources. Additional information includes current progress description, last checkpoint reference, total checkpoints created, spawning timestamp, and last update timestamp.

### List Machines
The GET /machines endpoint supports filtering by status and project_uuid with pagination via limit parameter. The response contains an array of machine objects, total count of all matching machines, a has_more boolean indicating additional results, and next_offset for pagination continuation.

### Terminate Machine
The POST /machines/{machine_uuid}/terminate endpoint accepts a request body with termination reason and optional checkpoint creation flag. The response confirms termination status, provides the final checkpoint UUID if created, includes complete resource usage summary, and records the termination timestamp.

### Request Manual Checkpoint
The POST /machines/{machine_uuid}/checkpoint endpoint creates manual checkpoints with a request body containing context identifier and description text. The response provides the new checkpoint UUID, creation timestamp, and checkpoint size in megabytes.

## Project Management Endpoints

### Create Project
The POST /projects endpoint creates new projects with a request body containing project name, greediness factor (0-1), description, and resource budgets specifying daily money, time hours, API calls, and storage limits. The response includes the assigned project UUID, confirmed name and greediness, initial checkpoint reference, and creation timestamp.

### Get Project Details
The GET /projects/{project_uuid} endpoint retrieves comprehensive project information including UUID, name, greediness factor, current checkpoint reference, and today's resource usage with allocated, used, and remaining amounts for money, time, and API calls. Additional details include active machine count, total checkpoint count, creation timestamp, and last update timestamp.

### Update Project
The PUT /projects/{project_uuid} endpoint updates existing projects with a request body containing modified fields such as greediness factor and description text.

### List Projects
The GET /projects endpoint supports filtering by active status and sorting by greediness in descending order. The response contains an array of project objects with basic information including UUID, name, greediness, status, and daily resource usage summary, plus total count of matching projects.

## Checkpoint Management Endpoints

### List Project Checkpoints
The GET /projects/{project_uuid}/checkpoints endpoint supports pagination via limit parameter and filtering by checkpoint type. The response contains an array of checkpoint objects with UUID, creation timestamp, creating machine reference, type, size in megabytes, description, and file count, plus total count and has_more indicator for pagination.

### Get Checkpoint Details
The GET /checkpoints/{checkpoint_uuid} endpoint returns detailed checkpoint information including UUID, project UUID, parent checkpoint reference, creating machine UUID, and Proxmox snapshot ID. The response includes a file manifest with size and SHA256 hash for each file, creation timestamp, total size in bytes, and metadata containing trigger reason and machine context.

### Spawn from Checkpoint
The POST /checkpoints/{checkpoint_uuid}/spawn endpoint creates new machines from existing checkpoints with a request body containing machine type, prompt text, and optional file modifications. The response provides the new machine UUID, spawning timestamp, and parent checkpoint reference.

### Compare Checkpoints
The GET /checkpoints/{checkpoint1_uuid}/diff/{checkpoint2_uuid} endpoint compares two checkpoints and returns arrays of added, modified, and deleted files. The response includes total change count and a human-readable summary of the differences between checkpoints.

## Resource Management Endpoints

### Get Resource Overview
The GET /resources/overview endpoint provides system-wide resource statistics including daily money budget and usage, active machine count, and total project count. Resource utilization percentages are provided for money, time, API calls, and compute resources. The response includes top resource consumers with project details and usage amounts.

### Get Project Resource Usage
The GET /resources/projects/{project_uuid} endpoint supports period filtering (last_7_days) and returns total usage for money, time hours, API calls, and compute hours. The response includes daily breakdown with date-specific usage and active machine counts, plus efficiency metrics including cost per checkpoint, API calls per operation, and average machine runtime.

### Update Project Resource Budget
The PUT /resources/projects/{project_uuid}/budget endpoint updates project resource limits with a request body containing daily money allocation, daily API calls limit, and maximum concurrent machine count.

## Hub Management Endpoints

### Get Hub Status
The GET /hub/status endpoint returns comprehensive Hub health information including Hub UUID, status, uptime hours, version, and last checkpoint reference. The response provides active machine and project counts, resource utilization percentages for CPU, memory, and disk, plus API statistics including requests per hour, average response time in milliseconds, and error rate percentage.

### Create Hub Checkpoint
The POST /hub/checkpoint endpoint creates Hub checkpoints with a request body containing checkpoint type, reason, and description. The response provides the new checkpoint UUID, creation timestamp, and estimated completion time.

### Spawn Experimental Hub
The POST /hub/spawn-experimental endpoint creates experimental Hub instances with a request body specifying base checkpoint, experiment name, configuration changes including algorithm and optimization settings, and experiment duration in days. The response provides experimental Hub UUID, experiment ID, and estimated spawn time.

## Monitoring and Analytics Endpoints

### System Health
The GET /health endpoint returns overall system health status and individual service status for hub, database, proxmox, and api components. The response includes active machine count, overall resource utilization percentage, and timestamp of the last checkpoint.

### Analytics and Metrics
The GET /analytics/efficiency endpoint supports period filtering (last_30_days) and returns overall efficiency metrics including successful operations, total operations, success rate, and average cost per success. Machine performance is broken down by type with success rates, average costs, and timing. Resource trends provide daily efficiency scores and cost per operation over time.

## Error Responses

### Standard Error Format
All API errors use a consistent JSON format with success boolean set to false, error object containing code, message, details with context-specific information, unique request ID for tracking, and timestamp of the error occurrence.

### HTTP Status Codes
- **200**: Success
- **201**: Created (machines, projects, checkpoints)
- **400**: Bad Request (invalid parameters)
- **401**: Unauthorized (invalid API key)
- **403**: Forbidden (insufficient permissions)
- **404**: Not Found (resource doesn't exist)
- **409**: Conflict (resource limits exceeded)
- **429**: Too Many Requests (rate limited)
- **500**: Internal Server Error (system failure)

This comprehensive API enables full management of the autovibe system while maintaining security, resource controls, and detailed monitoring capabilities.