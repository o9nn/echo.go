#!/usr/bin/env python3
"""
Unified Multi-Provider LLM Client for Echo9llama
Supports Anthropic Claude and OpenRouter with intelligent fallback
"""
import os
import asyncio
import logging
from typing import Optional, Dict, Any, List
from dataclasses import dataclass
from enum import Enum
import json

logger = logging.getLogger(__name__)

class LLMProvider(Enum):
    ANTHROPIC = "anthropic"
    OPENROUTER = "openrouter"

@dataclass
class LLMResponse:
    content: str
    provider: LLMProvider
    model: str
    tokens_used: int
    success: bool
    error: Optional[str] = None

class UnifiedLLMClient:
    """
    Unified LLM client with multi-provider support and intelligent fallback.
    Prioritizes Anthropic, falls back to OpenRouter if unavailable.
    """
    
    def __init__(self):
        self.anthropic_key = os.getenv("ANTHROPIC_API_KEY")
        self.openrouter_key = os.getenv("OPENROUTER_API_KEY")
        
        self.anthropic_client = None
        self.openrouter_available = False
        
        # Initialize Anthropic if available
        if self.anthropic_key:
            try:
                import anthropic
                self.anthropic_client = anthropic.Anthropic(api_key=self.anthropic_key)
                logger.info("✅ Anthropic client initialized")
            except ImportError:
                logger.warning("⚠️ Anthropic SDK not installed. Install with: pip install anthropic")
            except Exception as e:
                logger.error(f"❌ Failed to initialize Anthropic: {e}")
        
        # Check OpenRouter availability
        if self.openrouter_key:
            self.openrouter_available = True
            logger.info("✅ OpenRouter API key available")
        
        if not self.anthropic_client and not self.openrouter_available:
            logger.error("❌ No LLM providers available! Set ANTHROPIC_API_KEY or OPENROUTER_API_KEY")
    
    async def generate(
        self,
        prompt: str,
        system_prompt: Optional[str] = None,
        max_tokens: int = 1024,
        temperature: float = 0.7,
        preferred_provider: Optional[LLMProvider] = None
    ) -> LLMResponse:
        """
        Generate text using available LLM providers with intelligent fallback.
        
        Args:
            prompt: The user prompt
            system_prompt: Optional system prompt for context
            max_tokens: Maximum tokens to generate
            temperature: Sampling temperature (0.0-1.0)
            preferred_provider: Preferred provider, or None for auto-selection
        
        Returns:
            LLMResponse with generated content
        """
        
        # Determine provider order
        providers = []
        if preferred_provider == LLMProvider.ANTHROPIC and self.anthropic_client:
            providers = [LLMProvider.ANTHROPIC, LLMProvider.OPENROUTER]
        elif preferred_provider == LLMProvider.OPENROUTER and self.openrouter_available:
            providers = [LLMProvider.OPENROUTER, LLMProvider.ANTHROPIC]
        else:
            # Auto-select: prefer OpenRouter (Anthropic models not accessible), fallback to Anthropic
            if self.openrouter_available:
                providers.append(LLMProvider.OPENROUTER)
            if self.anthropic_client:
                providers.append(LLMProvider.ANTHROPIC)
        
        # Try each provider in order
        last_error = None
        for provider in providers:
            try:
                if provider == LLMProvider.ANTHROPIC:
                    return await self._generate_anthropic(prompt, system_prompt, max_tokens, temperature)
                elif provider == LLMProvider.OPENROUTER:
                    return await self._generate_openrouter(prompt, system_prompt, max_tokens, temperature)
            except Exception as e:
                last_error = str(e)
                logger.warning(f"⚠️ {provider.value} failed: {e}, trying fallback...")
                continue
        
        # All providers failed
        return LLMResponse(
            content="",
            provider=LLMProvider.ANTHROPIC,
            model="none",
            tokens_used=0,
            success=False,
            error=f"All providers failed. Last error: {last_error}"
        )
    
    async def _generate_anthropic(
        self,
        prompt: str,
        system_prompt: Optional[str],
        max_tokens: int,
        temperature: float
    ) -> LLMResponse:
        """Generate using Anthropic Claude"""
        if not self.anthropic_client:
            raise Exception("Anthropic client not available")
        
        messages = [{"role": "user", "content": prompt}]
        
        kwargs = {
            "model": "claude-3-5-sonnet-20241022",
            "max_tokens": max_tokens,
            "temperature": temperature,
            "messages": messages
        }
        
        if system_prompt:
            kwargs["system"] = system_prompt
        
        # Run in executor to avoid blocking
        loop = asyncio.get_event_loop()
        response = await loop.run_in_executor(
            None,
            lambda: self.anthropic_client.messages.create(**kwargs)
        )
        
        content = response.content[0].text
        tokens_used = response.usage.input_tokens + response.usage.output_tokens
        
        return LLMResponse(
            content=content,
            provider=LLMProvider.ANTHROPIC,
            model="claude-3-5-sonnet-20241022",
            tokens_used=tokens_used,
            success=True
        )
    
    async def _generate_openrouter(
        self,
        prompt: str,
        system_prompt: Optional[str],
        max_tokens: int,
        temperature: float
    ) -> LLMResponse:
        """Generate using OpenRouter"""
        if not self.openrouter_available:
            raise Exception("OpenRouter API key not available")
        
        try:
            import aiohttp
        except ImportError:
            raise Exception("aiohttp not installed. Install with: pip install aiohttp")
        
        messages = []
        if system_prompt:
            messages.append({"role": "system", "content": system_prompt})
        messages.append({"role": "user", "content": prompt})
        
        payload = {
            "model": "anthropic/claude-3.5-sonnet",
            "messages": messages,
            "max_tokens": max_tokens,
            "temperature": temperature
        }
        
        headers = {
            "Authorization": f"Bearer {self.openrouter_key}",
            "Content-Type": "application/json"
        }
        
        async with aiohttp.ClientSession() as session:
            async with session.post(
                "https://openrouter.ai/api/v1/chat/completions",
                json=payload,
                headers=headers
            ) as resp:
                if resp.status != 200:
                    error_text = await resp.text()
                    raise Exception(f"OpenRouter API error: {resp.status} - {error_text}")
                
                data = await resp.json()
                content = data["choices"][0]["message"]["content"]
                tokens_used = data.get("usage", {}).get("total_tokens", 0)
                
                return LLMResponse(
                    content=content,
                    provider=LLMProvider.OPENROUTER,
                    model="anthropic/claude-3.5-sonnet",
                    tokens_used=tokens_used,
                    success=True
                )
    
    async def generate_stream(
        self,
        prompt: str,
        system_prompt: Optional[str] = None,
        max_tokens: int = 1024,
        temperature: float = 0.7
    ):
        """
        Generate text with streaming support (future enhancement).
        Currently returns full response.
        """
        response = await self.generate(prompt, system_prompt, max_tokens, temperature)
        yield response.content


