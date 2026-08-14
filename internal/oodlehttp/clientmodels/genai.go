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

// GenAIEvalTemplate is a reusable scoring definition: either an
// LLM-as-judge prompt or a code scorer. It is what the Oodle UI
// calls a Library template.
type GenAIEvalTemplate struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	// Type is "llm" or "code" and cannot be changed after create.
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
