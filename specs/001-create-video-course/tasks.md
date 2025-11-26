# Tasks: Create Video Course Creator

**Input**: Design documents from `/specs/001-create-video-course/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md

**Tests**: Tests are REQUIRED - 100% coverage as per constitution and spec.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md: Multi-project structure with core-processor/, creator-app/, player-app/, etc.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create project structure per implementation plan (core-processor/, creator-app/, etc.)
- [ ] T002 Initialize Python project in core-processor/ with pyproject.toml
- [ ] T003 Initialize Node.js projects in creator-app/, player-app/, mobile-player/
- [ ] T004 [P] Configure linting and formatting tools (black, eslint, prettier)
- [ ] T005 Setup shared TypeScript types in shared/types/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T006 Setup MCP SDK and base server framework in core-processor/src/mcp_servers/
- [ ] T007 [P] Implement Bark TTS MCP server in core-processor/src/mcp_servers/bark_server.py
- [ ] T008 [P] Implement SpeechT5 TTS MCP server in core-processor/src/mcp_servers/speecht5_server.py
- [ ] T009 [P] Implement Suno music generation MCP server in core-processor/src/mcp_servers/suno_server.py
- [ ] T010 [P] Implement LLaVA image analysis MCP server in core-processor/src/mcp_servers/llava_server.py
- [ ] T011 [P] Implement Pix2Struct UI parsing MCP server in core-processor/src/mcp_servers/pix2struct_server.py
- [ ] T012 Create course generation pipeline framework in core-processor/src/pipeline/
- [ ] T013 Setup SQLite database schema for course metadata in core-processor/src/models/
- [ ] T014 Implement markdown parsing utilities in core-processor/src/utils/markdown_parser.py
- [ ] T015 Setup FFmpeg integration for video/audio processing in core-processor/src/utils/ffmpeg.py
- [ ] T016 Configure environment and configuration management

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Markdown to Video Conversion (Priority: P1) 🎯 MVP

**Goal**: Convert markdown scripts to basic video courses with text-to-speech

**Independent Test**: Upload markdown file and verify video output with narration

### Tests for User Story 1 ⚠️

- [ ] T017 [P] [US1] Contract test for markdown parsing in core-processor/tests/contract/test_markdown_parsing.py
- [ ] T018 [P] [US1] Integration test for video generation pipeline in core-processor/tests/integration/test_video_generation.py

### Implementation for User Story 1

- [ ] T019 [US1] Create Course and Lesson data models in core-processor/src/models/
- [ ] T020 [US1] Implement basic TTS integration in core-processor/src/pipeline/tts_processor.py
- [ ] T021 [US1] Create video assembly from audio and text overlays in core-processor/src/pipeline/video_assembler.py
- [ ] T022 [US1] Implement course generation orchestrator in core-processor/src/pipeline/course_generator.py
- [ ] T023 [US1] Add error handling and validation for US1

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Visual Enhancements (Priority: P1)

**Goal**: Add colorful backgrounds, illustrations, diagrams to videos

**Independent Test**: Generate course and verify visual elements appear correctly

### Tests for User Story 2 ⚠️

- [ ] T024 [P] [US2] Contract test for visual processing in core-processor/tests/contract/test_visual_processing.py
- [ ] T025 [P] [US2] Integration test for enhanced video output in core-processor/tests/integration/test_visual_enhancement.py

### Implementation for User Story 2

- [ ] T026 [US2] Implement background generation system in core-processor/src/pipeline/background_generator.py
- [ ] T027 [US2] Add diagram/illustration embedding in core-processor/src/pipeline/diagram_processor.py
- [ ] T028 [US2] Update video assembler to include visual enhancements
- [ ] T029 [US2] Integrate with US1 pipeline

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Audio Features (Priority: P1)

**Goal**: Add background music and high-quality speech narration

**Independent Test**: Verify audio quality and music integration in generated videos

### Tests for User Story 3 ⚠️

- [ ] T030 [P] [US3] Contract test for audio processing in core-processor/tests/contract/test_audio_processing.py
- [ ] T031 [P] [US3] Integration test for audio-enhanced courses in core-processor/tests/integration/test_audio_features.py

### Implementation for User Story 3

- [ ] T032 [US3] Implement music selection and mixing in core-processor/src/pipeline/music_processor.py
- [ ] T033 [US3] Enhance TTS with voice options in core-processor/src/pipeline/enhanced_tts.py
- [ ] T034 [US3] Update video assembler for audio mixing
- [ ] T035 [US3] Integrate audio features with existing pipeline

**Checkpoint**: All P1 user stories should now be independently functional

---

## Phase 6: User Story 4 - Embedded Materials (Priority: P2)

**Goal**: Embed video clips, images, and code snippets in courses

**Independent Test**: Include media references in markdown and verify embedding

### Tests for User Story 4 ⚠️

- [ ] T036 [P] [US4] Contract test for media embedding in core-processor/tests/contract/test_media_embedding.py
- [ ] T037 [P] [US4] Integration test for rich content courses in core-processor/tests/integration/test_embedded_materials.py

### Implementation for User Story 4

- [ ] T038 [US4] Implement image/video embedding processor in core-processor/src/pipeline/media_embedder.py
- [ ] T039 [US4] Add code syntax highlighting in core-processor/src/pipeline/code_highlighter.py
- [ ] T040 [US4] Update markdown parser for media references
- [ ] T041 [US4] Integrate with video assembler

---

## Phase 7: User Story 5 - Interactive Code Sessions (Priority: P2)

**Goal**: Include interactive coding exercises in courses

**Independent Test**: Verify interactive elements are present and functional

### Tests for User Story 5 ⚠️

- [ ] T042 [P] [US5] Contract test for interactive elements in core-processor/tests/contract/test_interactive_elements.py
- [ ] T043 [P] [US5] Integration test for interactive courses in core-processor/tests/integration/test_interactive_sessions.py

### Implementation for User Story 5

- [ ] T044 [US5] Create interactive code session generator in core-processor/src/pipeline/interactive_generator.py
- [ ] T045 [US5] Implement code execution environment in core-processor/src/pipeline/code_executor.py
- [ ] T046 [US5] Add interactive UI components to player
- [ ] T047 [US5] Integrate with course structure

---

## Phase 8: User Story 6 - Multi-language Subtitles (Priority: P3)

**Goal**: Generate subtitles in multiple languages

**Independent Test**: Verify subtitle generation and accuracy

### Tests for User Story 6 ⚠️

- [ ] T048 [P] [US6] Contract test for subtitle generation in core-processor/tests/contract/test_subtitle_generation.py
- [ ] T049 [P] [US6] Integration test for multilingual courses in core-processor/tests/integration/test_multilang_subtitles.py

### Implementation for User Story 6

- [ ] T050 [US6] Implement translation service integration in core-processor/src/pipeline/translation_service.py
- [ ] T051 [US6] Create subtitle generator in core-processor/src/pipeline/subtitle_generator.py
- [ ] T052 [US6] Add subtitle embedding to video assembler
- [ ] T053 [US6] Support multiple language selection

---

## Phase 9: User Story 7 - Cross-platform Playback (Priority: P3)

**Goal**: Ensure courses play on all platforms

**Independent Test**: Test playback on desktop, web, and mobile

### Tests for User Story 7 ⚠️

- [ ] T054 [P] [US7] Contract test for cross-platform compatibility in tests/contract/test_cross_platform.py
- [ ] T055 [P] [US7] Integration test for multi-platform playback in tests/integration/test_platform_playback.py

### Implementation for User Story 7

- [ ] T056 [US7] Implement web player in player-app/src/components/VideoPlayer.tsx
- [ ] T057 [US7] Create Electron desktop wrapper in creator-app/
- [ ] T058 [US7] Develop React Native mobile player in mobile-player/src/
- [ ] T059 [US7] Add platform detection and adaptive UI
- [ ] T060 [US7] Test on Windows, macOS, Linux, iOS, Android

---

## Phase 11: User Story 8 - Flexible LLM Integration (Priority: P2)

**Goal**: Support both free and paid LLM services with seamless switching

**Independent Test**: Configure different LLM providers and verify consistent functionality

### Tests for User Story 8 ⚠️

- [ ] T061 [P] [US8] Contract test for LLM provider abstraction in core-processor/tests/contract/test_llm_providers.py
- [ ] T062 [P] [US8] Integration test for provider switching in core-processor/tests/integration/test_llm_switching.py

### Implementation for User Story 8

- [ ] T063 [US8] Create LLM provider abstraction layer in core-processor/src/llm/providers.go
- [ ] T064 [US8] Implement free LLM integrations (local models, public APIs)
- [ ] T065 [US8] Implement paid LLM integrations (OpenAI, Anthropic, etc.)
- [ ] T066 [US8] Add provider configuration and switching logic
- [ ] T067 [US8] Implement fallback mechanisms for service failures
- [ ] T068 [US8] Add usage tracking and cost monitoring

---

## Phase 12: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T069 [P] Create comprehensive documentation in docs/
- [ ] T070 [P] Write user tutorials and manuals
- [ ] T071 [P] Implement API documentation
- [ ] T072 Code cleanup and performance optimization
- [ ] T073 [P] Add remaining unit tests to achieve 100% coverage
- [ ] T074 Security hardening and privacy protection
- [ ] T075 Run final validation and testing

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-11)**: All depend on Foundational phase completion
  - P1 stories can proceed in parallel
  - P2 stories (including US8) after P1 completion
  - P3 stories after P2 completion
- **Polish (Phase 12)**: Depends on all user stories being complete

### Parallel Opportunities

- All MCP server implementations can run in parallel
- User stories within same priority can be developed in parallel
- Tests for different components can run in parallel
- UI development for different platforms can proceed in parallel

---

## Implementation Strategy

### MVP First (P1 Stories Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phases 3-5: All P1 stories
4. **STOP and VALIDATE**: Test complete basic course generation
5. Deploy/demo MVP

### Full Implementation

1. Complete Setup + Foundational → Foundation ready
2. Implement P1 stories → Basic functional product
3. Add P2 stories → Enhanced features
4. Add P3 stories → Complete product
5. Polish → Production ready

---

## Notes

- Each user story designed for independent implementation and testing
- 100% test coverage required across all components
- Focus on cross-platform compatibility throughout
- Regular commits and validation at checkpoints
- Documentation and tutorials created alongside code