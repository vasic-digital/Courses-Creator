package performance_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/course-creator/core-processor/jobs"
	"github.com/course-creator/core-processor/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(tb testing.TB) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		tb.Fatalf("Failed to create test database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.ProcessingJobDB{})
	if err != nil {
		tb.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func BenchmarkJobQueue_Enqueue(b *testing.B) {
	// Create mock DB
	db := setupTestDB(b)
	queue := jobs.NewJobQueue(db, 2)
	queue.Start()
	defer queue.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job := &jobs.Job{
			ID:       fmt.Sprintf("job-%d", i),
			Type:     jobs.JobTypeVideoProcessing,
			UserID:   "test-user",
			Status:   jobs.JobStatusPending,
			Priority: jobs.JobPriorityNormal,
			Payload:  map[string]interface{}{"test": "data"},
		}
		queue.Enqueue(context.Background(), job.Type, job.UserID, job.Payload, job.Priority)
	}
}

func BenchmarkJobQueue_GetJob(b *testing.B) {
	// Create mock DB
	db := setupTestDB(b)
	queue := jobs.NewJobQueue(db, 2)
	defer queue.Stop()

	// Pre-populate with jobs
	testJobs := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		jobID := fmt.Sprintf("job-%d", i)
		testJobs[i] = jobID
		queue.Enqueue(context.Background(), jobs.JobTypeVideoProcessing, "test-user", 
			map[string]interface{}{"test": "data"}, jobs.JobPriorityNormal)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jobID := testJobs[i%len(testJobs)]
		_, _ = queue.GetJob(context.Background(), jobID)
	}
}

