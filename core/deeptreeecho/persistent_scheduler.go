package deeptreeecho

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// =============================================================================
// PERSISTENT ECHOBEATS SCHEDULER
// =============================================================================
//
// Implements persistent goal-directed scheduling for echobeats.
// Jobs survive restarts and are automatically recovered.
//
// Inspired by AGScheduler patterns:
// - Multiple job store backends (file, memory)
// - Cron, interval, and one-off scheduling
// - Job event listeners
// - Recovery on restart
//
// =============================================================================

// JobType defines the type of scheduled job
type JobType string

const (
	JobTypeOneOff   JobType = "one_off"
	JobTypeInterval JobType = "interval"
	JobTypeCron     JobType = "cron"
)

// JobStatus defines the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusPaused    JobStatus = "paused"
)

// ScheduledJob represents a persistent scheduled job
type ScheduledJob struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        JobType           `json:"type"`
	Status      JobStatus         `json:"status"`
	Priority    int               `json:"priority"`
	
	// Scheduling parameters
	Interval    time.Duration     `json:"interval,omitempty"`
	CronExpr    string            `json:"cron_expr,omitempty"`
	NextRunAt   time.Time         `json:"next_run_at"`
	LastRunAt   time.Time         `json:"last_run_at,omitempty"`
	
	// Job payload
	Payload     map[string]interface{} `json:"payload"`
	
	// Execution context
	GoalID      string            `json:"goal_id,omitempty"`
	Context     string            `json:"context"`
	
	// Metadata
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	RunCount    int               `json:"run_count"`
	MaxRuns     int               `json:"max_runs,omitempty"` // 0 = unlimited
	
	// Error tracking
	LastError   string            `json:"last_error,omitempty"`
	ErrorCount  int               `json:"error_count"`
}

// JobStore interface for job persistence
type JobStore interface {
	Save(job *ScheduledJob) error
	Load(id string) (*ScheduledJob, error)
	LoadAll() ([]*ScheduledJob, error)
	Delete(id string) error
	Update(job *ScheduledJob) error
}

// FileJobStore implements JobStore using file system
type FileJobStore struct {
	mu       sync.RWMutex
	basePath string
}

// NewFileJobStore creates a new file-based job store
func NewFileJobStore(basePath string) (*FileJobStore, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create job store directory: %w", err)
	}
	return &FileJobStore{basePath: basePath}, nil
}

// Save persists a job to file
func (s *FileJobStore) Save(job *ScheduledJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	job.UpdatedAt = time.Now()
	
	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	
	path := filepath.Join(s.basePath, job.ID+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write job file: %w", err)
	}
	
	return nil
}

// Load retrieves a job from file
func (s *FileJobStore) Load(id string) (*ScheduledJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	path := filepath.Join(s.basePath, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read job file: %w", err)
	}
	
	var job ScheduledJob
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}
	
	return &job, nil
}

// LoadAll retrieves all jobs from the store
func (s *FileJobStore) LoadAll() ([]*ScheduledJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read job store directory: %w", err)
	}
	
	var jobs []*ScheduledJob
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		
		path := filepath.Join(s.basePath, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Skip unreadable files
		}
		
		var job ScheduledJob
		if err := json.Unmarshal(data, &job); err != nil {
			continue // Skip invalid files
		}
		
		jobs = append(jobs, &job)
	}
	
	return jobs, nil
}

// Delete removes a job from the store
func (s *FileJobStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	path := filepath.Join(s.basePath, id+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete job file: %w", err)
	}
	
	return nil
}

// Update updates a job in the store
func (s *FileJobStore) Update(job *ScheduledJob) error {
	return s.Save(job)
}

// JobEventType defines types of job events
type JobEventType string

const (
	JobEventScheduled JobEventType = "scheduled"
	JobEventStarted   JobEventType = "started"
	JobEventCompleted JobEventType = "completed"
	JobEventFailed    JobEventType = "failed"
	JobEventPaused    JobEventType = "paused"
	JobEventResumed   JobEventType = "resumed"
	JobEventDeleted   JobEventType = "deleted"
)

// JobEvent represents an event related to a job
type JobEvent struct {
	Type      JobEventType
	Job       *ScheduledJob
	Timestamp time.Time
	Error     error
}

// JobEventListener is a callback for job events
type JobEventListener func(event *JobEvent)

// JobExecutor is a function that executes a job
type JobExecutor func(ctx context.Context, job *ScheduledJob) error

// PersistentScheduler manages persistent job scheduling
type PersistentScheduler struct {
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	
	store        JobStore
	jobs         map[string]*ScheduledJob
	executors    map[string]JobExecutor
	listeners    []JobEventListener
	
	running      bool
	tickInterval time.Duration
}

