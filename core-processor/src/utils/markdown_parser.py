"""
Markdown parsing utilities for course content.
"""

import re
from typing import Dict, Any, List
from pathlib import Path


class MarkdownParser:
    """Parser for markdown course content."""

    def __init__(self):
        self.header_pattern = re.compile(r'^#+\s+(.+)$', re.MULTILINE)
        self.code_pattern = re.compile(r'```(\w+)?\n(.*?)\n```', re.DOTALL)
        self.image_pattern = re.compile(r'!\[([^\]]*)\]\(([^)]+)\)')

    def parse(self, file_path: str) -> Dict[str, Any]:
        """
        Parse markdown file into course structure.

        Args:
            file_path: Path to markdown file

        Returns:
            Parsed course data
        """
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Extract title from first header
        title_match = self.header_pattern.search(content)
        title = title_match.group(1).strip() if title_match else "Untitled Course"

        # Extract sections (split by level 1 headers)
        sections = self._extract_sections(content)

        # Extract metadata from frontmatter or comments
        metadata = self._extract_metadata(content)

        return {
            "title": title,
            "description": self._extract_description(content),
            "sections": sections,
            "metadata": metadata
        }

    def _extract_sections(self, content: str) -> List[Dict[str, Any]]:
        """Extract sections from markdown content."""
        sections = []
        lines = content.split('\n')
        current_section = None
        current_content = []
        order = 0

        for line in lines:
            if line.startswith('# '):
                # Save previous section
                if current_section:
                    sections.append({
                        "title": current_section,
                        "content": '\n'.join(current_content).strip(),
                        "order": order
                    })
                    order += 1

                # Start new section
                current_section = line[2:].strip()
                current_content = []
            else:
                if current_section:
                    current_content.append(line)

        # Add last section
        if current_section:
            sections.append({
                "title": current_section,
                "content": '\n'.join(current_content).strip(),
                "order": order
            })

        return sections

    def _extract_description(self, content: str) -> str:
        """Extract course description from content."""
        # Look for content before first header
        first_header = self.header_pattern.search(content)
        if first_header:
            description = content[:first_header.start()].strip()
            return description
        return ""

    def _extract_metadata(self, content: str) -> Dict[str, Any]:
        """Extract metadata from frontmatter or comments."""
        # Placeholder for metadata extraction
        # Could parse YAML frontmatter, author comments, etc.
        return {
            "author": "Unknown",
            "language": "en",
            "tags": []
        }