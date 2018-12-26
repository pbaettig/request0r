# Test
A test defines
- n URLSpecs
- The number of requests that should be performed
- The concurrency factor
- Rate Limits / Delay

For each test n goroutines are started (according to the concurrency factor)