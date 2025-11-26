"""
Main course generation orchestrator.
"""

from typing import Dict, Any, List
from pathlib import Path
import asyncio
from ..models.course import Course, Lesson
from ..utils.markdown_parser import MarkdownParser
from .tts_processor import TTSProcessor
from .video_assembler import VideoAssembler


class CourseGenerator:
    """Orchestrates the entire course generation process."""

    def __init__(self):
        self.markdown_parser = MarkdownParser()
        self.tts_processor = TTSProcessor()
        self.video_assembler = VideoAssembler()

    async def generate_course(self, markdown_path: str, output_dir: str, options: Dict[str, Any]) -> Course:
        """
        Generate a complete video course from markdown.

        Args:
            markdown_path: Path to markdown file
            output_dir: Directory to save outputs
            options: Generation options

        Returns:
            Generated Course object
        """
        print(f"Starting course generation from {markdown_path}")

        # Parse markdown
        course_data = self.markdown_parser.parse(markdown_path)

        # Create course object
        course = Course(
            id=f"course_{Path(markdown_path).stem}",
            title=course_data.get("title", "Generated Course"),
            description=course_data.get("description", ""),
            lessons=[],
            metadata=course_data.get("metadata", {}),
            created_at=None,
            updated_at=None
        )

        # Generate lessons
        lessons = []
        for section in course_data.get("sections", []):
            lesson = await self._generate_lesson(section, output_dir, options)
            lessons.append(lesson)

        course.lessons = lessons

        # Assemble final course
        await self._assemble_course(course, output_dir, options)

        return course

    async def _generate_lesson(self, section: Dict[str, Any], output_dir: str, options: Dict[str, Any]) -> Lesson:
        """Generate a single lesson from section data."""
        print(f"Generating lesson: {section.get('title', 'Untitled')}")

        # Generate TTS audio
        audio_path = await self.tts_processor.generate_audio(section["content"], options)

        # Create video
        video_path = await self.video_assembler.create_video(
            audio_path=audio_path,
            text_content=section["content"],
            output_dir=output_dir,
            options=options
        )

        lesson = Lesson(
            id=f"lesson_{hash(section['content'])}",
            title=section.get("title", "Lesson"),
            content=section["content"],
            video_url=video_path,
            audio_url=audio_path,
            subtitles=[],  # Will be added later
            interactive_elements=[],  # Will be added later
            duration=0,  # Calculate from audio/video
            order=section.get("order", 0)
        )

        return lesson

    async def _assemble_course(self, course: Course, output_dir: str, options: Dict[str, Any]) -> None:
        """Assemble the final course package."""
        print("Assembling final course package")

        # Placeholder for course assembly
        # - Generate course index
        # - Create player files
        # - Package everything

        pass