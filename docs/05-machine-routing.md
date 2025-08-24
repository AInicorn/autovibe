# AutoVibe Intelligent Machine Routing Design [DRAFT]

> **⚠️ DRAFT DOCUMENT**: This design may need redesign. Currently simplified to MVP-only requirements.

## MVP Machine Routing (Simplified)

### Single Machine Type
For MVP, routing is extremely simple:
- **Only Claude Code**: All operations use Claude Code machines
- **No complexity analysis**: Every prompt gets the same treatment
- **Simple resource check**: Basic budget verification before spawning

### MVP Routing Process
1. **Budget Check**: Verify sufficient funds exist for operation
2. **VM Availability**: Check if worker VM slots are available  
3. **Spawn Decision**: If both pass, spawn Claude Code machine
4. **Queue Operation**: If resources unavailable, queue for later

### Resource Monitoring
- **Budget Tracking**: Monitor spending per operation
- **Usage Logging**: Log all API calls and costs
- **Simple Reporting**: Basic metrics on operations and costs

This simplified approach allows for rapid MVP development while establishing the foundation for more sophisticated routing in future iterations.