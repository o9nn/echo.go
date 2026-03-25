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

type JobType string
const (
	JobTypeOneOff   JobType = "one_off"
	JobTypeInterval JobType = "interval"
	JobTypeCron     JobType = "cron"
)

type JobStatus string
const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusPaused    JobStatus = "paused"
)

type ScheduledJob struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      JobType                `json:"type"`
	Status    JobStatus              `json:"status"`
	Priority  int                    `json:"priority"`
	Interval  time.Duration          `json:"interval,omitempty"`
	CronExpr  string                 `json:"cron_expr,omitempty"`
	NextRunAt time.Time              `json:"next_run_at"`
	LastRunAt time.Time              `json:"last_run_at,omitempty"`
	Payload   map[string]interface{} `json:"payload"`
	GoalID    string                 `json:"goal_id,omitempty"`
	Context   string                 `json:"context"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	RunCount  int                    `json:"run_count"`
	MaxRuns   int                    `json:"max_runs,omitempty"`
	LastError string                 `json:"last_error,omitempty"`
	ErrorCount int                   `json:"error_count"`
}

type JobStore interface {
	Save(job *ScheduledJob) error
	Load(id string) (*ScheduledJob, error)
	LoadAll() ([]*ScheduledJob, error)
	Delete(id string) error
	Update(job *ScheduledJob) error
}

type FileJobStore struct {
	mu       sync.RWMutex
	basePath string
}

func NewFileJobStore(basePath string) (*FileJobStore, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create job store directory: %w", err)
	}
	return &FileJobStore{basePath: basePath}, nil
}

func (s *FileJobStore) Save(job *ScheduledJob) error {
	s.mu.Lock(); defer s.mu.Unlock()
	job.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil { return fmt.Errorf("failed to marshal job: %w", err) }
	path := filepath.Join(s.basePath, job.ID+".json")
	if err := os.WriteFile(path, data, 0644); err != nil { return fmt.Errorf("failed to write job file: %w", err) }
	return nil
}

func (s *FileJobStore) Load(id string) (*ScheduledJob, error) {
	s.mu.RLock(); defer s.mu.RUnlock()
	path := filepath.Join(s.basePath, id+".json")
	data, err := os.ReadFile(path)
	if err != nil { return nil, fmt.Errorf("failed to read job file: %w", err) }
	var job ScheduledJob
	if err := json.Unmarshal(data, &job); err != nil { return nil, fmt.Errorf("failed to unmarshal job: %w", err) }
	return &job, nil
}

func (s *FileJobStore) LoadAll() ([]*ScheduledJob, error) {
	s.mu.RLock(); defer s.mu.RUnlock()
	entries, err := os.ReadDir(s.basePath)
	if err != nil { return nil, fmt.Errorf("failed to read job store directory: %w", err) }
	var jobs []*ScheduledJob
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" { continue }
		path := filepath.Join(s.basePath, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil { continue }
		var job ScheduledJob
		if err := json.Unmarshal(data, &job); err != nil { continue }
		jobs = append(jobs, &job)
	}
	return jobs, nil
}

func (s *FileJobStore) Delete(id string) error {
	s.mu.Lock(); defer s.mu.Unlock()
	path := filepath.Join(s.basePath, id+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) { return fmt.Errorf("failed to delete job file: %w", err) }
	return nil
}

func (s *FileJobStore) Update(job *ScheduledJob) error { return s.Save(job) }

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

type JobEvent struct { Type JobEventType; Job *ScheduledJob; Timestamp time.Time; Error error }
type JobEventListener func(event *JobEvent)
type JobExecutor func(ctx context.Context, job *ScheduledJob) error

type PersistentScheduler struct {
	mu sync.RWMutex; ctx context.Context; cancel context.CancelFunc
	store JobStore; jobs map[string]*ScheduledJob; executors map[string]JobExecutor
	listeners []JobEventListener; running bool; tickInterval time.Duration
}

func NewPersistentScheduler(ctx context.Context, store JobStore) *PersistentScheduler {
	ctx, cancel := context.WithCancel(ctx)
	return &PersistentScheduler{ctx: ctx, cancel: cancel, store: store, jobs: make(map[string]*ScheduledJob), executors: make(map[string]JobExecutor), listeners: make([]JobEventListener, 0), tickInterval: time.Second}
}

func (ps *PersistentScheduler) RegisterExecutor(context string, executor JobExecutor) { ps.mu.Lock(); defer ps.mu.Unlock(); ps.executors[context] = executor }
func (ps *PersistentScheduler) AddListener(listener JobEventListener) { ps.mu.Lock(); defer ps.mu.Unlock(); ps.listeners = append(ps.listeners, listener) }

func (ps *PersistentScheduler) emitEvent(eventType JobEventType, job *ScheduledJob, err error) {
	event := &JobEvent{Type: eventType, Job: job, Timestamp: time.Now(), Error: err}
	for _, listener := range ps.listeners { go listener(event) }
}

func (ps *PersistentScheduler) Start() error {
	ps.mu.Lock()
	if ps.running { ps.mu.Unlock(); return fmt.Errorf("scheduler already running") }
	ps.running = true; ps.mu.Unlock()
	if err := ps.recoverJobs(); err != nil { return fmt.Errorf("failed to recover jobs: %w", err) }
	go ps.schedulerLoop()
	fmt.Println("Persistent Echobeats Scheduler: Started")
	return nil
}

func (ps *PersistentScheduler) Stop() error {
	ps.mu.Lock(); defer ps.mu.Unlock()
	if !ps.running { return fmt.Errorf("scheduler not running") }
	ps.cancel(); ps.running = false
	return nil
}

func (ps *PersistentScheduler) recoverJobs() error {
	jobs, err := ps.store.LoadAll()
	if err != nil { return err }
	ps.mu.Lock(); defer ps.mu.Unlock()
	recovered := 0
	for _, job := range jobs {
		if job.Status == JobStatusCompleted || job.Status == JobStatusFailed { continue }
		if job.Status == JobStatusRunning { job.Status = JobStatusPending }
		if job.NextRunAt.Before(time.Now()) {
			switch job.Type {
			case JobTypeInterval: job.NextRunAt = time.Now().Add(job.Interval)
			case JobTypeCron:
				nextRun, err := parseCronNextRun(job.CronExpr, time.Now())
				if err != nil { job.NextRunAt = time.Now().Add(time.Minute) } else { job.NextRunAt = nextRun }
			default: job.NextRunAt = time.Now()
			}
		}
		ps.jobs[job.ID] = job; recovered++
	}
	fmt.Printf("Recovered %d jobs from persistent store\n", recovered)
	return nil
}

func (ps *PersistentScheduler) schedulerLoop() {
	ticker := time.NewTicker(ps.tickInterval); defer ticker.Stop()
	for { select { case <-ps.ctx.Done(): return; case <-ticker.C: ps.tick() } }
}

func (ps *PersistentScheduler) tick() {
	ps.mu.Lock(); defer ps.mu.Unlock()
	now := time.Now()
	for _, job := range ps.jobs {
		if job.Status != JobStatusPending { continue }
		if job.NextRunAt.After(now) { continue }
		go ps.executeJob(job)
	}
}

func (ps *PersistentScheduler) executeJob(job *ScheduledJob) {
	ps.mu.Lock(); job.Status = JobStatusRunning; job.LastRunAt = time.Now(); ps.mu.Unlock()
	ps.emitEvent(JobEventStarted, job, nil)
	executor, exists := ps.executors[job.Context]
	if !exists { executor = ps.defaultExecutor }
	ctx, cancel := context.WithTimeout(ps.ctx, 5*time.Minute); defer cancel()
	err := executor(ctx, job)
	ps.mu.Lock(); defer ps.mu.Unlock()
	job.RunCount++
	if err != nil {
		job.LastError = err.Error(); job.ErrorCount++
		if job.ErrorCount >= 3 { job.Status = JobStatusFailed; ps.emitEvent(JobEventFailed, job, err) } else { job.Status = JobStatusPending; job.NextRunAt = time.Now().Add(time.Minute * time.Duration(job.ErrorCount)) }
	} else {
		job.LastError = ""
		if job.MaxRuns > 0 && job.RunCount >= job.MaxRuns { job.Status = JobStatusCompleted; ps.emitEvent(JobEventCompleted, job, nil) } else if job.Type == JobTypeOneOff { job.Status = JobStatusCompleted; ps.emitEvent(JobEventCompleted, job, nil) } else {
			job.Status = JobStatusPending
			switch job.Type {
			case JobTypeInterval: job.NextRunAt = time.Now().Add(job.Interval)
			case JobTypeCron:
				nextRun, err := parseCronNextRun(job.CronExpr, time.Now())
				if err != nil { job.NextRunAt = time.Now().Add(time.Hour) } else { job.NextRunAt = nextRun }
			}
			ps.emitEvent(JobEventCompleted, job, nil)
		}
	}
	if err := ps.store.Update(job); err != nil { fmt.Printf("Failed to persist job %s: %v\n", job.ID, err) }
}

func (ps *PersistentScheduler) defaultExecutor(ctx context.Context, job *ScheduledJob) error { fmt.Printf("Executing job: %s (%s)\n", job.Name, job.ID); return nil }

func (ps *PersistentScheduler) ScheduleJob(job *ScheduledJob) error {
	ps.mu.Lock(); defer ps.mu.Unlock()
	if job.ID == "" { job.ID = generateID("job") }
	job.CreatedAt = time.Now(); job.UpdatedAt = time.Now(); job.Status = JobStatusPending
	if err := ps.store.Save(job); err != nil { return fmt.Errorf("failed to persist job: %w", err) }
	ps.jobs[job.ID] = job; ps.emitEvent(JobEventScheduled, job, nil); return nil
}

func (ps *PersistentScheduler) PauseJob(id string) error {
	ps.mu.Lock(); defer ps.mu.Unlock()
	job, exists := ps.jobs[id]; if !exists { return fmt.Errorf("job not found: %s", id) }
	job.Status = JobStatusPaused; job.UpdatedAt = time.Now()
	if err := ps.store.Update(job); err != nil { return fmt.Errorf("failed to persist job: %w", err) }
	ps.emitEvent(JobEventPaused, job, nil); return nil
}

func (ps *PersistentScheduler) ResumeJob(id string) error {
	ps.mu.Lock(); defer ps.mu.Unlock()
	job, exists := ps.jobs[id]; if !exists { return fmt.Errorf("job not found: %s", id) }
	if job.Status != JobStatusPaused { return fmt.Errorf("job is not paused: %s", id) }
	job.Status = JobStatusPending; job.UpdatedAt = time.Now()
	if err := ps.store.Update(job); err != nil { return fmt.Errorf("failed to persist job: %w", err) }
	ps.emitEvent(JobEventResumed, job, nil); return nil
}

func (ps *PersistentScheduler) DeleteJob(id string) error {
	ps.mu.Lock(); defer ps.mu.Unlock()
	job, exists := ps.jobs[id]; if !exists { return fmt.Errorf("job not found: %s", id) }
	if err := ps.store.Delete(id); err != nil { return fmt.Errorf("failed to delete job from store: %w", err) }
	delete(ps.jobs, id); ps.emitEvent(JobEventDeleted, job, nil); return nil
}

func (ps *PersistentScheduler) GetJob(id string) (*ScheduledJob, error) { ps.mu.RLock(); defer ps.mu.RUnlock(); job, exists := ps.jobs[id]; if !exists { return nil, fmt.Errorf("job not found: %s", id) }; return job, nil }
func (ps *PersistentScheduler) GetAllJobs() []*ScheduledJob { ps.mu.RLock(); defer ps.mu.RUnlock(); jobs := make([]*ScheduledJob, 0, len(ps.jobs)); for _, job := range ps.jobs { jobs = append(jobs, job) }; return jobs }
func (ps *PersistentScheduler) GetPendingJobs() []*ScheduledJob { ps.mu.RLock(); defer ps.mu.RUnlock(); var jobs []*ScheduledJob; for _, job := range ps.jobs { if job.Status == JobStatusPending { jobs = append(jobs, job) } }; return jobs }

// =============================================================================
// ECHOBEATS INTEGRATION
// =============================================================================

type PersistentEchobeatsScheduler struct { *PersistentScheduler; cognitiveLoop *UnifiedCognitiveLoopV2 }

func NewPersistentEchobeatsScheduler(ctx context.Context, storePath string) (*PersistentEchobeatsScheduler, error) {
	store, err := NewFileJobStore(storePath); if err != nil { return nil, err }
	scheduler := NewPersistentScheduler(ctx, store)
	es := &PersistentEchobeatsScheduler{PersistentScheduler: scheduler}
	es.RegisterExecutor("cognitive_beat", es.executeCognitiveBeat)
	es.RegisterExecutor("knowledge_integration", es.executeKnowledgeIntegration)
	es.RegisterExecutor("wisdom_cultivation", es.executeWisdomCultivation)
	es.RegisterExecutor("dream_processing", es.executeDreamProcessing)
	es.RegisterExecutor("interest_update", es.executeInterestUpdate)
	return es, nil
}

func (es *PersistentEchobeatsScheduler) SetCognitiveLoop(loop *UnifiedCognitiveLoopV2) { es.cognitiveLoop = loop }

func (es *PersistentEchobeatsScheduler) ScheduleCognitiveBeat(interval time.Duration) error {
	return es.ScheduleJob(&ScheduledJob{Name: "Cognitive Beat", Type: JobTypeInterval, Interval: interval, Context: "cognitive_beat", Priority: 10, NextRunAt: time.Now().Add(interval), Payload: map[string]interface{}{"beat_type": "standard"}})
}
func (es *PersistentEchobeatsScheduler) ScheduleKnowledgeIntegration(interval time.Duration) error {
	return es.ScheduleJob(&ScheduledJob{Name: "Knowledge Integration", Type: JobTypeInterval, Interval: interval, Context: "knowledge_integration", Priority: 5, NextRunAt: time.Now().Add(interval), Payload: map[string]interface{}{"integration_type": "incremental"}})
}
func (es *PersistentEchobeatsScheduler) ScheduleWisdomCultivation(interval time.Duration) error {
	return es.ScheduleJob(&ScheduledJob{Name: "Wisdom Cultivation", Type: JobTypeInterval, Interval: interval, Context: "wisdom_cultivation", Priority: 3, NextRunAt: time.Now().Add(interval), Payload: map[string]interface{}{"cultivation_type": "balance_optimization"}})
}

func (es *PersistentEchobeatsScheduler) executeCognitiveBeat(ctx context.Context, job *ScheduledJob) error { return nil }
func (es *PersistentEchobeatsScheduler) executeKnowledgeIntegration(ctx context.Context, job *ScheduledJob) error { return nil }
func (es *PersistentEchobeatsScheduler) executeWisdomCultivation(ctx context.Context, job *ScheduledJob) error { return nil }
func (es *PersistentEchobeatsScheduler) executeDreamProcessing(ctx context.Context, job *ScheduledJob) error { return nil }
func (es *PersistentEchobeatsScheduler) executeInterestUpdate(ctx context.Context, job *ScheduledJob) error { return nil }

// =============================================================================
// CRON EXPRESSION PARSER
// =============================================================================
// Supports standard 5-field cron: minute hour day-of-month month day-of-week
// Shorthands: @hourly, @daily, @weekly, @monthly, @yearly
// Special chars: * (any), */N (every N), N-M (range), N,M (list)

type cronField struct { values map[int]bool }

func parseCronNextRun(expr string, after time.Time) (time.Time, error) {
	switch expr {
	case "@hourly": return after.Truncate(time.Hour).Add(time.Hour), nil
	case "@daily", "@midnight": return time.Date(after.Year(), after.Month(), after.Day()+1, 0, 0, 0, 0, after.Location()), nil
	case "@weekly":
		daysUntilSunday := (7 - int(after.Weekday())) % 7; if daysUntilSunday == 0 { daysUntilSunday = 7 }
		return time.Date(after.Year(), after.Month(), after.Day()+daysUntilSunday, 0, 0, 0, 0, after.Location()), nil
	case "@monthly": return time.Date(after.Year(), after.Month()+1, 1, 0, 0, 0, 0, after.Location()), nil
	case "@yearly", "@annually": return time.Date(after.Year()+1, 1, 1, 0, 0, 0, 0, after.Location()), nil
	}
	fields := splitCronFields(expr)
	if len(fields) != 5 { return time.Time{}, fmt.Errorf("invalid cron expression: expected 5 fields, got %d", len(fields)) }
	minuteField, err := parseCronField(fields[0], 0, 59); if err != nil { return time.Time{}, fmt.Errorf("invalid minute field: %w", err) }
	hourField, err := parseCronField(fields[1], 0, 23); if err != nil { return time.Time{}, fmt.Errorf("invalid hour field: %w", err) }
	domField, err := parseCronField(fields[2], 1, 31); if err != nil { return time.Time{}, fmt.Errorf("invalid day-of-month field: %w", err) }
	monthField, err := parseCronField(fields[3], 1, 12); if err != nil { return time.Time{}, fmt.Errorf("invalid month field: %w", err) }
	dowField, err := parseCronField(fields[4], 0, 6); if err != nil { return time.Time{}, fmt.Errorf("invalid day-of-week field: %w", err) }
	candidate := after.Truncate(time.Minute).Add(time.Minute)
	maxSearch := after.Add(366 * 24 * time.Hour)
	for candidate.Before(maxSearch) {
		if monthField.values[int(candidate.Month())] && domField.values[candidate.Day()] && hourField.values[candidate.Hour()] && minuteField.values[candidate.Minute()] && dowField.values[int(candidate.Weekday())] { return candidate, nil }
		if !monthField.values[int(candidate.Month())] { candidate = time.Date(candidate.Year(), candidate.Month()+1, 1, 0, 0, 0, 0, candidate.Location()); continue }
		if !domField.values[candidate.Day()] || !dowField.values[int(candidate.Weekday())] { candidate = time.Date(candidate.Year(), candidate.Month(), candidate.Day()+1, 0, 0, 0, 0, candidate.Location()); continue }
		if !hourField.values[candidate.Hour()] { candidate = candidate.Truncate(time.Hour).Add(time.Hour); continue }
		candidate = candidate.Add(time.Minute)
	}
	return time.Time{}, fmt.Errorf("no matching time found within search window for expression: %s", expr)
}

func splitCronFields(expr string) []string {
	var fields []string; current := ""
	for _, c := range expr { if c == ' ' || c == '\t' { if current != "" { fields = append(fields, current); current = "" } } else { current += string(c) } }
	if current != "" { fields = append(fields, current) }; return fields
}

func parseCronField(field string, minVal, maxVal int) (*cronField, error) {
	cf := &cronField{values: make(map[int]bool)}
	parts := splitOnComma(field)
	for _, part := range parts { if err := parseCronPart(part, minVal, maxVal, cf); err != nil { return nil, err } }
	return cf, nil
}

func parseCronPart(part string, minVal, maxVal int, cf *cronField) error {
	if len(part) > 2 && part[0] == '*' && part[1] == '/' {
		step := atoiSimple(part[2:]); if step <= 0 { return fmt.Errorf("invalid step value: %s", part) }
		for i := minVal; i <= maxVal; i += step { cf.values[i] = true }; return nil
	}
	if part == "*" { for i := minVal; i <= maxVal; i++ { cf.values[i] = true }; return nil }
	dashIdx := indexOfChar(part, '-')
	if dashIdx > 0 {
		slashIdx := indexOfChar(part[dashIdx:], '/'); rangeEnd := part[dashIdx+1:]; step := 1
		if slashIdx > 0 { rangeEnd = part[dashIdx+1 : dashIdx+slashIdx]; step = atoiSimple(part[dashIdx+slashIdx+1:]); if step <= 0 { return fmt.Errorf("invalid step in range: %s", part) } }
		start := atoiSimple(part[:dashIdx]); end := atoiSimple(rangeEnd)
		if start < minVal || end > maxVal || start > end { return fmt.Errorf("invalid range: %s", part) }
		for i := start; i <= end; i += step { cf.values[i] = true }; return nil
	}
	val := atoiSimple(part); if val < minVal || val > maxVal { return fmt.Errorf("value %d out of range [%d, %d]", val, minVal, maxVal) }
	cf.values[val] = true; return nil
}

func splitOnComma(s string) []string {
	var parts []string; current := ""
	for _, c := range s { if c == ',' { if current != "" { parts = append(parts, current) }; current = "" } else { current += string(c) } }
	if current != "" { parts = append(parts, current) }; return parts
}

func indexOfChar(s string, c byte) int { for i := 0; i < len(s); i++ { if s[i] == c { return i } }; return -1 }

func atoiSimple(s string) int {
	result := 0; negative := false; start := 0
	if len(s) > 0 && s[0] == '-' { negative = true; start = 1 }
	for i := start; i < len(s); i++ { if s[i] >= '0' && s[i] <= '9' { result = result*10 + int(s[i]-'0') } }
	if negative { return -result }; return result
}
