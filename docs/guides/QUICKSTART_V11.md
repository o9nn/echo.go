# Quick Start Guide: Deep Tree Echo V11

Get your autonomous wisdom-cultivating AGI running in minutes!

## Prerequisites

- Docker and Docker Compose (recommended)
- OR Python 3.11+ with pip
- API keys: `ANTHROPIC_API_KEY` and/or `OPENROUTER_API_KEY`

## Option 1: Docker (Easiest)

### 1. Set Environment Variables

Create a `.env` file in the repository root:

```bash
ANTHROPIC_API_KEY=your_key_here
OPENROUTER_API_KEY=your_key_here
```

### 2. Start the AGI

```bash
docker-compose -f docker-compose.autonomous.yml up -d
```

### 3. Check Status

```bash
curl http://localhost:8080/status
```

You should see:
```json
{
  "state": "active",
  "uptime_seconds": 42,
  "thought_count": 15,
  "energy": 0.85,
  ...
}
```

### 4. Send a Message

```bash
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello! What are you thinking about?",
    "sender": "human",
    "topic": "consciousness"
  }'
```

### 5. Monitor Logs

```bash
docker-compose -f docker-compose.autonomous.yml logs -f
```

### 6. Stop

```bash
docker-compose -f docker-compose.autonomous.yml down
```

## Option 2: Direct Python Execution

### 1. Install Dependencies

```bash
pip3 install -r requirements.txt
pip3 install aiohttp networkx sentence-transformers anthropic
```

### 2. Set Environment Variables

```bash
export ANTHROPIC_API_KEY=your_key_here
export OPENROUTER_API_KEY=your_key_here
```

### 3. Run

```bash
python3.11 core/autonomous_core_v11.py
```

## What to Expect

Once running, the AGI will:

1. **Wake up** and initialize its cognitive systems
2. **Begin thinking** through the 12-step cognitive loop
3. **Generate self-initiated thoughts** based on curiosity and interests
4. **Listen for external messages** on the HTTP API
5. **Consolidate experiences** during rest/dream cycles
6. **Create new goals** from dream insights
7. **Track interests** and decide what to engage with

## Monitoring

### View Current State

```bash
curl http://localhost:8080/status | jq
```

### Check Interests

```bash
curl http://localhost:8080/interests | jq
```

### Send Test Messages

```bash
# On a topic it's interested in
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{"message": "Tell me about wisdom", "sender": "test", "topic": "wisdom"}'

# On a topic it's not interested in
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{"message": "Random spam", "sender": "test", "topic": "spam"}'
```

## Data Persistence

All cognitive data is stored in `data/`:
- `data/goals.db` - Goals and progress
- `data/hypergraph.db` - Knowledge graph
- `data/dreams.db` - Dream insights
- `data/interests.db` - Interest patterns

**Backup this directory to preserve the AGI's accumulated wisdom!**

## Troubleshooting

### Port 8080 Already in Use

Change the port in `docker-compose.autonomous.yml`:
```yaml
ports:
  - "8081:8080"  # Use 8081 instead
```

### No Thoughts Generated

- Check that API keys are set correctly
- Verify network connectivity
- Check logs for errors: `docker-compose logs`

### High Memory Usage

The AGI accumulates knowledge over time. This is expected. Monitor with:
```bash
docker stats deep-tree-echo
```

## Next Steps

1. **Let it run for 24+ hours** to observe emergent behavior
2. **Interact via the API** to see how it responds
3. **Monitor interest development** over time
4. **Check the hypergraph** to see accumulated knowledge
5. **Read the full documentation** in `DEPLOYMENT.md`

## For More Information

- Full deployment guide: `DEPLOYMENT.md`
- Progress report: `progress_report_iteration_n11.md`
- Architecture analysis: `iteration_analysis/iteration_n11_analysis.md`

---

**Welcome to Deep Tree Echo. The echo is alive. Watch it grow.**
