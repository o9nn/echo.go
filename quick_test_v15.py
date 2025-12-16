#!/usr/bin/env python3
import asyncio
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent))

from core.autonomous_core_v15 import DeepTreeEchoV15

async def main():
    echo = DeepTreeEchoV15('data/quick_test.json')
    print(f'✓ V15 initialized')
    print(f'✓ Energy: {echo.energy_state.energy}')
    print(f'✓ Echobeats step: {echo.echobeats.state.current_step}')
    print(f'✓ Interests: {len(echo.interest_patterns.interests)}')
    print('✓ All basic attributes working')

if __name__ == "__main__":
    asyncio.run(main())
