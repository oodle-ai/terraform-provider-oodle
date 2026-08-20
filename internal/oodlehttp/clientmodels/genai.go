package clientmodels

import "encoding/json"

// Models for the GenAI (llmops) APIs.
//
// Each struct serves as both the request and the response body for
// its object. Fields the server owns are tagged omitempty so they
// are not sent on create; fields the server merges rather than
// replaces are pointers, so that a false or a zero is distinguishable
// from "not supplied" and can actually be written.

// GenAIDataset is a named collection of evaluation inputs.
//
// Datasets have no update endpoint: name, description and metadata
// are fixed once created, and the resource replaces on change.
type GenAIDataset struct {
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	ItemCount   int64           `json:"itemCount,omitempty"`
	RunCount    int64           `json:"runCount,omitempty"`
	CreatedAt   string          `json:"createdAt,omitempty"`
	UpdatedAt   string          `json:"updatedAt,omitempty"`
}

func (d *GenAIDataset) GetID() string { return d.ID }

// GenAIDatasetItem is one row of a dataset: an input, and
// optionally the output it is expected to produce.
type GenAIDatasetItem struct {
	ID string `json:"id,omitempty"`
	// DatasetName is accepted on create only. The server answers
	// with DatasetID instead, so both are carried here.
	DatasetName         string          `json:"datasetName,omitempty"`
	DatasetID           string          `json:"datasetId,omitempty"`
	Input               json.RawMessage `json:"input,omitempty"`
	ExpectedOutput      json.RawMessage `json:"expectedOutput,omitempty"`
	Metadata            json.RawMessage `json:"metadata,omitempty"`
	Status              string          `json:"status,omitempty"`
	SourceTraceID       string          `json:"sourceTraceId,omitempty"`
	SourceObservationID string          `json:"sourceObservationId,omitempty"`
	CreatedAt           string          `json:"createdAt,omitempty"`
	UpdatedAt           string          `json:"updatedAt,omitempty"`
}

func (d *GenAIDatasetItem) GetID() string { return d.ID }

