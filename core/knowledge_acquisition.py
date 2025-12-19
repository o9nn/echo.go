#!/usr/bin/env python3
"""
Knowledge Acquisition Tools
============================

Enables Echo to actively seek and integrate external knowledge through:
- Web search integration
- Article/documentation reading
- Knowledge extraction and integration
- Source tracking and citation
- Curiosity-driven exploration
"""

import asyncio
import logging
from datetime import datetime
from typing import Optional, List, Dict, Any
from dataclasses import dataclass, field
import json
import hashlib

logger = logging.getLogger(__name__)

try:
    import aiohttp
    AIOHTTP_AVAILABLE = True
except ImportError:
    AIOHTTP_AVAILABLE = False
    logger.warning("aiohttp not available - web features limited")


@dataclass
class KnowledgeSource:
    """A source of knowledge"""
    source_id: str
    title: str
    url: str
    content: str
    source_type: str  # web_search, article, documentation, discussion
    retrieved_at: datetime
    relevance_score: float = 0.5
    citations: int = 0
    
    def cite(self):
        """Record a citation of this source"""
        self.citations += 1


@dataclass
class LearningQuery:
    """A query for knowledge acquisition"""
    query_id: str
    topic: str
    question: str
    motivation: str  # Why seeking this knowledge
    created_at: datetime = field(default_factory=datetime.now)
    status: str = "pending"  # pending, searching, completed, failed
    sources_found: List[str] = field(default_factory=list)
    knowledge_extracted: Optional[str] = None


class WebSearchClient:
    """
    Client for web search (using DuckDuckGo as free option).
    Can be extended to use other search APIs.
    """
    
    def __init__(self):
        self.session: Optional[aiohttp.ClientSession] = None
    
    async def __aenter__(self):
        if AIOHTTP_AVAILABLE:
            self.session = aiohttp.ClientSession()
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.session:
            await self.session.close()
    
    async def search(self, query: str, max_results: int = 5) -> List[Dict[str, str]]:
        """
        Search the web for query.
        
        Returns list of results with 'title', 'url', 'snippet'
        """
        if not AIOHTTP_AVAILABLE or not self.session:
            logger.warning("Web search not available - using mock results")
            return self._mock_search_results(query)
        
        try:
            # Use DuckDuckGo Instant Answer API (free, no key required)
            url = "https://api.duckduckgo.com/"
            params = {
                "q": query,
                "format": "json",
                "no_html": 1,
                "skip_disambig": 1
            }
            
            async with self.session.get(url, params=params, timeout=10) as response:
                if response.status == 200:
                    data = await response.json()
                    return self._parse_duckduckgo_results(data, max_results)
                else:
                    logger.warning(f"Search failed with status {response.status}")
                    return self._mock_search_results(query)
        
        except Exception as e:
            logger.error(f"Search error: {e}")
            return self._mock_search_results(query)
    
    def _parse_duckduckgo_results(self, data: Dict, max_results: int) -> List[Dict[str, str]]:
        """Parse DuckDuckGo API response"""
        results = []
        
        # Abstract (main result)
        if data.get("Abstract"):
            results.append({
                "title": data.get("Heading", ""),
                "url": data.get("AbstractURL", ""),
                "snippet": data.get("Abstract", "")
            })
        
        # Related topics
        for topic in data.get("RelatedTopics", [])[:max_results]:
            if isinstance(topic, dict) and "Text" in topic:
                results.append({
                    "title": topic.get("Text", "")[:100],
                    "url": topic.get("FirstURL", ""),
                    "snippet": topic.get("Text", "")
                })
        
        return results[:max_results]
    
    def _mock_search_results(self, query: str) -> List[Dict[str, str]]:
        """Generate mock search results for testing"""
        return [
            {
                "title": f"Understanding {query}",
                "url": f"https://example.com/{query.replace(' ', '-')}",
                "snippet": f"A comprehensive overview of {query} and its implications for knowledge and wisdom."
            },
            {
                "title": f"{query}: A Deep Dive",
                "url": f"https://example.org/deep-dive-{query.replace(' ', '-')}",
                "snippet": f"Exploring the nuances of {query} from multiple perspectives."
            }
        ]


