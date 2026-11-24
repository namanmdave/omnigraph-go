package omnigraph

import "encoding/json"

type DispositionOutput struct {
	Action  string `json:"action"`
	Effects string `json:"effects"`
}

type ErrorOutput struct {
	Code                            string             `json:"code"`
	Error                           string             `json:"error"`
	Disposition                     *DispositionOutput `json:"disposition"`
	MergeConflicts                  []any              `json:"merge_conflicts,omitempty"`
	PublishedDatasetVersionConflict any                `json:"published_dataset_version_conflict,omitempty"`
	ReadSetConflict                 any                `json:"read_set_conflict,omitempty"`
	KeyConflict                     any                `json:"key_conflict,omitempty"`
	ChangeDiffRefusal               any                `json:"change_diff_refusal,omitempty"`
	ChangeFeedGap                   any                `json:"change_feed_gap,omitempty"`
	FullTextIndexRebuildRequired    any                `json:"full_text_index_rebuild_required,omitempty"`
	PreconditionFailure             any                `json:"precondition_failure,omitempty"`
	BlobPreconditionFailure         any                `json:"blob_precondition_failure,omitempty"`
	Diagnostic                      any                `json:"diagnostic,omitempty"`
}

type HealthOutput struct {
	Status                string  `json:"status"`
	Version               string  `json:"version"`
	SourceVersion         *string `json:"source_version,omitempty"`
	InternalSchemaVersion int     `json:"internal_schema_version"`
}

type ReadinessOutput struct {
	Status                string  `json:"status"`
	Ready                 bool    `json:"ready"`
	ReadyGraphCount       int     `json:"ready_graph_count"`
	LoadingGraphCount     int     `json:"loading_graph_count"`
	BlockedGraphCount     int     `json:"blocked_graph_count"`
	ServedGraphCount      int     `json:"served_graph_count"`
	ShutdownGraceSeconds  int     `json:"shutdown_grace_seconds"`
	StateRevision         int     `json:"state_revision"`
	StateCAS              *string `json:"state_cas,omitempty"`
	BootedServingDigest   *string `json:"booted_serving_digest,omitempty"`
}

type AccessInput struct {
	Branch       string
	TargetBranch string
}

type AccessOutput struct {
	GraphID      string   `json:"graph_id"`
	Actions      []string `json:"actions"`
	GraphWide    []string `json:"graph_wide"`
	ActorID      *string  `json:"actor_id,omitempty"`
	Branch       *string  `json:"branch,omitempty"`
	TargetBranch *string  `json:"target_branch,omitempty"`
}

type SnapshotInput struct {
	Branch string
}

type SnapshotDatasetOutput struct {
	Name    string `json:"name"`
	Version int    `json:"version,omitempty"`
}

type SnapshotOutput struct {
	GraphBranch           string                  `json:"graph_branch"`
	GraphManifestVersion  int                     `json:"graph_manifest_version"`
	InternalSchemaVersion int                     `json:"internal_schema_version"`
	Datasets              []SnapshotDatasetOutput `json:"datasets"`
}

type SettingsRequest struct {
	Engine              string `json:"engine,omitempty"`
	MergeLineage        string `json:"merge_lineage,omitempty"`
	ANNNProbes          int    `json:"ann_nprobes,omitempty"`
	HistoryReleaseBytes int    `json:"history_release_bytes,omitempty"`
	LoadSourceMaxBytes  int    `json:"load_source_max_bytes,omitempty"`
	QueryMemoryBytes    int    `json:"query_memory_bytes,omitempty"`
	TraversalWorkLimit  int    `json:"traversal_work_limit,omitempty"`
	WriteMaxBytes       int    `json:"write_max_bytes,omitempty"`
}

type QueryInput struct {
	Query    string           `json:"query"`
	Name     string           `json:"name,omitempty"`
	Params   map[string]any   `json:"params,omitempty"`
	Branch   string           `json:"branch,omitempty"`
	Snapshot string           `json:"snapshot,omitempty"`
	Settings *SettingsRequest `json:"settings,omitempty"`
}

type ReadTargetOutput struct {
	Branch   string `json:"branch,omitempty"`
	Snapshot string `json:"snapshot,omitempty"`
}

type ReadOutput struct {
	QueryName     string           `json:"query_name"`
	RowCount      int              `json:"row_count"`
	Rows          any              `json:"rows"`
	Columns       []any            `json:"columns,omitempty"`
	GraphCommitID *string          `json:"graph_commit_id,omitempty"`
	Page          any              `json:"page,omitempty"`
	Target        ReadTargetOutput `json:"target"`
}

type MutationInput struct {
	Query    string           `json:"query"`
	Name     string           `json:"name,omitempty"`
	Params   map[string]any   `json:"params,omitempty"`
	Branch   string           `json:"branch,omitempty"`
	Settings *SettingsRequest `json:"settings,omitempty"`
}

