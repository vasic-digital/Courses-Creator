# Research: AI Models for Multimedia Course Generation

## Overview
This research covers LLMs and tools for generating video courses from markdown scripts, focusing on text-to-speech, music, image analysis, and video creation. All models will be integrated via Model Context Protocol (MCP) for secure, standardized access from the opencode coding agent.

## Model Research

### VideoPoet (Google)
- **Description**: Multimodal model for text-to-video generation using world model approach
- **Capabilities**: Creates coherent, open-ended videos from text prompts with high fidelity
- **Access**: Limited; requires research partnerships or demos (not fully open-source)
- **Integration**: Custom MCP server using Google's API/research endpoints
- **Use Case**: Generate video segments for course sections

### SONIQUE/Suno AI Music Generation
- **Description**: AI music generation tools for creating background tracks
- **Capabilities**: Generates instrumental music from text prompts
- **Access**: Via Suno AI platform or open-source alternatives
- **Integration**: MCP server with Suno API calls
- **Use Case**: Automatic background music selection and mixing

### Suno-ai Bark (TTS)
- **Description**: Open-source TTS model for realistic speech generation
- **Capabilities**: Multi-language speech, voice presets, long-form audio
- **Access**: GitHub (suno-ai/bark), Hugging Face
- **Integration**: Local MCP server with bark-tts library
- **Use Case**: High-quality narration from markdown text

### Microsoft SpeechT5 (TTS)
- **Description**: Advanced TTS with voice cloning and multilingual support
- **Capabilities**: High-quality speech synthesis, customizable voices
- **Access**: Azure Speech Services or Hugging Face
- **Integration**: MCP server using transformers pipeline
- **Use Case**: Alternative/complementary TTS for varied voice options

### LLaVA (Image Analysis)
- **Description**: Vision-language model for image understanding
- **Capabilities**: Visual Q&A, captioning, instruction following
- **Access**: GitHub (haotian-liu/LLaVA), Hugging Face
- **Integration**: Local MCP server for image description
- **Use Case**: Analyze UI screenshots or embedded images in courses

### Pix2Struct (UI/Image Parsing)
- **Description**: Model for parsing structured data from images
- **Capabilities**: Converts images to text/HTML representations
- **Access**: Hugging Face (google/pix2struct-base)
- **Integration**: MCP server for screenshot analysis
- **Use Case**: Extract text/code from UI images for interactive elements

## MCP Integration Strategy

### Architecture
- **MCP Servers**: Each AI model runs as a separate MCP server process
- **Client**: Opencode acts as MCP client, calling tools via standardized protocol
- **Security**: Isolated server processes prevent direct model access
- **Workflow**: Parse markdown → Call MCP tools → Assemble multimedia → Generate course

### Implementation Details
- Use MCP Python SDK for server development
- Models loaded locally where possible (open-source)
- API-based access for proprietary models
- Error handling for model failures/timeouts
- Caching for repeated requests

### Code Examples
See attached code snippets for each MCP server implementation, including tool definitions and basic usage patterns.

## Recommendations
1. Start with open-source models (Bark, LLaVA, Pix2Struct) for core functionality
2. Use API-based models (Suno, SpeechT5) for enhanced features
3. Implement VideoPoet last due to access limitations
4. Focus on local processing to maintain privacy
5. Test integrations incrementally with small inputs

## Risks & Mitigations
- **Model Access**: Use open-source alternatives where proprietary models unavailable
- **Performance**: Optimize model loading and caching
- **Quality**: Implement quality checks and fallbacks
- **Privacy**: Ensure no user data sent to external APIs without consent