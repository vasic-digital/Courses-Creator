"""
Base MCP Server implementation for Course Creator AI integrations.

Note: This is a placeholder until MCP SDK is properly installed.
Actual implementation will use the MCP SDK for tool registration.
"""

from abc import ABC, abstractmethod
from typing import Any, Callable, Dict, List


class BaseMCPServer(ABC):
    """Base class for MCP servers in Course Creator."""

    def __init__(self, name: str):
        self.name = name
        # self.app = MCPApp()  # Placeholder

    @abstractmethod
    def register_tools(self) -> None:
        """Register MCP tools for this server."""
        pass

    def run(self) -> None:
        """Run the MCP server."""
        print(f"Starting {self.name} MCP server...")
        self.register_tools()
        # self.app.run()  # Placeholder

    def add_tool(self, tool_func: Callable, name: str, description: str) -> None:
        """Add a tool to the MCP app."""
        print(f"Registering tool: {name} - {description}")
        # self.app.tool(name=name, description=description)(tool_func)  # Placeholder


class AIProcessingError(Exception):
    """Error raised when AI processing fails."""
    pass


class ModelLoadError(Exception):
    """Error raised when model loading fails."""
    pass