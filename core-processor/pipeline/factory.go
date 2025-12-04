package pipeline

import (
	"github.com/course-creator/core-processor/config"
	storage "github.com/course-creator/core-processor/filestorage"
	"github.com/course-creator/core-processor/llm"
)

// PipelineFactory creates configured pipeline components
type PipelineFactory struct {
	config *config.Config
}

// NewPipelineFactory creates a new pipeline factory
func NewPipelineFactory(cfg *config.Config) *PipelineFactory {
	return &PipelineFactory{
		config: cfg,
	}
}

// NewCourseGenerator creates a new course generator with configuration
func (pf *PipelineFactory) NewCourseGenerator() *CourseGenerator {
	// Create storage manager with configuration
	storageConfigs := make(map[string]storage.StorageConfig)
	for name, cfg := range pf.config.Storage {
		storageConfigs[name] = storage.StorageConfig{
			Type:      cfg.Type,
			BasePath:  cfg.BasePath,
			PublicURL: cfg.PublicURL,
			Settings:  cfg.Settings,
		}
	}

	storageManager, err := storage.NewStorageManager(storageConfigs)
	if err != nil {
		// Fallback to default local storage
		defaultConfig := storage.DefaultStorageConfig()
		storageManager, _ = storage.NewStorageManagerWithDefault(defaultConfig)
	}

	// Create TTS processor with configuration
	ttsConfig := TTSConfig{
		DefaultProvider: TTSProvider(pf.config.TTS.Provider),
		OutputDir:       "/tmp/course_audio",
		SampleRate:      24000,
		BitRate:         128000,
		Format:          "wav",
		Timeout:         pf.config.TTS.Timeout,
		MaxRetries:      3,
		ChunkSize:       200,
		Parallelism:     2,
	}

	ttsProcessor := NewTTSProcessorWithConfig(ttsConfig)

	// Create other components
	videoAssembler := NewVideoAssembler(storageManager.DefaultProvider())
	diagramProcessor := NewDiagramProcessor(storageManager.DefaultProvider())

	// Create LLM content generator with configuration
	contentGen := llm.NewCourseContentGenerator(&pf.config.LLM)

	return &CourseGenerator{
		ttsProcessor:     ttsProcessor,
		videoAssembler:   videoAssembler,
		diagramProcessor: diagramProcessor,
		contentGen:       contentGen,
		storage:          storageManager,
	}
}

// GetLLMManager returns an LLM provider manager with configuration
func (pf *PipelineFactory) GetLLMManager() *llm.ProviderManager {
	return llm.NewProviderManager(&pf.config.LLM)
}
