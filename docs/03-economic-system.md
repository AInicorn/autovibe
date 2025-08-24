# AutoVibe Economic System

## Budget-First Architecture

Every operation in AutoVibe is governed by economic principles that prevent overspending and optimize for cost-effectiveness.

## Budget Allocation Framework

### Total System Money Budget
- **Bi-Weekly Limit**: $42 CAD default (adjustable)
- **Daily Distribution**: Automatic daily allocation based on project priorities
- **Emergency Reserve**: 10% held back for critical operations

### Project Greediness System
Projects receive budget allocation based on their "greediness" factor (0.0 to 1.0):

Example configuration:
- **project-0-autovibe**: greediness 0.42 (highest priority - core system)
- **project-2-crypto-profit**: greediness 0.1 (medium priority)
- **project-4-cryptopulse**: greediness 0.0 (free tier only)

### Daily Budget Calculation
Daily budget for each project is calculated as:
`daily_budget = (total_biweekly_budget / 14) * project_greediness`

## Cost Control Mechanisms

### Machine Spawn Controls
- **Pre-spawn Resource Check**: Verify sufficient resources before creating Intelligent Machine
- **Estimated Resource Validation**: Compare estimated vs available resources
- **Queue System**: Queue operations when resources insufficient

### Real-time Monitoring  
- **Resource Tracking**: Monitor all resource types (API calls, compute, time, money) in real-time
- **VM Resource Metrics**: Track compute, storage, network usage
- **Comprehensive Usage**: Track API calls as a resource alongside monetary costs

### Automatic Safeguards
- **Resource Limits**: When any resource budget is exhausted, machines cannot spawn new operations
- **Daily Reset**: Budget limits refresh at midnight UTC
- **Zero-balance Handling**: System gracefully handles zero-balance without transaction fees

## Resource Types

### Primary Resources
- **Money**: CAD, USD, EUR (tracked in base currency)
- **Time**: Machine runtime minutes/hours
- **API Calls**: Claude, OpenAI, other service calls
- **Compute**: CPU hours, memory GB-hours
- **Storage**: Disk GB usage
- **Network**: Data transfer GB

### Resource Budgets
Each project has allocations for multiple resource types:
- Monetary budget (daily/weekly/monthly)
- API call limits (per service)
- Compute hour limits
- Storage quotas
- Time limits for operations

## Budget Tracking

### Real-time Usage
- Track resource consumption as it happens
- Update budget allocations immediately
- Alert on threshold crossing
- Prevent overspending before it occurs

### Historical Analysis
- Resource usage patterns over time
- Cost per operation metrics
- Efficiency trends
- Budget utilization rates

## Post-MVP Features

See POST_MVP.md for advanced economic features including:
- Multi-currency support with crypto payments
- Budget forecasting and predictive analytics
- Dynamic budget reallocation based on ROI
- Cost center accounting
- Payment method cascades
- Advanced optimization algorithms

This economic system ensures AutoVibe remains cost-effective while maximizing project evolution capabilities.