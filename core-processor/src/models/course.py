"""
Data models for courses and lessons.
"""

from dataclasses import dataclass, field
from typing import List, Dict, Any, Optional
from datetime import datetime


@dataclass
class CourseMetadata:
    """Metadata for a course."""
    author: str
    language: str
    tags: List[str]
    thumbnail_url: Optional[str] = None
    total_duration: int = 0


@dataclass
class Subtitle:
    """Subtitle information."""
    language: str
    content: str
    timestamps: List[Dict[str, Any]]


@dataclass
class InteractiveElement:
    """Interactive element in a lesson."""
    id: str
    type: str  # 'code', 'quiz', 'exercise'
    content: str
    position: int  # seconds into video


@dataclass
class Lesson:
    """A lesson within a course."""
    id: str
    title: str
    content: str
    video_url: Optional[str] = None
    audio_url: Optional[str] = None
    subtitles: List[Subtitle] = field(default_factory=list)
    interactive_elements: List[InteractiveElement] = field(default_factory=list)
    duration: int = 0
    order: int = 0


@dataclass
class Course:
    """A complete video course."""
    id: str
    title: str
    description: str
    lessons: List[Lesson]
    metadata: CourseMetadata
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None


@dataclass
class ProcessingOptions:
    """Options for course processing."""
    voice: Optional[str] = None
    background_music: bool = True
    languages: List[str] = field(default_factory=lambda: ["en"])
    quality: str = "standard"  # 'standard' or 'high'


@dataclass
class ProcessingResult:
    """Result of course processing."""
    course: Course
    output_path: str
    duration: int
    status: str  # 'success' or 'failed'
    errors: Optional[List[str]] = None