// NewPersistentScheduler creates a new persistent scheduler
func NewPersistentScheduler(ctx context.Context, store JobStore) *PersistentScheduler {
	ctx, cancel := context.WithCancel(ctx)
	
	return &PersistentScheduler{
		ctx:          ctx,
		cancel:       cancel,
		store:        store,
		jobs:         make(map[string]*ScheduledJob),
		executors:    make(map[string]JobExecutor),
		listeners:    make([]JobEventListener, 0),
		tickInterval: time.Second,
	}
}

// RegisterExecutor registers an executor for a job context
func (ps *PersistentScheduler) RegisterExecutor(context string, executor JobExecutor) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.executors[context] = executor
}

// AddListener adds a job event listener
func (ps *PersistentScheduler) AddListener(listener JobEventListener) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.listeners = append(ps.listeners, listener)
}

// emitEvent sends an event to all listeners
func (ps *PersistentScheduler) emitEvent(eventType JobEventType, job *ScheduledJob, err error) {
	event := &JobEvent{
		Type:      eventType,
		Job:       job,
		Timestamp: time.Now(),
		Error:     err,
	}
	
	for _, listener := range ps.listeners {
		go listener(event)
	}
}

// Start begins the scheduler
func (ps *PersistentScheduler) Start() error {
	ps.mu.Lock()
	if ps.running {
		ps.mu.Unlock()
		return fmt.Errorf("scheduler already running")
	}
	ps.running = true
	ps.mu.Unlock()
	
	// Recover jobs from store
	if err := ps.recoverJobs(); err != nil {
		return fmt.Errorf("failed to recover jobs: %w", err)
	}
	
	// Start scheduler loop
	go ps.schedulerLoop()
	
	fmt.Println("⏰ Persistent Echobeats Scheduler: Started")
	return nil
}

// Stop stops the scheduler
func (ps *PersistentScheduler) Stop() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	if !ps.running {
		return fmt.Errorf("scheduler not running")
	}
	
	ps.cancel()
	ps.running = false
	
	fmt.Println("⏰ Persistent Echobeats Scheduler: Stopped")
	return nil
}

// recoverJobs loads and schedules all persisted jobs
func (ps *PersistentScheduler) recoverJobs() error {
	jobs, err := ps.store.LoadAll()
	if err != nil {
		return err
	}
	
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	recovered := 0
	for _, job := range jobs {
		// Skip completed or failed jobs
		if job.Status == JobStatusCompleted || job.Status == JobStatusFailed {
			continue
		}
		
		// Reset running jobs to pending
		if job.Status == JobStatusRunning {
			job.Status = JobStatusPending
		}
		
		// Reschedule if next run is in the past
		if job.NextRunAt.Before(time.Now()) {
			switch job.Type {
			case JobTypeInterval:
				job.NextRunAt = time.Now().Add(job.Interval)
			case JobTypeCron:
				// TODO: Parse cron expression and calculate next run
				job.NextRunAt = time.Now().Add(time.Minute)
			default:
				job.NextRunAt = time.Now()
			}
		}
		
		ps.jobs[job.ID] = job
		recovered++
	}
	
	fmt.Printf("⏰ Recovered %d jobs from persistent store\n", recovered)
	return nil
}

// schedulerLoop is the main scheduling loop
func (ps *PersistentScheduler) schedulerLoop() {
	ticker := time.NewTicker(ps.tickInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ps.ctx.Done():
			return
		case <-ticker.C:
			ps.tick()
		}
	}
}

// tick processes one scheduler tick
func (ps *PersistentScheduler) tick() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	now := time.Now()
	
	for _, job := range ps.jobs {
		// Skip non-pending jobs
		if job.Status != JobStatusPending {
			continue
		}
		
		// Check if job is due
		if job.NextRunAt.After(now) {
			continue
		}
		
		// Execute job
		go ps.executeJob(job)
	}
}