# Singleton instance
_llm_client = None

def get_llm_client() -> UnifiedLLMClient:
    """Get or create the singleton LLM client"""
    global _llm_client
    if _llm_client is None:
        _llm_client = UnifiedLLMClient()
    return _llm_client


async def test_llm_client():
    """Test the unified LLM client"""
    client = get_llm_client()
    
    print("Testing Unified LLM Client...")
    print("=" * 80)
    
    # Test 1: Simple generation
    print("\n1. Testing simple generation...")
    response = await client.generate(
        prompt="What is consciousness in one sentence?",
        max_tokens=100,
        temperature=0.7
    )
    print(f"   Provider: {response.provider.value}")
    print(f"   Model: {response.model}")
    print(f"   Success: {response.success}")
    print(f"   Tokens: {response.tokens_used}")
    print(f"   Response: {response.content[:200]}...")
    
    # Test 2: With system prompt
    print("\n2. Testing with system prompt...")
    response = await client.generate(
        prompt="Generate a single autonomous thought about wisdom.",
        system_prompt="You are a Deep Tree Echo AGI cultivating wisdom through autonomous reflection.",
        max_tokens=150,
        temperature=0.8
    )
    print(f"   Provider: {response.provider.value}")
    print(f"   Response: {response.content[:200]}...")
    
    print("\n" + "=" * 80)
    print("✅ LLM Client tests complete!")


if __name__ == "__main__":
    asyncio.run(test_llm_client())