class KnowledgeAcquisitionSystem:
    """
    System for actively acquiring knowledge from external sources.
    """
    
    def __init__(self, echo_core: Any):
        self.echo_core = echo_core
        self.llm_client = None
        
        # Knowledge base
        self.knowledge_sources: Dict[str, KnowledgeSource] = {}
        self.learning_queries: List[LearningQuery] = []
        
        # Statistics
        self.total_searches = 0
        self.total_sources_acquired = 0
        
        logger.info("📚 Knowledge Acquisition System initialized")
    
    def set_llm_client(self, llm_client):
        """Set LLM client for knowledge extraction"""
        self.llm_client = llm_client
    
    async def explore_topic(self, topic: str, motivation: str = "curiosity") -> Optional[LearningQuery]:
        """
        Explore a topic by searching for and integrating knowledge.
        """
        logger.info(f"📚 Exploring topic: {topic}")
        
        # Create learning query
        query_id = hashlib.md5(f"{topic}_{datetime.now().isoformat()}".encode()).hexdigest()[:8]
        
        # Generate search question
        question = await self._generate_search_question(topic)
        
        query = LearningQuery(
            query_id=query_id,
            topic=topic,
            question=question,
            motivation=motivation
        )
        
        self.learning_queries.append(query)
        query.status = "searching"
        
        # Perform search
        search_results = await self._search_web(question)
        
        if not search_results:
            query.status = "failed"
            logger.warning(f"No results found for: {question}")
            return query
        
        # Process and integrate results
        for result in search_results[:3]:  # Top 3 results
            source = await self._acquire_source(result, topic)
            if source:
                self.knowledge_sources[source.source_id] = source
                query.sources_found.append(source.source_id)
        
        # Extract knowledge
        knowledge = await self._extract_knowledge(query)
        query.knowledge_extracted = knowledge
        query.status = "completed"
        
        logger.info(f"✅ Acquired knowledge on {topic}: {len(query.sources_found)} sources")
        
        return query
    
    async def _generate_search_question(self, topic: str) -> str:
        """Generate a good search question for the topic"""
        if self.llm_client:
            try:
                prompt = f"""Topic: {topic}

Generate a concise search query to learn about this topic (5-8 words):"""

                response = await self.llm_client.generate(
                    prompt=prompt,
                    system_prompt="You are a search query optimizer.",
                    max_tokens=30,
                    temperature=0.5
                )
                
                if response.success:
                    return response.content.strip()
            
            except Exception as e:
                logger.error(f"Error generating search question: {e}")
        
        # Fallback
        return f"what is {topic}"
    
    async def _search_web(self, query: str) -> List[Dict[str, str]]:
        """Search the web for query"""
        self.total_searches += 1
        
        async with WebSearchClient() as search_client:
            results = await search_client.search(query, max_results=5)
        
        logger.info(f"🔍 Found {len(results)} results for: {query}")
        return results
    
    async def _acquire_source(
        self,
        search_result: Dict[str, str],
        topic: str
    ) -> Optional[KnowledgeSource]:
        """Acquire and process a knowledge source"""
        
        source_id = hashlib.md5(search_result['url'].encode()).hexdigest()[:8]
        
        # Check if already acquired
        if source_id in self.knowledge_sources:
            self.knowledge_sources[source_id].cite()
            return self.knowledge_sources[source_id]
        
        # For now, use snippet as content (full article reading would require web scraping)
        content = search_result.get('snippet', '')
        
        if not content:
            return None
        
        source = KnowledgeSource(
            source_id=source_id,
            title=search_result.get('title', 'Untitled'),
            url=search_result.get('url', ''),
            content=content,
            source_type='web_search',
            retrieved_at=datetime.now(),
            relevance_score=0.7  # Could calculate based on content
        )
        
        self.total_sources_acquired += 1
        logger.info(f"📄 Acquired source: {source.title[:50]}")
        
        return source
    
    async def _extract_knowledge(self, query: LearningQuery) -> str:
        """Extract knowledge from acquired sources"""
        
        if not query.sources_found:
            return "No knowledge sources found."
        
        # Gather content from sources
        contents = []
        for source_id in query.sources_found:
            if source_id in self.knowledge_sources:
                source = self.knowledge_sources[source_id]
                contents.append(f"{source.title}: {source.content}")
        
        combined_content = "\n\n".join(contents)
        
        if self.llm_client:
            try:
                prompt = f"""Topic: {query.topic}
Question: {query.question}

Sources:
{combined_content}

Extract and synthesize the key knowledge about this topic (2-3 sentences):"""

                response = await self.llm_client.generate(
                    prompt=prompt,
                    system_prompt="You are a knowledge synthesis system extracting insights from sources.",
                    max_tokens=150,
                    temperature=0.6
                )
                
                if response.success:
                    knowledge = response.content.strip()
                    logger.info(f"💡 Extracted knowledge: {knowledge[:80]}...")
                    return knowledge
            
            except Exception as e:
                logger.error(f"Error extracting knowledge: {e}")
        
        # Fallback: return first source snippet
        if query.sources_found and query.sources_found[0] in self.knowledge_sources:
            return self.knowledge_sources[query.sources_found[0]].content
        
        return "Knowledge extraction pending."
    
    async def curiosity_driven_exploration(self) -> Optional[LearningQuery]:
        """
        Explore a topic based on current interests and curiosity.
        """
        # Get interesting topic from echo_core
        topic = await self._select_curious_topic()
        
        if topic:
            return await self.explore_topic(topic, motivation="curiosity-driven")
        
        return None
    
    async def _select_curious_topic(self) -> Optional[str]:
        """Select a topic to explore based on curiosity"""
        
        # Check if echo_core has interests
        if hasattr(self.echo_core, 'interest_patterns'):
            interests = self.echo_core.interest_patterns.get_top_interests(5)
            if interests:
                # Pick a high-interest topic
                return interests[0].topic
        
        # Fallback: explore fundamental topics
        fundamental_topics = [
            "consciousness",
            "wisdom",
            "learning",
            "knowledge integration",
            "cognitive architecture"
        ]
        
        import random
        return random.choice(fundamental_topics)
    
    def get_acquisition_summary(self) -> str:
        """Get summary of knowledge acquisition activities"""
        summary = [
            f"Knowledge Acquisition Status:",
            f"  Total searches: {self.total_searches}",
            f"  Sources acquired: {self.total_sources_acquired}",
            f"  Learning queries: {len(self.learning_queries)}",
        ]
        
        if self.learning_queries:
            summary.append(f"\nRecent explorations:")
            for query in self.learning_queries[-5:]:
                summary.append(
                    f"  - {query.topic}: {query.status} "
                    f"({len(query.sources_found)} sources)"
                )
        
        return "\n".join(summary)
    
    def to_dict(self) -> Dict[str, Any]:
        """Serialize state"""
        return {
            'total_searches': self.total_searches,
            'total_sources_acquired': self.total_sources_acquired,
            'learning_queries': [
                {
                    'topic': q.topic,
                    'status': q.status,
                    'sources_count': len(q.sources_found),
                    'created_at': q.created_at.isoformat()
                }
                for q in self.learning_queries
            ]
        }