type ChangeOutput struct {
	QueryName     string        `json:"query_name"`
	Branch        string        `json:"branch"`
	AffectedNodes int           `json:"affected_nodes"`
	AffectedEdges int           `json:"affected_edges"`
	ActorID       *string       `json:"actor_id,omitempty"`
	Commit        *CommitOutput `json:"commit"`
	Executed      any           `json:"executed,omitempty"`
	Outcome       any           `json:"outcome,omitempty"`
}

type IngestInput struct {
	Data   string `json:"data"`
	Branch string `json:"branch,omitempty"`
	From   string `json:"from,omitempty"`
	Mode   any    `json:"mode,omitempty"`
}

type IngestOutput struct {
	URI                 string   `json:"uri"`
	Branch              string   `json:"branch"`
	BranchCreated       bool     `json:"branch_created"`
	BaseBranch          *string  `json:"base_branch,omitempty"`
	ActorID             *string  `json:"actor_id,omitempty"`
	Commit              any      `json:"commit,omitempty"`
	Mode                string   `json:"mode"`
	Nodes               []any    `json:"nodes"`
	Edges               []any    `json:"edges"`
	TotalEntities       int      `json:"total_entities"`
	EmbeddingGeneration any      `json:"embedding_generation,omitempty"`
}

type GraphBatchLoadOutput struct {
	Branch              string  `json:"branch"`
	BranchCreated       bool    `json:"branch_created"`
	BaseBranch          *string `json:"base_branch,omitempty"`
	ActorID             *string `json:"actor_id,omitempty"`
	Commit              any     `json:"commit,omitempty"`
	Mode                string  `json:"mode"`
	Nodes               []any   `json:"nodes"`
	Edges               []any   `json:"edges"`
	TotalEntities       int     `json:"total_entities"`
	EmbeddingGeneration any     `json:"embedding_generation,omitempty"`
}

type ExportInput struct {
	Branch    string   `json:"branch,omitempty"`
	TypeNames []string `json:"type_names,omitempty"`
}

type ExportRow struct {
	Type string         `json:"type,omitempty"`
	Edge string         `json:"edge,omitempty"`
	From any            `json:"from,omitempty"`
	To   any            `json:"to,omitempty"`
	Data map[string]any `json:"data,omitempty"`
}

type DiffInput struct {
	From      string
	To        string
	Kind      []string
	Op        []string
	Type      []string
	PageToken string
	Limit     int
}

type DiffOutput struct {
	FromCommitID  string  `json:"from_commit_id"`
	ToCommitID    string  `json:"to_commit_id"`
	BaseCommitID  *string `json:"base_commit_id,omitempty"`
	Changes       []any   `json:"changes"`
	NextPageToken *string `json:"next_page_token,omitempty"`
}

type GraphInfo struct {
	GraphID        string `json:"graph_id"`
	URI            string `json:"uri"`
	State          string `json:"state"`
	Action         string `json:"action"`
	ReadAvailable  bool   `json:"read_available"`
	WriteAvailable bool   `json:"write_available"`
	Failure        any    `json:"failure,omitempty"`
}

type GraphListOutput struct {
	Graphs []GraphInfo `json:"graphs"`
}

type GraphDiscoveryEntry struct {
	GraphID     string `json:"graph_id"`
	DisplayName string `json:"display_name"`
}

type GraphDiscoveryOutput struct {
	Graphs []GraphDiscoveryEntry `json:"graphs"`
}

type BranchListOutput struct {
	Branches []string `json:"branches"`
}

type BranchCreateInput struct {
	Name       string `json:"name"`
	From       string `json:"from,omitempty"`
	FromCommit string `json:"from_commit,omitempty"`
}

type BranchCreateOutput struct {
	Name         string  `json:"name"`
	URI          string  `json:"uri"`
	From         string  `json:"from"`
	FromCommitID string  `json:"from_commit_id"`
	ActorID      *string `json:"actor_id,omitempty"`
}

type BranchDeleteOutput struct {
	Name    string  `json:"name"`
	URI     string  `json:"uri"`
	ActorID *string `json:"actor_id,omitempty"`
}

type BranchMergeInput struct {
	Source       string           `json:"source"`
	Target       string           `json:"target,omitempty"`
	AsCommit     bool             `json:"as_commit,omitempty"`
	DeleteBranch bool             `json:"delete_branch,omitempty"`
	DryRun       bool             `json:"dry_run,omitempty"`
	Settings     *SettingsRequest `json:"settings,omitempty"`
}

type BranchMergeOutput struct {
	Source            string        `json:"source"`
	Target            string        `json:"target"`
	Outcome           string        `json:"outcome"`
	Commit            *CommitOutput `json:"commit"`
	DryRun            bool          `json:"dry_run,omitempty"`
	BranchDeleted     *bool         `json:"branch_deleted,omitempty"`
	BranchDeleteError any           `json:"branch_delete_error_details,omitempty"`
	ActorID           *string       `json:"actor_id,omitempty"`
}

