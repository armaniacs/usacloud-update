package bdd

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// ContinuousTestRunner manages continuous test execution
type ContinuousTestRunner struct {
	mu            sync.RWMutex
	scenarios     []string
	schedule      *Scheduler
	results       *TestResultStore
	notifications *NotificationService
	running       bool
	stopChan      chan struct{}
}

// TestRun represents a complete test execution
type TestRun struct {
	ID        string
	Timestamp time.Time
	Duration  time.Duration
	Results   map[string]TestResult
	Summary   TestSummary
}

// TestResult represents the result of a single test scenario
type TestResult struct {
	ScenarioName string
	Passed       bool
	Duration     time.Duration
	Error        string
	Details      map[string]interface{}
}

// TestSummary provides aggregate test results
type TestSummary struct {
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	PassRate     float64
}

// TestResultStore manages persistent storage of test results
type TestResultStore struct {
	mu      sync.RWMutex
	runs    map[string]TestRun
	maxRuns int
}

// NotificationService handles test result notifications
type NotificationService struct {
	channels map[string]NotificationChannel
	enabled  bool
}

// NotificationChannel defines a notification delivery method
type NotificationChannel interface {
	SendAlert(message string) error
	SendInfo(message string) error
	SendSummary(summary TestSummary) error
}

// Scheduler manages test scheduling
type Scheduler struct {
	jobs    map[string]*ScheduledJob
	ticker  *time.Ticker
	running bool
	stopCh  chan struct{}
}

// ScheduledJob represents a scheduled test job
type ScheduledJob struct {
	Name     string
	Schedule string // cron-like schedule
	Job      func() error
	LastRun  time.Time
	NextRun  time.Time
}

// NewContinuousTestRunner creates a new continuous test runner
func NewContinuousTestRunner() *ContinuousTestRunner {
	return &ContinuousTestRunner{
		scenarios:     make([]string, 0),
		schedule:      NewScheduler(),
		results:       NewTestResultStore(),
		notifications: NewNotificationService(),
		stopChan:      make(chan struct{}),
	}
}

// AddScenario adds a test scenario to the runner
func (ctr *ContinuousTestRunner) AddScenario(scenarioName string) {
	ctr.mu.Lock()
	defer ctr.mu.Unlock()

	ctr.scenarios = append(ctr.scenarios, scenarioName)
}

// ScheduleTests sets up the test execution schedule
func (ctr *ContinuousTestRunner) ScheduleTests() error {
	// Full test suite - daily at 3 AM
	if err := ctr.schedule.AddJob("full-suite", "0 3 * * *", ctr.runFullTestSuite); err != nil {
		return fmt.Errorf("failed to schedule full test suite: %w", err)
	}

	// Smoke tests - hourly
	if err := ctr.schedule.AddJob("smoke-tests", "0 * * * *", ctr.runSmokeTests); err != nil {
		return fmt.Errorf("failed to schedule smoke tests: %w", err)
	}

	// Regression tests - every 6 hours
	if err := ctr.schedule.AddJob("regression-tests", "0 */6 * * *", ctr.runRegressionTests); err != nil {
		return fmt.Errorf("failed to schedule regression tests: %w", err)
	}

	return ctr.schedule.Start()
}

