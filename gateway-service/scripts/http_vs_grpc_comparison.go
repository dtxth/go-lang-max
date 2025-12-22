package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"gateway-service/internal/config"
	grpcClient "gateway-service/internal/infrastructure/grpc"

	authpb "auth-service/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ComparisonResult holds performance comparison results
type ComparisonResult struct {
	Protocol        string
	TotalRequests   int
	SuccessfulReqs  int
	FailedReqs      int
	TotalDuration   time.Duration
	AverageDuration time.Duration
	MinDuration     time.Duration
	MaxDuration     time.Duration
	RequestsPerSec  float64
	ErrorRate       float64
}

// PerformanceComparator compares HTTP vs gRPC performance
type PerformanceComparator struct {
	httpBaseURL  string
	grpcAddress  string
	config       *config.Config
}

// NewPerformanceComparator creates a new performance comparator
func NewPerformanceComparator() *PerformanceComparator {
	return &PerformanceComparator{
		httpBaseURL: "http://localhost:8080", // Gateway HTTP endpoint
		grpcAddress: "localhost:9090",        // Auth service gRPC endpoint
		config:      config.Load(),
	}
}

// CompareProtocols compares HTTP and gRPC performance
func (pc *PerformanceComparator) CompareProtocols(numRequests int, concurrency int) {
	fmt.Printf("=== Performance Comparison: HTTP vs gRPC ===\n")
	fmt.Printf("Requests: %d, Concurrency: %d\n\n", numRequests, concurrency)

	// Test HTTP performance
	fmt.Println("Testing HTTP performance...")
	httpResult := pc.testHTTPPerformance(numRequests, concurrency)

	// Test gRPC performance
	fmt.Println("Testing gRPC performance...")
	grpcResult := pc.testGRPCPerformance(numRequests, concurrency)

	// Display results
	pc.displayResults(httpResult, grpcResult)
}