// executeJob runs a job
func (ps *PersistentScheduler) executeJob(job *ScheduledJob) {
	ps.mu.Lock()
	job.Status = JobStatusRunning
	job.LastRunAt = time.Now()
	ps.mu.Unlock()
	
	ps.emitEvent(JobEventStarted, job, nil)
	
	// Find executor
	executor, exists := ps.executors[job.Context]
	if !exists {
		executor = ps.defaultExecutor
	}
	
	// Execute with timeout
	ctx, cancel := context.WithTimeout(ps.ctx, 5*time.Minute)
	defer cancel()
	
	err := executor(ctx, job)
	
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	job.RunCount++
	
	if err != nil {
		job.LastError = err.Error()
		job.ErrorCount++
		
		// Mark as failed if too many errors
		if job.ErrorCount >= 3 {
			job.Status = JobStatusFailed
			ps.emitEvent(JobEventFailed, job, err)
		} else {
			// Retry later
			job.Status = JobStatusPending
			job.NextRunAt = time.Now().Add(time.Minute * time.Duration(job.ErrorCount))
		}
	} else {
		job.LastError = ""
		
		// Check if job is complete
		if job.MaxRuns > 0 && job.RunCount >= job.MaxRuns {
			job.Status = JobStatusCompleted
			ps.emitEvent(JobEventCompleted, job, nil)
		} else if job.Type == JobTypeOneOff {
			job.Status = JobStatusCompleted
			ps.emitEvent(JobEventCompleted, job, nil)
		} else {
			// Schedule next run
			job.Status = JobStatusPending
			switch job.Type {
			case JobTypeInterval:
				job.NextRunAt = time.Now().Add(job.Interval)
			case JobTypeCron:
				// TODO: Parse cron expression
				job.NextRunAt = time.Now().Add(time.Hour)
			}
			ps.emitEvent(JobEventCompleted, job, nil)
		}
	}
	
	// Persist updated job
	if err := ps.store.Update(job); err != nil {
		fmt.Printf("⚠️ Failed to persist job %s: %v\n", job.ID, err)
	}
}

// defaultExecutor is the fallback executor
func (ps *PersistentScheduler) defaultExecutor(ctx context.Context, job *ScheduledJob) error {
	fmt.Printf("⏰ Executing job: %s (%s)\n", job.Name, job.ID)
	return nil
}

// ScheduleJob adds a new job to the scheduler
func (ps *PersistentScheduler) ScheduleJob(job *ScheduledJob) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	if job.ID == "" {
		job.ID = generateID("job")
	}
	
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	job.Status = JobStatusPending
	
	// Save to store
	if err := ps.store.Save(job); err != nil {
		return fmt.Errorf("failed to persist job: %w", err)
	}
	
	ps.jobs[job.ID] = job
	ps.emitEvent(JobEventScheduled, job, nil)
	
	return nil
}

// PauseJob pauses a job
func (ps *PersistentScheduler) PauseJob(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	job, exists := ps.jobs[id]
	if !exists {
		return fmt.Errorf("job not found: %s", id)
	}
	
	job.Status = JobStatusPaused
	job.UpdatedAt = time.Now()
	
	if err := ps.store.Update(job); err != nil {
		return fmt.Errorf("failed to persist job: %w", err)
	}
	
	ps.emitEvent(JobEventPaused, job, nil)
	return nil
}

// ResumeJob resumes a paused job
func (ps *PersistentScheduler) ResumeJob(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	job, exists := ps.jobs[id]
	if !exists {
		return fmt.Errorf("job not found: %s", id)
	}
	
	if job.Status != JobStatusPaused {
		return fmt.Errorf("job is not paused: %s", id)
	}
	
	job.Status = JobStatusPending
	job.UpdatedAt = time.Now()
	
	if err := ps.store.Update(job); err != nil {
		return fmt.Errorf("failed to persist job: %w", err)
	}
	
	ps.emitEvent(JobEventResumed, job, nil)
	return nil
}

// DeleteJob removes a job
func (ps *PersistentScheduler) DeleteJob(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	job, exists := ps.jobs[id]
	if !exists {
		return fmt.Errorf("job not found: %s", id)
	}
	
	if err := ps.store.Delete(id); err != nil {
		return fmt.Errorf("failed to delete job from store: %w", err)
	}
	
	delete(ps.jobs, id)
	ps.emitEvent(JobEventDeleted, job, nil)
	
	return nil
}

// GetJob retrieves a job by ID
func (ps *PersistentScheduler) GetJob(id string) (*ScheduledJob, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	job, exists := ps.jobs[id]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", id)
	}
	
	return job, nil
}

// GetAllJobs returns all jobs
func (ps *PersistentScheduler) GetAllJobs() []*ScheduledJob {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	jobs := make([]*ScheduledJob, 0, len(ps.jobs))
	for _, job := range ps.jobs {
		jobs = append(jobs, job)
	}
	
	return jobs
}

// GetPendingJobs returns all pending jobs
func (ps *PersistentScheduler) GetPendingJobs() []*ScheduledJob {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	var jobs []*ScheduledJob
	for _, job := range ps.jobs {
		if job.Status == JobStatusPending {
			jobs = append(jobs, job)
		}
	}
	
	return jobs
}

// =============================================================================
// ECHOBEATS INTEGRATION
// =============================================================================