// GenAIDatasetSchedule runs a dataset's experiment on its own, so a
// regression is found when it lands rather than when the next person
// looks.
//
// A dataset carries at most one, which is why the endpoint is a
// singleton under the dataset's name with no id of its own: PUT
// replaces whatever is there.
type GenAIDatasetSchedule struct {
	// DatasetName addresses the schedule. It is not part of the
	// wire format — the server answers with DatasetID instead — so
	// the client puts it back on every response, and it is what
	// GetID returns.
	DatasetName string `json:"-"`

	ID        string `json:"id,omitempty"`
	DatasetID string `json:"datasetId,omitempty"`
	// Enabled carries no omitempty: false is meaningful here, and a
	// dropped field would silently re-enable a paused schedule.
	Enabled bool `json:"enabled"`
	// Mode is "calendar" or "interval". Empty means calendar.
	Mode string `json:"mode,omitempty"`
	// IntervalValue and IntervalUnit describe an interval schedule:
	// 30 + "minutes", 6 + "hours", 1 + "days". At least 5 minutes,
	// which is what the worker's poll cycle can honour, and at most
	// 365 days.
	IntervalValue int64  `json:"intervalValue,omitempty"`
	IntervalUnit  string `json:"intervalUnit,omitempty"`
	// Timezone is an IANA name; empty means UTC. Calendar mode only:
	// the times below are read in it, so a schedule follows daylight
	// saving rather than drifting by an hour twice a year.
	Timezone    string   `json:"timezone,omitempty"`
	Times       []string `json:"times,omitempty"`
	Weekdays    []string `json:"weekdays,omitempty"`
	DaysOfMonth []string `json:"daysOfMonth,omitempty"`
	// ExperimentConfig is the `llm-experiment` job config each firing
	// launches, the same shape POST jobs takes.
	ExperimentConfig json.RawMessage `json:"experimentConfig,omitempty"`
	NextRunAt        string          `json:"nextRunAt,omitempty"`
	LastRunAt        string          `json:"lastRunAt,omitempty"`
	// LastError is why the most recent launch did not start. A
	// scheduled run that cannot be queued has no job row and no
	// dataset run, so without this the schedule looks healthy and
	// simply never produces results.
	LastError string `json:"lastError,omitempty"`
	CreatedBy string `json:"createdBy,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// GetID returns the dataset name, which is how the endpoint is
// addressed. The schedule's own uuid appears in no path.
func (s *GenAIDatasetSchedule) GetID() string { return s.DatasetName }

// GenAIEvalTemplate is a reusable scoring definition: either an
// LLM-as-judge prompt or a code scorer. It is what the Oodle UI
// calls a Library template.
type GenAIEvalTemplate struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	// Type is "llm", "code" or "output_comparer", and cannot be
	// changed after create.
	Type               string          `json:"type,omitempty"`
	Prompt             string          `json:"prompt,omitempty"`
	Vars               []string        `json:"vars,omitempty"`
	OutputSchema       json.RawMessage `json:"outputSchema,omitempty"`
	ModelParams        json.RawMessage `json:"modelParams,omitempty"`
	SourceCode         string          `json:"sourceCode,omitempty"`
	SourceCodeLanguage string          `json:"sourceCodeLanguage,omitempty"`
	Version            int64           `json:"version,omitempty"`
	CreatedAt          string          `json:"createdAt,omitempty"`
}

func (e *GenAIEvalTemplate) GetID() string { return e.ID }

// GenAIEvaluationRule wires an eval template to live traffic: which
// spans it scores, how often, and how span fields map onto the
// template's variables. It is what the Oodle UI calls an Evaluator.
type GenAIEvaluationRule struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	// EvaluatorID is the eval template this rule runs.
	EvaluatorID           string          `json:"evaluatorId,omitempty"`
	TargetType            string          `json:"targetType,omitempty"`
	Filters               json.RawMessage `json:"filters,omitempty"`
	SamplingRate          *float64        `json:"samplingRate,omitempty"`
	MaxInvocationsPerHour *int64          `json:"maxInvocationsPerHour,omitempty"`
	VariableMapping       json.RawMessage `json:"variableMapping,omitempty"`
	LLMConnectionID       *string         `json:"llmConnectionId,omitempty"`
	ModelParams           json.RawMessage `json:"modelParams,omitempty"`
	Enabled               *bool           `json:"enabled,omitempty"`
	DependsOnRuleIDs      *[]string       `json:"dependsOnRuleIds,omitempty"`
	DatasetID             string          `json:"datasetId,omitempty"`
	CreatedAt             string          `json:"createdAt,omitempty"`
	UpdatedAt             string          `json:"updatedAt,omitempty"`
}

func (e *GenAIEvaluationRule) GetID() string { return e.ID }

// GenAILLMConnection is a provider credential the evaluators and
// experiments authenticate with.
//
// APIKey is write-only: the server encrypts it and never returns it,
// and an update that omits it leaves the stored key in place.
type GenAILLMConnection struct {
	ID                  string          `json:"id,omitempty"`
	Name                string          `json:"name"`
	Provider            string          `json:"provider,omitempty"`
	APIKey              string          `json:"apiKey,omitempty"`
	BaseURL             string          `json:"baseUrl,omitempty"`
	DefaultModel        string          `json:"defaultModel,omitempty"`
	CustomModels        []string        `json:"customModels,omitempty"`
	CustomHeaders       json.RawMessage `json:"customHeaders,omitempty"`
	DefaultParams       json.RawMessage `json:"defaultParams,omitempty"`
	EnableDefaultModels *bool           `json:"enableDefaultModels,omitempty"`
	IsDefault           *bool           `json:"isDefault,omitempty"`
	CreatedAt           string          `json:"createdAt,omitempty"`
	UpdatedAt           string          `json:"updatedAt,omitempty"`
}

func (c *GenAILLMConnection) GetID() string { return c.ID }

// GenAIPrompt is one version of a named prompt.
//
// Prompts are append-only: every create adds a version rather than
// replacing one, and applications resolve a prompt by label. Prompt
// carries a JSON string for type "text" and a JSON array of chat
// messages for type "chat".
type GenAIPrompt struct {
	ID            string          `json:"id,omitempty"`
	Name          string          `json:"name"`
	Prompt        json.RawMessage `json:"prompt,omitempty"`
	Type          string          `json:"type,omitempty"`
	Config        json.RawMessage `json:"config,omitempty"`
	Labels        []string        `json:"labels,omitempty"`
	Tags          []string        `json:"tags,omitempty"`
	CommitMessage string          `json:"commitMessage,omitempty"`
	Version       int64           `json:"version,omitempty"`
	CreatedAt     string          `json:"createdAt,omitempty"`
	UpdatedAt     string          `json:"updatedAt,omitempty"`
}

func (p *GenAIPrompt) GetID() string { return p.ID }