func BenchmarkJobQueue_GetUserJobs(b *testing.B) {
	// Create mock DB
	db := setupTestDB(b)
	queue := jobs.NewJobQueue(db, 2)
	defer queue.Stop()

	// Pre-populate with jobs for multiple users
	for user := 0; user < 10; user++ {
		userID := fmt.Sprintf("user-%d", user)
		for i := 0; i < 100; i++ {
			queue.Enqueue(context.Background(), jobs.JobTypeVideoProcessing, userID,
				map[string]interface{}{"test": "data"}, jobs.JobPriorityNormal)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := fmt.Sprintf("user-%d", i%10)
		_, _ = queue.GetUserJobs(context.Background(), userID, 100, 0)
	}
}

func TestJobQueuePerformance(t *testing.T) {
	tests := []struct {
		name         string
		testFunc     func(b *testing.B)
		minOpsPerSec float64
		maxMemMB     float64
	}{
		{
			name:         "JobQueue_Enqueue",
			testFunc:     BenchmarkJobQueue_Enqueue,
			minOpsPerSec: 10000,
			maxMemMB:     50,
		},
		{
			name:         "JobQueue_GetJob",
			testFunc:     BenchmarkJobQueue_GetJob,
			minOpsPerSec: 5000,
			maxMemMB:     20,
		},
		{
			name:         "JobQueue_GetUserJobs",
			testFunc:     BenchmarkJobQueue_GetUserJobs,
			minOpsPerSec: 1000,
			maxMemMB:     30,
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

func TestJobQueueScalability(t *testing.T) {
	// Create mock DB
	db := setupTestDB(t)
	queue := jobs.NewJobQueue(db, 2)
	queue.Start()
	defer queue.Stop()

	// Test scalability with different job counts
	jobCounts := []int{100, 1000, 10000, 100000}

	for _, count := range jobCounts {
		t.Run(fmt.Sprintf("Jobs_%d", count), func(t *testing.T) {
			// Record initial memory
			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			// Enqueue jobs
			start := time.Now()
			for i := 0; i < count; i++ {
				queue.Enqueue(context.Background(), jobs.JobTypeVideoProcessing,
					"test-user", map[string]interface{}{"test": "data"},
					jobs.JobPriorityNormal)
			}
			enqueueTime := time.Since(start)

			// Get jobs
			start = time.Now()
			for i := 0; i < 100; i++ { // Sample 100 jobs
				jobID := fmt.Sprintf("%d", i%count)
				_, _ = queue.GetJob(context.Background(), jobID)
			}
			getTime := time.Since(start)

			// Check memory
			runtime.GC()
			runtime.ReadMemStats(&m2)
			memUsed := (m2.Alloc - m1.Alloc) / 1024 / 1024 // MB

			t.Logf("Scalability test with %d jobs:", count)
			t.Logf("  Enqueue time: %v (%.2f ms/job)", enqueueTime, float64(enqueueTime.Nanoseconds())/float64(count)/1000000)
			t.Logf("  Get time (100 jobs): %v (%.2f ms/job)", getTime, float64(getTime.Nanoseconds())/100/1000000)
			t.Logf("  Memory used: %d MB", memUsed)
			t.Logf("  Memory per job: %.2f KB", float64(memUsed)*1024/float64(count))

			// Performance assertions
			enqueueRate := float64(count) / enqueueTime.Seconds()
			if enqueueRate < 1000 {
				t.Errorf("Enqueue rate too low: %.2f jobs/sec < 1000 jobs/sec", enqueueRate)
			}

			memPerJob := float64(memUsed)*1024 / float64(count)
			if memPerJob > 10 { // 10 KB per job
				t.Errorf("Memory per job too high: %.2f KB/job > 10 KB/job", memPerJob)
			}
		})
	}
}

func TestJobQueueConcurrency(t *testing.T) {
	// Create mock DB
	db := setupTestDB(t)
	queue := jobs.NewJobQueue(db, 4)
	queue.Start()
	defer queue.Stop()

	// Test concurrent enqueue and get operations
	numGoroutines := 50
	numOperations := 1000

	start := time.Now()
	done := make(chan bool, numGoroutines*2)

	// Concurrent enqueuers
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			for j := 0; j < numOperations/numGoroutines; j++ {
				queue.Enqueue(context.Background(), jobs.JobTypeVideoProcessing,
					fmt.Sprintf("user-%d", workerID),
					map[string]interface{}{"test": "data"},
					jobs.JobPriorityNormal)
			}
			done <- true
		}(i)
	}

	// Concurrent getters
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			for j := 0; j < numOperations/numGoroutines; j++ {
				jobID := fmt.Sprintf("%d", j)
				_, _ = queue.GetJob(context.Background(), jobID)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines*2; i++ {
		<-done
	}

	elapsed := time.Since(start)
	totalOps := numOperations * 2
	opsPerSec := float64(totalOps) / elapsed.Seconds()

	t.Logf("Concurrency test completed:")
	t.Logf("  Goroutines: %d", numGoroutines*2)
	t.Logf("  Total operations: %d", totalOps)
	t.Logf("  Elapsed time: %v", elapsed)
	t.Logf("  Operations/sec: %.2f", opsPerSec)

	// Performance assertion
	if opsPerSec < 5000 {
		t.Errorf("Concurrent performance below threshold: %.2f ops/sec < 5000 ops/sec", opsPerSec)
	}
}

func TestJobQueueMemoryLeak(t *testing.T) {
	// Create mock DB
	db := setupTestDB(t)
	queue := jobs.NewJobQueue(db, 2)
	queue.Start()
	defer queue.Stop()

	// Baseline memory
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Perform many enqueue and get operations
	iterations := 5000
	for i := 0; i < iterations; i++ {
		job, err := queue.Enqueue(context.Background(), jobs.JobTypeVideoProcessing,
			"test-user", map[string]interface{}{"test": "data"},
			jobs.JobPriorityNormal)
		if err == nil {
			queue.GetJob(context.Background(), job.ID)
		}
	}

	// Check memory after operations
	runtime.GC()
	runtime.ReadMemStats(&m2)

	memIncrease := (m2.Alloc - m1.Alloc) / 1024 / 1024 // MB

	t.Logf("Memory leak detection:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Memory increase: %d MB", memIncrease)

	// Memory leak assertion
	if memIncrease > 100 { // Allow some memory increase but not excessive
		t.Errorf("Potential memory leak detected: %d MB increase after %d operations", 
			memIncrease, iterations)
	}
}