// Start begins continuous test execution
func (ctr *ContinuousTestRunner) Start() error {
	ctr.mu.Lock()
	defer ctr.mu.Unlock()

	if ctr.running {
		return fmt.Errorf("continuous test runner is already running")
	}

	ctr.running = true

	// Start the scheduler
	if err := ctr.ScheduleTests(); err != nil {
		ctr.running = false
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	log.Println("Continuous test runner started")
	return nil
}

// Stop stops continuous test execution
func (ctr *ContinuousTestRunner) Stop() {
	ctr.mu.Lock()
	defer ctr.mu.Unlock()

	if !ctr.running {
		return
	}

	ctr.running = false
	close(ctr.stopChan)
	ctr.schedule.Stop()

	log.Println("Continuous test runner stopped")
}

// runFullTestSuite executes all available test scenarios
func (ctr *ContinuousTestRunner) runFullTestSuite() error {
	log.Println("Starting full BDD test suite...")

	startTime := time.Now()
	results := make(map[string]TestResult)

	ctr.mu.RLock()
	scenarios := make([]string, len(ctr.scenarios))
	copy(scenarios, ctr.scenarios)
	ctr.mu.RUnlock()

	for _, scenario := range scenarios {
		result := ctr.runScenario(scenario)
		results[scenario] = result

		// Send immediate alert for failures
		if !result.Passed {
			message := fmt.Sprintf("BDD Test Failed: %s - %s", scenario, result.Error)
			ctr.notifications.SendAlert(message)
		}
	}

	duration := time.Since(startTime)

	// Create test run record
	testRun := TestRun{
		ID:        ctr.generateRunID(),
		Timestamp: startTime,
		Duration:  duration,
		Results:   results,
		Summary:   ctr.calculateSummary(results),
	}

	// Store results
	ctr.results.Save(testRun)

	// Send summary notification
	ctr.sendSummaryNotification(testRun)

	log.Printf("Full test suite completed in %v", duration)
	return nil
}

// runSmokeTests executes critical smoke tests
func (ctr *ContinuousTestRunner) runSmokeTests() error {
	log.Println("Starting smoke tests...")

	smokeScenarios := []string{
		"network_connection_test",
		"basic_authentication_test",
		"configuration_validation_test",
	}

	results := make(map[string]TestResult)

	for _, scenario := range smokeScenarios {
		result := ctr.runScenario(scenario)
		results[scenario] = result

		if !result.Passed {
			message := fmt.Sprintf("Smoke Test Failed: %s - %s", scenario, result.Error)
			ctr.notifications.SendAlert(message)
		}
	}

	log.Printf("Smoke tests completed - %d passed, %d failed",
		ctr.countPassed(results), ctr.countFailed(results))

	return nil
}

// runRegressionTests executes regression test scenarios
func (ctr *ContinuousTestRunner) runRegressionTests() error {
	log.Println("Starting regression tests...")

	regressionScenarios := []string{
		"known_bug_regression_test",
		"performance_regression_test",
		"security_regression_test",
	}

	results := make(map[string]TestResult)

	for _, scenario := range regressionScenarios {
		result := ctr.runScenario(scenario)
		results[scenario] = result

		if !result.Passed {
			message := fmt.Sprintf("Regression Test Failed: %s - %s", scenario, result.Error)
			ctr.notifications.SendAlert(message)
		}
	}

	log.Printf("Regression tests completed - %d passed, %d failed",
		ctr.countPassed(results), ctr.countFailed(results))

	return nil
}

// runScenario executes a single test scenario
func (ctr *ContinuousTestRunner) runScenario(scenarioName string) TestResult {
	startTime := time.Now()

	// Simulate scenario execution (in real implementation, this would run actual BDD scenarios)
	var err error
	var passed bool

	switch scenarioName {
	case "network_connection_test":
		err = ctr.simulateNetworkTest()
		passed = err == nil
	case "basic_authentication_test":
		err = ctr.simulateAuthTest()
		passed = err == nil
	case "configuration_validation_test":
		err = ctr.simulateConfigTest()
		passed = err == nil
	default:
		err = ctr.simulateGenericTest(scenarioName)
		passed = err == nil
	}

	duration := time.Since(startTime)

	result := TestResult{
		ScenarioName: scenarioName,
		Passed:       passed,
		Duration:     duration,
		Error:        "",
		Details:      make(map[string]interface{}),
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

// sendSummaryNotification sends a summary of test results
func (ctr *ContinuousTestRunner) sendSummaryNotification(run TestRun) {
	summary := run.Summary
	message := fmt.Sprintf(
		"BDD Test Run Complete\nPassed: %d, Failed: %d, Duration: %v, Pass Rate: %.1f%%",
		summary.PassedTests, summary.FailedTests, run.Duration, summary.PassRate*100)

	if summary.FailedTests > 0 {
		ctr.notifications.SendAlert(message)
	} else {
		ctr.notifications.SendInfo(message)
	}

	// Send detailed summary
	ctr.notifications.SendSummary(summary)
}

// calculateSummary calculates test summary statistics
func (ctr *ContinuousTestRunner) calculateSummary(results map[string]TestResult) TestSummary {
	total := len(results)
	passed := ctr.countPassed(results)
	failed := ctr.countFailed(results)

	passRate := 0.0
	if total > 0 {
		passRate = float64(passed) / float64(total)
	}

	return TestSummary{
		TotalTests:   total,
		PassedTests:  passed,
		FailedTests:  failed,
		SkippedTests: 0,
		PassRate:     passRate,
	}
}

// countPassed counts the number of passed tests
func (ctr *ContinuousTestRunner) countPassed(results map[string]TestResult) int {
	count := 0
	for _, result := range results {
		if result.Passed {
			count++
		}
	}
	return count
}

// countFailed counts the number of failed tests
func (ctr *ContinuousTestRunner) countFailed(results map[string]TestResult) int {
	count := 0
	for _, result := range results {
		if !result.Passed {
			count++
		}
	}
	return count
}

// generateRunID generates a unique run identifier
func (ctr *ContinuousTestRunner) generateRunID() string {
	return fmt.Sprintf("run_%d", time.Now().Unix())
}

// Simulation methods for testing

func (ctr *ContinuousTestRunner) simulateNetworkTest() error {
	// Simulate network connectivity test
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (ctr *ContinuousTestRunner) simulateAuthTest() error {
	// Simulate authentication test
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (ctr *ContinuousTestRunner) simulateConfigTest() error {
	// Simulate configuration validation test
	time.Sleep(25 * time.Millisecond)
	return nil
}

func (ctr *ContinuousTestRunner) simulateGenericTest(scenarioName string) error {
	// Simulate generic test execution
	time.Sleep(75 * time.Millisecond)

	// Randomly fail some tests for demonstration
	if time.Now().UnixNano()%10 == 0 {
		return fmt.Errorf("simulated failure for scenario: %s", scenarioName)
	}

	return nil
}

// TestResultStore implementation

// NewTestResultStore creates a new test result store
func NewTestResultStore() *TestResultStore {
	return &TestResultStore{
		runs:    make(map[string]TestRun),
		maxRuns: 100, // Keep last 100 runs
	}
}

// Save stores a test run result
func (trs *TestResultStore) Save(run TestRun) {
	trs.mu.Lock()
	defer trs.mu.Unlock()

	trs.runs[run.ID] = run

	// Clean up old runs if we exceed maxRuns
	if len(trs.runs) > trs.maxRuns {
		trs.cleanup()
	}
}

// Get retrieves a test run by ID
func (trs *TestResultStore) Get(id string) (TestRun, bool) {
	trs.mu.RLock()
	defer trs.mu.RUnlock()

	run, exists := trs.runs[id]
	return run, exists
}

// GetLatest returns the most recent test runs
func (trs *TestResultStore) GetLatest(count int) []TestRun {
	trs.mu.RLock()
	defer trs.mu.RUnlock()

	runs := make([]TestRun, 0, len(trs.runs))
	for _, run := range trs.runs {
		runs = append(runs, run)
	}

	// Sort by timestamp (most recent first)
	for i := 0; i < len(runs)-1; i++ {
		for j := i + 1; j < len(runs); j++ {
			if runs[i].Timestamp.Before(runs[j].Timestamp) {
				runs[i], runs[j] = runs[j], runs[i]
			}
		}
	}

	if count > len(runs) {
		count = len(runs)
	}

	return runs[:count]
}

// cleanup removes old test runs
func (trs *TestResultStore) cleanup() {
	// Keep only the most recent runs
	latest := trs.GetLatest(trs.maxRuns / 2)

	// Clear and rebuild with latest runs
	trs.runs = make(map[string]TestRun)
	for _, run := range latest {
		trs.runs[run.ID] = run
	}
}

// NotificationService implementation

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{
		channels: make(map[string]NotificationChannel),
		enabled:  true,
	}
}

// AddChannel adds a notification channel
func (ns *NotificationService) AddChannel(name string, channel NotificationChannel) {
	ns.channels[name] = channel
}

// SendAlert sends an alert message to all channels
func (ns *NotificationService) SendAlert(message string) {
	if !ns.enabled {
		return
	}

	for name, channel := range ns.channels {
		if err := channel.SendAlert(message); err != nil {
			log.Printf("Failed to send alert via %s: %v", name, err)
		}
	}

	// Always log alerts
	log.Printf("ALERT: %s", message)
}

// SendInfo sends an info message to all channels
func (ns *NotificationService) SendInfo(message string) {
	if !ns.enabled {
		return
	}

	for name, channel := range ns.channels {
		if err := channel.SendInfo(message); err != nil {
			log.Printf("Failed to send info via %s: %v", name, err)
		}
	}

	log.Printf("INFO: %s", message)
}

// SendSummary sends a test summary to all channels
func (ns *NotificationService) SendSummary(summary TestSummary) {
	if !ns.enabled {
		return
	}

	for name, channel := range ns.channels {
		if err := channel.SendSummary(summary); err != nil {
			log.Printf("Failed to send summary via %s: %v", name, err)
		}
	}
}

// Scheduler implementation

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs:   make(map[string]*ScheduledJob),
		stopCh: make(chan struct{}),
	}
}

// AddJob adds a scheduled job
func (s *Scheduler) AddJob(name, schedule string, job func() error) error {
	s.jobs[name] = &ScheduledJob{
		Name:     name,
		Schedule: schedule,
		Job:      job,
		NextRun:  s.calculateNextRun(schedule),
	}
	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	s.running = true
	s.ticker = time.NewTicker(1 * time.Minute) // Check every minute

	go s.run()
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	if !s.running {
		return
	}

	s.running = false
	s.ticker.Stop()
	close(s.stopCh)
}

// run executes the scheduler loop
func (s *Scheduler) run() {
	for {
		select {
		case <-s.ticker.C:
			s.checkAndRunJobs()
		case <-s.stopCh:
			return
		}
	}
}

// checkAndRunJobs checks if any jobs should run and executes them
func (s *Scheduler) checkAndRunJobs() {
	now := time.Now()

	for _, job := range s.jobs {
		if now.After(job.NextRun) || now.Equal(job.NextRun) {
			go func(j *ScheduledJob) {
				log.Printf("Running scheduled job: %s", j.Name)
				if err := j.Job(); err != nil {
					log.Printf("Scheduled job %s failed: %v", j.Name, err)
				}
				j.LastRun = time.Now()
				j.NextRun = s.calculateNextRun(j.Schedule)
			}(job)
		}
	}
}

// calculateNextRun calculates the next run time for a schedule
func (s *Scheduler) calculateNextRun(schedule string) time.Time {
	// Simplified schedule parsing - in a real implementation, use a proper cron parser
	now := time.Now()

	switch schedule {
	case "0 3 * * *": // Daily at 3 AM
		next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		return next
	case "0 * * * *": // Hourly
		next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(1 * time.Hour)
		}
		return next
	case "0 */6 * * *": // Every 6 hours
		next := time.Date(now.Year(), now.Month(), now.Day(), (now.Hour()/6)*6, 0, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(6 * time.Hour)
		}
		return next
	default:
		// Default to 1 hour from now
		return now.Add(1 * time.Hour)
	}
}