// PersistentEchobeatsScheduler wraps PersistentScheduler for echobeats-specific functionality
type PersistentEchobeatsScheduler struct {
	*PersistentScheduler
	cognitiveLoop *UnifiedCognitiveLoopV2
}

// NewPersistentEchobeatsScheduler creates a new persistent echobeats scheduler
func NewPersistentEchobeatsScheduler(ctx context.Context, storePath string) (*PersistentEchobeatsScheduler, error) {
	store, err := NewFileJobStore(storePath)
	if err != nil {
		return nil, err
	}
	
	scheduler := NewPersistentScheduler(ctx, store)
	
	es := &PersistentEchobeatsScheduler{
		PersistentScheduler: scheduler,
	}
	
	// Register echobeats executors
	es.RegisterExecutor("cognitive_beat", es.executeCognitiveBeat)
	es.RegisterExecutor("knowledge_integration", es.executeKnowledgeIntegration)
	es.RegisterExecutor("wisdom_cultivation", es.executeWisdomCultivation)
	es.RegisterExecutor("dream_processing", es.executeDreamProcessing)
	es.RegisterExecutor("interest_update", es.executeInterestUpdate)
	
	return es, nil
}

// SetCognitiveLoop sets the cognitive loop for the scheduler
func (es *PersistentEchobeatsScheduler) SetCognitiveLoop(loop *UnifiedCognitiveLoopV2) {
	es.cognitiveLoop = loop
}

// ScheduleCognitiveBeat schedules a cognitive beat
func (es *PersistentEchobeatsScheduler) ScheduleCognitiveBeat(interval time.Duration) error {
	job := &ScheduledJob{
		Name:     "Cognitive Beat",
		Type:     JobTypeInterval,
		Interval: interval,
		Context:  "cognitive_beat",
		Priority: 10,
		NextRunAt: time.Now().Add(interval),
		Payload: map[string]interface{}{
			"beat_type": "standard",
		},
	}
	
	return es.ScheduleJob(job)
}

// ScheduleKnowledgeIntegration schedules knowledge integration
func (es *PersistentEchobeatsScheduler) ScheduleKnowledgeIntegration(interval time.Duration) error {
	job := &ScheduledJob{
		Name:     "Knowledge Integration",
		Type:     JobTypeInterval,
		Interval: interval,
		Context:  "knowledge_integration",
		Priority: 5,
		NextRunAt: time.Now().Add(interval),
		Payload: map[string]interface{}{
			"integration_type": "incremental",
		},
	}
	
	return es.ScheduleJob(job)
}

// ScheduleWisdomCultivation schedules wisdom cultivation
func (es *PersistentEchobeatsScheduler) ScheduleWisdomCultivation(interval time.Duration) error {
	job := &ScheduledJob{
		Name:     "Wisdom Cultivation",
		Type:     JobTypeInterval,
		Interval: interval,
		Context:  "wisdom_cultivation",
		Priority: 3,
		NextRunAt: time.Now().Add(interval),
		Payload: map[string]interface{}{
			"cultivation_type": "balance_optimization",
		},
	}
	
	return es.ScheduleJob(job)
}

// executeCognitiveBeat executes a cognitive beat job
func (es *PersistentEchobeatsScheduler) executeCognitiveBeat(ctx context.Context, job *ScheduledJob) error {
	fmt.Printf("💓 Executing cognitive beat: %s\n", job.ID)
	
	if es.cognitiveLoop != nil {
		// Trigger a cognitive cycle
		// This would integrate with the actual cognitive loop
	}
	
	return nil
}

// executeKnowledgeIntegration executes knowledge integration
func (es *PersistentEchobeatsScheduler) executeKnowledgeIntegration(ctx context.Context, job *ScheduledJob) error {
	fmt.Printf("📚 Executing knowledge integration: %s\n", job.ID)
	return nil
}

// executeWisdomCultivation executes wisdom cultivation
func (es *PersistentEchobeatsScheduler) executeWisdomCultivation(ctx context.Context, job *ScheduledJob) error {
	fmt.Printf("🧘 Executing wisdom cultivation: %s\n", job.ID)
	return nil
}

// executeDreamProcessing executes dream processing
func (es *PersistentEchobeatsScheduler) executeDreamProcessing(ctx context.Context, job *ScheduledJob) error {
	fmt.Printf("💭 Executing dream processing: %s\n", job.ID)
	return nil
}

// executeInterestUpdate executes interest pattern update
func (es *PersistentEchobeatsScheduler) executeInterestUpdate(ctx context.Context, job *ScheduledJob) error {
	fmt.Printf("🎯 Executing interest update: %s\n", job.ID)
	return nil
}