type BranchDiffInput struct {
	Source    string
	Target    string
	Kind      []string
	Op        []string
	Type      []string
	PageToken string
	Limit     int
}

type MergeBaseInput struct {
	Source string
	Target string
}

type CommitOutput struct {
	GraphCommitID        string  `json:"graph_commit_id"`
	Kind                 string  `json:"kind"`
	CreatedAt            int64   `json:"created_at"`
	GraphManifestVersion int     `json:"graph_manifest_version"`
	GraphBranch          *string `json:"graph_branch,omitempty"`
	ParentCommitID       *string `json:"parent_commit_id,omitempty"`
	MergedParentCommitID *string `json:"merged_parent_commit_id,omitempty"`
	RootCommitID         *string `json:"root_commit_id,omitempty"`
	ActorID              *string `json:"actor_id,omitempty"`
	Published            bool    `json:"published,omitempty"`
}

type MergeBaseOutput struct {
	Source         string       `json:"source"`
	Target         string       `json:"target"`
	SourceCommitID string       `json:"source_commit_id"`
	TargetCommitID string       `json:"target_commit_id"`
	MergeBase      CommitOutput `json:"merge_base"`
}

type ListCommitsInput struct {
	Branch    string
	PageToken string
	Limit     int
}

type CommitListOutput struct {
	Commits       []any   `json:"commits"`
	HeadCommitID  *string `json:"head_commit_id,omitempty"`
	NextPageToken *string `json:"next_page_token,omitempty"`
}

type CommitChangesOutput struct {
	Cause         any     `json:"cause"`
	Changes       []any   `json:"changes"`
	NextPageToken *string `json:"next_page_token,omitempty"`
}

type PollChangesInput struct {
	Cursor    string
	Branch    string
	Kind      []string
	Op        []string
	Type      []string
	PageToken string
	Limit     int
}

type ChangeFeedOutput struct {
	Blocks        []any   `json:"blocks"`
	CaughtUp      *bool   `json:"caught_up,omitempty"`
	Cursor        *string `json:"cursor,omitempty"`
	NextPageToken *string `json:"next_page_token,omitempty"`
}

type ChangeBaselineInput struct {
	Branch string   `json:"branch,omitempty"`
	Kind   []string `json:"kind,omitempty"`
	Op     []string `json:"op,omitempty"`
	Type   []string `json:"type,omitempty"`
}

type ChangeBaselineOutput struct {
	ResumeCursor     string `json:"resume_cursor"`
	SnapshotCommitID string `json:"snapshot_commit_id"`
}

type SchemaOutput struct {
	SchemaSource    string  `json:"schema_source"`
	CatalogueSource *string `json:"catalogue_source,omitempty"`
	SystemColumns   any     `json:"system_columns,omitempty"`
}

type ParamDescriptor struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Nullable  bool   `json:"nullable"`
	ItemKind  any    `json:"item_kind,omitempty"`
	VectorDim *int   `json:"vector_dim,omitempty"`
}

type QueryCatalogEntry struct {
	Name        string            `json:"name"`
	ToolName    string            `json:"tool_name"`
	Mutation    bool              `json:"mutation"`
	Description *string           `json:"description,omitempty"`
	Instruction *string           `json:"instruction,omitempty"`
	Params      []ParamDescriptor `json:"params"`
}

type QueriesCatalogOutput struct {
	Queries []QueryCatalogEntry `json:"queries"`
}

type InvokeQueryInput struct {
	Branch         string         `json:"branch,omitempty"`
	Snapshot       string         `json:"snapshot,omitempty"`
	Params         map[string]any `json:"params,omitempty"`
	ExpectMutation *bool          `json:"expect_mutation,omitempty"`
}

type InvokeQueryOutput = json.RawMessage

type BlobInput struct {
	Type   string
	Key    string
	Field  string
	Branch string
	Range  string
}

type BlobWriteInput struct {
	Type        string
	Key         string
	Field       string
	Branch      string
	IfMatch     string
	IfNoneMatch string
}

type BlobWriteOutput struct {
	Kind     string  `json:"kind"`
	Branch   string  `json:"branch"`
	Selector any     `json:"selector"`
	Commit   any     `json:"commit"`
	ETag     *string `json:"etag,omitempty"`
	Size     *int64  `json:"size,omitempty"`
	ActorID  *string `json:"actor_id,omitempty"`
}

type DeploymentInput struct {
	DeploymentID string         `json:"deployment_id"`
	Deployment   map[string]any `json:"deployment"`
}

type DeploymentPlanInput struct {
	Deployment map[string]any `json:"deployment"`
}

type DeploymentOutput struct {
	Active     bool           `json:"active"`
	InProgress bool           `json:"in_progress"`
	Deployment map[string]any `json:"deployment"`
}

type DeploymentStatusOutput struct {
	Active     bool           `json:"active"`
	InProgress bool           `json:"in_progress"`
	Status     map[string]any `json:"status"`
}