// testHTTPPerformance tests HTTP endpoint performance
func (pc *PerformanceComparator) testHTTPPerformance(numRequests int, concurrency int) *ComparisonResult {
	result := &ComparisonResult{
		Protocol:      "HTTP",
		TotalRequests: numRequests,
		MinDuration:   time.Duration(1<<63 - 1),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	durations := make([]time.Duration, 0, numRequests)

	// Create HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Prepare request payload
	loginReq := map[string]string{
		"phone":    "+1234567890",
		"password": "testpassword",
	}
	jsonData, _ := json.Marshal(loginReq)

	start := time.Now()

	// Create semaphore for concurrency control
	sem := make(chan struct{}, concurrency)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			sem <- struct{}{} // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			reqStart := time.Now()
			
			// Make HTTP request
			resp, err := client.Post(
				pc.httpBaseURL+"/auth/login",
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			
			duration := time.Since(reqStart)
			
			mu.Lock()
			durations = append(durations, duration)
			
			if err != nil || (resp != nil && resp.StatusCode >= 400) {
				result.FailedReqs++
			} else {
				result.SuccessfulReqs++
			}
			
			if resp != nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	result.TotalDuration = time.Since(start)

	// Calculate statistics
	pc.calculateStatistics(result, durations)
	
	return result
}

// testGRPCPerformance tests gRPC endpoint performance
func (pc *PerformanceComparator) testGRPCPerformance(numRequests int, concurrency int) *ComparisonResult {
	result := &ComparisonResult{
		Protocol:      "gRPC",
		TotalRequests: numRequests,
		MinDuration:   time.Duration(1<<63 - 1),
	}

	// Create gRPC connection
	conn, err := grpc.Dial(pc.grpcAddress, 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("Failed to connect to gRPC server: %v", err)
		// Return empty result if connection fails
		return result
	}
	defer conn.Close()

	client := authpb.NewAuthServiceClient(conn)

	var wg sync.WaitGroup
	var mu sync.Mutex
	durations := make([]time.Duration, 0, numRequests)

	start := time.Now()

	// Create semaphore for concurrency control
	sem := make(chan struct{}, concurrency)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			sem <- struct{}{} // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			reqStart := time.Now()
			
			// Make gRPC request
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			
			_, err := client.Login(ctx, &authpb.LoginRequest{
				Phone:    "+1234567890",
				Password: "testpassword",
			})
			
			duration := time.Since(reqStart)
			
			mu.Lock()
			durations = append(durations, duration)
			
			if err != nil {
				result.FailedReqs++
			} else {
				result.SuccessfulReqs++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	result.TotalDuration = time.Since(start)

	// Calculate statistics
	pc.calculateStatistics(result, durations)
	
	return result
}

// calculateStatistics calculates performance statistics
func (pc *PerformanceComparator) calculateStatistics(result *ComparisonResult, durations []time.Duration) {
	if len(durations) == 0 {
		return
	}

	var totalDuration time.Duration
	for _, d := range durations {
		totalDuration += d
		if d < result.MinDuration {
			result.MinDuration = d
		}
		if d > result.MaxDuration {
			result.MaxDuration = d
		}
	}

	result.AverageDuration = totalDuration / time.Duration(len(durations))
	result.RequestsPerSec = float64(result.TotalRequests) / result.TotalDuration.Seconds()
	result.ErrorRate = float64(result.FailedReqs) / float64(result.TotalRequests) * 100
}

// displayResults displays comparison results
func (pc *PerformanceComparator) displayResults(httpResult, grpcResult *ComparisonResult) {
	fmt.Printf("\n=== Performance Comparison Results ===\n\n")

	// Display HTTP results
	fmt.Printf("HTTP Results:\n")
	pc.printResult(httpResult)

	// Display gRPC results
	fmt.Printf("\ngRPC Results:\n")
	pc.printResult(grpcResult)

	// Display comparison
	fmt.Printf("\n=== Comparison Summary ===\n")
	
	if httpResult.RequestsPerSec > 0 && grpcResult.RequestsPerSec > 0 {
		throughputImprovement := (grpcResult.RequestsPerSec - httpResult.RequestsPerSec) / httpResult.RequestsPerSec * 100
		latencyImprovement := (float64(httpResult.AverageDuration.Nanoseconds()) - float64(grpcResult.AverageDuration.Nanoseconds())) / float64(httpResult.AverageDuration.Nanoseconds()) * 100

		fmt.Printf("Throughput: gRPC is %.1f%% %s than HTTP\n", 
			abs(throughputImprovement), 
			comparison(throughputImprovement > 0))
		
		fmt.Printf("Latency: gRPC is %.1f%% %s than HTTP\n", 
			abs(latencyImprovement), 
			comparison(latencyImprovement > 0))

		// Recommendations
		fmt.Printf("\n=== Recommendations ===\n")
		if grpcResult.RequestsPerSec > httpResult.RequestsPerSec {
			fmt.Println("✅ gRPC shows better performance for inter-service communication")
		} else {
			fmt.Println("⚠️  HTTP shows better performance - investigate gRPC configuration")
		}

		if grpcResult.ErrorRate < httpResult.ErrorRate {
			fmt.Println("✅ gRPC shows better reliability")
		} else if grpcResult.ErrorRate > httpResult.ErrorRate {
			fmt.Println("⚠️  gRPC shows higher error rate - check service health")
		}
	} else {
		fmt.Println("⚠️  Unable to complete comparison due to connection issues")
	}
}

// printResult prints individual result
func (pc *PerformanceComparator) printResult(result *ComparisonResult) {
	fmt.Printf("  Protocol: %s\n", result.Protocol)
	fmt.Printf("  Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("  Successful: %d\n", result.SuccessfulReqs)
	fmt.Printf("  Failed: %d\n", result.FailedReqs)
	fmt.Printf("  Error Rate: %.2f%%\n", result.ErrorRate)
	fmt.Printf("  Total Duration: %v\n", result.TotalDuration)
	fmt.Printf("  Average Latency: %v\n", result.AverageDuration)
	fmt.Printf("  Min Latency: %v\n", result.MinDuration)
	fmt.Printf("  Max Latency: %v\n", result.MaxDuration)
	fmt.Printf("  Throughput: %.2f req/sec\n", result.RequestsPerSec)
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func comparison(better bool) string {
	if better {
		return "better"
	}
	return "worse"
}

// runLoadTest runs a comprehensive load test
func (pc *PerformanceComparator) runLoadTest() {
	fmt.Println("=== Comprehensive Load Test ===\n")

	testCases := []struct {
		name        string
		requests    int
		concurrency int
	}{
		{"Light Load", 100, 5},
		{"Medium Load", 500, 20},
		{"Heavy Load", 1000, 50},
		{"Stress Test", 2000, 100},
	}

	for _, tc := range testCases {
		fmt.Printf("--- %s ---\n", tc.name)
		pc.CompareProtocols(tc.requests, tc.concurrency)
		fmt.Println()
		time.Sleep(2 * time.Second) // Brief pause between tests
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run http_vs_grpc_comparison.go <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  compare <requests> <concurrency>  - Compare HTTP vs gRPC performance")
		fmt.Println("  load-test                         - Run comprehensive load test")
		fmt.Println("Examples:")
		fmt.Println("  go run http_vs_grpc_comparison.go compare 1000 50")
		fmt.Println("  go run http_vs_grpc_comparison.go load-test")
		os.Exit(1)
	}

	comparator := NewPerformanceComparator()
	command := os.Args[1]

	switch command {
	case "compare":
		if len(os.Args) < 4 {
			fmt.Println("Usage: compare <requests> <concurrency>")
			os.Exit(1)
		}
		
		var requests, concurrency int
		fmt.Sscanf(os.Args[2], "%d", &requests)
		fmt.Sscanf(os.Args[3], "%d", &concurrency)
		
		if requests <= 0 || concurrency <= 0 {
			fmt.Println("Requests and concurrency must be positive integers")
			os.Exit(1)
		}
		
		comparator.CompareProtocols(requests, concurrency)
		
	case "load-test":
		comparator.runLoadTest()
		
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}