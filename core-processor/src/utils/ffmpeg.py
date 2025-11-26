"""
FFmpeg utilities for video and audio processing.
"""

import subprocess
import asyncio
from typing import List, Optional
from pathlib import Path


class FFmpegProcessor:
    """Wrapper for FFmpeg operations."""

    @staticmethod
    async def run_command(cmd: List[str]) -> str:
        """Run FFmpeg command asynchronously."""
        process = await asyncio.create_subprocess_exec(
            *cmd,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        stdout, stderr = await process.communicate()

        if process.returncode != 0:
            raise RuntimeError(f"FFmpeg failed: {stderr.decode()}")

        return stdout.decode()

    @staticmethod
    async def combine_audio_video(
        video_path: str,
        audio_path: str,
        output_path: str,
        subtitles_path: Optional[str] = None
    ) -> str:
        """Combine video and audio files."""
        cmd = ["ffmpeg", "-i", video_path, "-i", audio_path]

        if subtitles_path:
            cmd.extend(["-i", subtitles_path])

        cmd.extend(["-c:v", "copy", "-c:a", "aac", output_path])

        await FFmpegProcessor.run_command(cmd)
        return output_path

    @staticmethod
    async def add_text_overlay(
        input_path: str,
        output_path: str,
        text: str,
        font_size: int = 24
    ) -> str:
        """Add text overlay to video."""
        cmd = [
            "ffmpeg", "-i", input_path,
            "-vf", f"drawtext=text='{text}':fontsize={font_size}:fontcolor=white",
            "-c:a", "copy", output_path
        ]

        await FFmpegProcessor.run_command(cmd)
        return output_path

    @staticmethod
    async def mix_audio(
        audio1_path: str,
        audio2_path: str,
        output_path: str,
        audio1_volume: float = 1.0,
        audio2_volume: float = 0.3
    ) -> str:
        """Mix two audio files."""
        cmd = [
            "ffmpeg", "-i", audio1_path, "-i", audio2_path,
            "-filter_complex",
            f"[0:a]volume={audio1_volume}[a];[1:a]volume={audio2_volume}[b];[a][b]amix=inputs=2:duration=first",
            "-c:a", "aac", output_path
        ]

        await FFmpegProcessor.run_command(cmd)
        return output_path