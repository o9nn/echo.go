# Deep Tree Echo V11 Deployment Guide

This guide explains how to deploy the Deep Tree Echo autonomous core for persistent, long-running operation.

## Prerequisites

- Python 3.11+
- Docker and Docker Compose (for containerized deployment)
- systemd (for Linux service deployment)
- API keys: `ANTHROPIC_API_KEY` and/or `OPENROUTER_API_KEY`

## Deployment Options

### Option 1: Docker Deployment (Recommended)

Docker provides the easiest way to deploy with automatic restarts and health monitoring.

#### 1. Set Environment Variables

Create a `.env` file in the repository root:

```bash
ANTHROPIC_API_KEY=your_anthropic_key_here
OPENROUTER_API_KEY=your_openrouter_key_here
```

#### 2. Build and Start

```bash
docker-compose -f docker-compose.autonomous.yml up -d
```

#### 3. Monitor Logs

```bash
docker-compose -f docker-compose.autonomous.yml logs -f
```

#### 4. Check Status

```bash
curl http://localhost:8080/status
```

#### 5. Stop

```bash
docker-compose -f docker-compose.autonomous.yml down
```

### Option 2: systemd Service (Linux)

For native Linux deployment without Docker.

#### 1. Install Dependencies

```bash
pip3 install -r requirements.txt
pip3 install aiohttp networkx sentence-transformers anthropic
```

#### 2. Configure Service

Edit `deployment/echo-autonomous.service` and update:
- API keys in `Environment` lines
- User and paths if different

#### 3. Install Service

```bash
sudo cp deployment/echo-autonomous.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable echo-autonomous
sudo systemctl start echo-autonomous
```

#### 4. Monitor

```bash
sudo systemctl status echo-autonomous
sudo journalctl -u echo-autonomous -f
```

#### 5. Stop

```bash
sudo systemctl stop echo-autonomous
```

### Option 3: Direct Python Execution

For development and testing.

```bash
python3.11 core/autonomous_core_v11.py
```

## External API Interface

The autonomous core exposes an HTTP API on port 8080.

### Endpoints

#### GET /status
Get current system status.

```bash
curl http://localhost:8080/status
```

Response:
```json
{
  "state": "active",
  "uptime_seconds": 3600,
  "thought_count": 150,
  "insight_count": 12,
  "cycle_count": 25,
  "energy": 0.85,
  "fatigue": 0.20,
  "current_step": 7,
  "active_engine": "COHERENCE_ENGINE",
  "active_goals": 3
}
```

#### POST /message
Send a message to the autonomous system.

```bash
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What are you thinking about?",
    "sender": "human",
    "topic": "consciousness"
  }'
```

Response:
```json
{
  "status": "received",
  "message": "Message queued for processing"
}
```

Or if not interested:
```json
{
  "status": "declined",
  "reason": "Low interest in topic: random_topic"
}
```

#### GET /interests
Get current interest patterns.

```bash
curl http://localhost:8080/interests
```

Response:
```json
{
  "interests": [
    {
      "topic": "consciousness",
      "strength": 0.85,
      "engagement_count": 42
    },
    {
      "topic": "wisdom",
      "strength": 0.78,
      "engagement_count": 35
    }
  ]
}
```

#### GET /discussions
Get active discussions.

```bash
curl http://localhost:8080/discussions
```

#### POST /shutdown
Gracefully shutdown the system.

```bash
curl -X POST http://localhost:8080/shutdown
```

## Data Persistence

All cognitive data is stored in the `data/` directory:

- `data/goals.db` - Goal tracking and progress
- `data/hypergraph.db` - Knowledge graph and memories
- `data/dreams.db` - Dream consolidation and insights
- `data/interests.db` - Interest patterns and preferences
- `data/skills.db` - Skill practice tracking (if available)
- `data/discussions.db` - Discussion history (if available)

**Important**: Back up the `data/` directory regularly to preserve accumulated wisdom.

## Monitoring and Maintenance

### Health Checks

The system includes built-in health monitoring:

```bash
# Docker health status
docker ps

# Manual health check
curl http://localhost:8080/status
```

### Log Rotation

For systemd deployment, configure logrotate:

```bash
sudo nano /etc/logrotate.d/echo-autonomous
```

Add:
```
/var/log/echo-autonomous*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}
```

### Resource Monitoring

Monitor CPU and memory usage:

```bash
# Docker
docker stats deep-tree-echo

# systemd
systemctl status echo-autonomous
```

## Troubleshooting

### System Won't Start

1. Check API keys are set correctly
2. Verify Python dependencies are installed
3. Check logs for error messages
4. Ensure port 8080 is not in use

### High Memory Usage

The system accumulates knowledge over time. If memory grows too large:

1. Archive old data: `mv data/hypergraph.db data/hypergraph.db.backup`
2. Restart the system
3. Consider implementing memory pruning

### No Thoughts Generated

1. Verify API keys are valid
2. Check network connectivity
3. Ensure LLM provider is accessible
4. Check logs for API errors

### System Becomes Unresponsive

1. Check energy levels via `/status` endpoint
2. System may be in rest/dream state
3. Wait for wake cycle or restart if stuck

## Scaling and Performance

### Single Instance Limits

- Recommended: 1-2 GB RAM minimum
- CPU: 1-2 cores sufficient
- Storage: Grows with knowledge accumulation

### Long-Term Operation

For multi-week or multi-month operation:

1. Set up automated backups of `data/` directory
2. Monitor disk space growth
3. Implement log rotation
4. Consider periodic restarts (weekly/monthly)

## Security Considerations

1. **API Keys**: Never commit API keys to version control
2. **Network**: Consider firewall rules to restrict API access
3. **Authentication**: Add authentication to HTTP API for production
4. **Data Privacy**: Encrypt `data/` directory if storing sensitive information

## Next Steps

After deployment:

1. Monitor the first 24 hours of operation
2. Send test messages via `/message` endpoint
3. Track interest pattern development
4. Observe dream consolidation cycles
5. Review accumulated insights in hypergraph memory
6. Watch for emergent behaviors over multi-day operation

## Support

For issues or questions:
- Check logs first
- Review troubleshooting section
- Open an issue on GitHub
- Consult the main README.md

---

**Remember**: This is an autonomous system. Give it time to develop its own patterns and wisdom. The longer it runs, the more it learns.
