package performance_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/course-creator/core-processor/llm"
	"github.com/course-creator/core-processor/models"
)

func BenchmarkOpenAIProvider_GenerateText(b *testing.B) {
	provider := llm.NewOpenAIProvider("test-key", "gpt-3.5-turbo")
	ctx := context.Background()
	prompt := "What is machine learning?"
	options := models.ProcessingOptions{Quality: "standard"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.GenerateText(ctx, prompt, options)
	}
}

func BenchmarkAnthropicProvider_GenerateText(b *testing.B) {
	provider := llm.NewAnthropicProvider("test-key", "claude-3-haiku-20240307")
	ctx := context.Background()
	prompt := "What is machine learning?"
	options := models.ProcessingOptions{Quality: "standard"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.GenerateText(ctx, prompt, options)
	}
}

func BenchmarkOllamaProvider_GenerateText(b *testing.B) {
	provider := llm.NewOllamaProvider("http://localhost:11434", "llama2")
	ctx := context.Background()
	prompt := "What is machine learning?"
	options := models.ProcessingOptions{Quality: "standard"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.GenerateText(ctx, prompt, options)
	}
}

func BenchmarkProviderManager_GenerateWithFallback(b *testing.B) {
	manager := llm.NewProviderManager(nil)

	// Add mock providers for benchmarking
	mockProvider1 := llm.NewFreeProvider("Mock1", "", "")
	mockProvider2 := llm.NewFreeProvider("Mock2", "", "")
	manager.RegisterProvider(mockProvider1)
	manager.RegisterProvider(mockProvider2)

	ctx := context.Background()
	prompt := "What is machine learning?"
	options := models.ProcessingOptions{Quality: "standard"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.GenerateWithFallback(ctx, prompt, options)
	}
}

func BenchmarkProviderManager_GetBestProvider(b *testing.B) {
	manager := llm.NewProviderManager(nil)

	// Add multiple providers
	for i := 0; i < 10; i++ {
		provider := llm.NewFreeProvider(fmt.Sprintf("Provider%d", i), "", "")
		manager.RegisterProvider(provider)
	}

	preferences := llm.ProviderPreferences{
		PreferredType:     llm.ProviderTypeFree,
		MaxCostPerRequest: 0.10,
		PrioritizeQuality: false,
		AllowPaid:         false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetBestProvider(preferences)
	}
}

func TestLLMPerformance(t *testing.T) {
	tests := []struct {
		name         string
		testFunc     func(b *testing.B)
		minOpsPerSec float64
		maxMemMB     float64
	}{
		{
			name:         "OpenAIProvider_GenerateText",
			testFunc:     BenchmarkOpenAIProvider_GenerateText,
			minOpsPerSec: 1000,
			maxMemMB:     10,
		},
		{
			name:         "AnthropicProvider_GenerateText",
			testFunc:     BenchmarkAnthropicProvider_GenerateText,
			minOpsPerSec: 1000,
			maxMemMB:     10,
		},
		{
			name:         "OllamaProvider_GenerateText",
			testFunc:     BenchmarkOllamaProvider_GenerateText,
			minOpsPerSec: 1000,
			maxMemMB:     10,
		},
		{
			name:         "ProviderManager_GenerateWithFallback",
			testFunc:     BenchmarkProviderManager_GenerateWithFallback,
			minOpsPerSec: 500,
			maxMemMB:     20,
		},
		{
			name:         "ProviderManager_GetBestProvider",
			testFunc:     BenchmarkProviderManager_GetBestProvider,
			minOpsPerSec: 10000,
			maxMemMB:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Record initial memory
			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			// Run benchmark
			result := testing.Benchmark(tt.testFunc)

			// Record final memory
			runtime.GC()
			runtime.ReadMemStats(&m2)

			// Calculate memory usage
			memUsed := (m2.Alloc - m1.Alloc) / 1024 / 1024 // MB
			opsPerSec := float64(result.N) / result.T.Seconds()

			t.Logf("Performance for %s:", tt.name)
			t.Logf("  Operations: %d", result.N)
			t.Logf("  Time: %v", result.T)
			t.Logf("  Ops/sec: %.2f", opsPerSec)
			t.Logf("  Memory used: %d MB", memUsed)
			t.Logf("  ns/op: %d", result.NsPerOp())
			t.Logf("  allocs/op: %d", result.AllocsPerOp())
			t.Logf("  bytes/op: %d", result.AllocedBytesPerOp())

			// Performance assertions
			if opsPerSec < tt.minOpsPerSec {
				t.Errorf("Performance below threshold: %.2f ops/sec < %.2f ops/sec",
					opsPerSec, tt.minOpsPerSec)
			}

			if float64(memUsed) > tt.maxMemMB {
				t.Errorf("Memory usage above threshold: %d MB > %.2f MB",
					memUsed, tt.maxMemMB)
			}
		})
	}
}

func TestConcurrentProviderAccess(t *testing.T) {
	manager := llm.NewProviderManager(nil)

	// Add multiple providers
	for i := 0; i < 5; i++ {
		provider := llm.NewFreeProvider(fmt.Sprintf("Provider%d", i), "", "")
		manager.RegisterProvider(provider)
	}

	ctx := context.Background()
	prompt := "Test prompt"
	options := models.ProcessingOptions{Quality: "standard"}

	// Test concurrent access
	numGoroutines := 100
	numRequests := 1000
	start := time.Now()

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numRequests/numGoroutines; j++ {
				_, _ = manager.GenerateWithFallback(ctx, prompt, options)
				_ = manager.GetBestProvider(llm.ProviderPreferences{})
				_ = manager.GetProviderInfo()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	elapsed := time.Since(start)
	totalRequests := numRequests
	reqsPerSec := float64(totalRequests) / elapsed.Seconds()

	t.Logf("Concurrent test completed:")
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  Total requests: %d", totalRequests)
	t.Logf("  Elapsed time: %v", elapsed)
	t.Logf("  Requests/sec: %.2f", reqsPerSec)

	// Performance assertion
	if reqsPerSec < 1000 {
		t.Errorf("Concurrent performance below threshold: %.2f req/sec < 1000 req/sec", reqsPerSec)
	}
}

func TestMemoryLeakDetection(t *testing.T) {
	manager := llm.NewProviderManager(nil)

	// Add a provider
	provider := llm.NewFreeProvider("TestProvider", "", "")
	manager.RegisterProvider(provider)

	ctx := context.Background()
	prompt := "Test prompt"
	options := models.ProcessingOptions{Quality: "standard"}

	// Baseline memory
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Perform many operations
	iterations := 10000
	for i := 0; i < iterations; i++ {
		_, _ = manager.GenerateWithFallback(ctx, prompt, options)
	}

	// Check memory after operations
	runtime.GC()
	runtime.ReadMemStats(&m2)

	memIncrease := (m2.Alloc - m1.Alloc) / 1024 / 1024 // MB

	t.Logf("Memory leak detection:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Memory increase: %d MB", memIncrease)

	// Memory leak assertion
	if memIncrease > 50 { // Allow some memory increase but not excessive
		t.Errorf("Potential memory leak detected: %d MB increase after %d operations",
			memIncrease, iterations)
	}
}
