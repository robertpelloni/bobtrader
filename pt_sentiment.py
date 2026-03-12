import logging
import json
import requests
from typing import Optional, Dict, Any

class SentimentAnalyzer:
    """
    Handles fetching and normalizing sentiment data from external MCP servers.
    Default setup assumes a local MCP server that provides generic Fear/Greed
    and specific coin sentiment.
    """
    def __init__(self, config=None):
        self.config = config

    def _make_mcp_request(self, endpoint: str, payload: dict = None) -> Optional[dict]:
        """
        Generic helper to connect to an MCP server exposed via HTTP.
        """
        if not self.config or not self.config.mcp_url:
            return None
            
        url = f"{self.config.mcp_url.rstrip('/')}/{endpoint.lstrip('/')}"
        try:
            if payload:
                response = requests.post(url, json=payload, timeout=5)
            else:
                response = requests.get(url, timeout=5)
                
            if response.status_code == 200:
                return response.json()
            else:
                logging.warning(f"MCP server returned {response.status_code} for {endpoint}")
        except requests.exceptions.RequestException as e:
            logging.debug(f"Failed to connect to MCP Sentiment server: {e}")
            
        return None

    def get_global_fear_greed_index(self) -> float:
        """
        Fetches the global fear and greed index.
        Returns a normalized score:
          -1.0 = Extreme Fear
           0.0 = Neutral
           1.0 = Extreme Greed
        """
        if not self.config or not self.config.fear_greed_enabled:
            return 0.0
            
        data = self._make_mcp_request("/api/fear-greed")
        if data and "score" in data:
            # Assuming MCP returns 0-100 (Cryptocurrency Fear & Greed Index standard)
            # 0 = Extreme Fear, 100 = Extreme Greed
            raw_score = float(data["score"])
            normalized = (raw_score - 50.0) / 50.0
            return normalized
            
        return 0.0

    def get_coin_sentiment(self, symbol: str) -> float:
        """
        Fetches the social/news sentiment specifically for a given coin via MCP.
        Returns a normalized score from -1.0 (very negative) to 1.0 (very positive).
        """
        if not self.config or not self.config.social_sentiment_enabled:
            return 0.0
            
        data = self._make_mcp_request("/api/coin-sentiment", payload={"symbol": symbol})
        if data and "sentiment_score" in data:
             # Assuming MCP returns a normalized score -1.0 to 1.0 directly
             # Or adjust normalization logic based on the actual MCP payload
             return float(data["sentiment_score"])
             
        return 0.0

    def calculate_combined_sentiment(self, symbol: str) -> float:
        """
        Calculates a blended sentiment score using both global market context and
        coin-specific news/social data.
        
        Returns the blended score multiplied by the user's `impact_multiplier`.
        """
        if not self.config or not self.config.enabled:
            return 0.0
            
        global_fg = self.get_global_fear_greed_index()
        coin_score = self.get_coin_sentiment(symbol)
        
        # Blend logic: 70% Coin-specific sentiment, 30% Global Market trend
        blended = (coin_score * 0.70) + (global_fg * 0.30)
        
        # Apply the user's impact multiplier (0.0 to 5.0)
        final_score = blended * self.config.impact_multiplier
        
        # Bound it strictly between -1.0 and 1.0 after multiplier, 
        # so it acts cleanly as a signal modifier downstream.
        return max(-1.0, min(1.0, final_score))

if __name__ == "__main__":
    from dataclasses import dataclass
    
    @dataclass
    class MockConfig:
        enabled: bool = True
        fear_greed_enabled: bool = True
        social_sentiment_enabled: bool = True
        mcp_url: str = "http://localhost:8000"
        impact_multiplier: float = 1.0

    print("Testing Sentiment Analyzer Module (Standalone)")
    analyzer = SentimentAnalyzer(MockConfig())
    score = analyzer.calculate_combined_sentiment("BTC")
    print(f"Blended BTC Sentiment Score (Assuming API unreachable): {score}")
