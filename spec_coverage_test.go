package omnigraph_test

import (
	"encoding/json"
	"os"
	"testing"
)

func TestOpenAPISpecCoverage(t *testing.T) {
	raw, err := os.ReadFile("spec/openapi.json")
	if err != nil {
		t.Fatalf("read spec/openapi.json: %v", err)
	}

	var doc struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
		Paths map[string]map[string]struct {
			OperationID string `json:"operationId"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse spec/openapi.json: %v", err)
	}

	covered := map[string]bool{
		"oauthProtectedResourceMetadata":              true,
		"oauthProtectedResourceMetadataAtResourcePath": true,
		"deployment_status":                           true,
		"deployment_apply":                            true,
		"deployment_lookup":                           true,
		"deployment_plan":                             true,
		"listGraphs":                                  true,
		"discoverGraphs":                              true,
		"access":                                      true,
		"blobGet":                                     true,
		"blobPut":                                     true,
		"blobClear":                                   true,
		"blobStat":                                    true,
		"branchList":                                  true,
		"branchCreate":                                true,
		"branchDiff":                                  true,
		"branchMerge":                                 true,
		"branchMergeBase":                             true,
		"branchDelete":                                true,
		"changesPoll":                                 true,
		"changesBaseline":                             true,
		"commitList":                                  true,
		"commitShow":                                  true,
		"commitDrop":                                  true,
		"commitChanges":                               true,
		"commitPublish":                               true,
		"diff":                                        true,
		"export":                                      true,
		"load":                                        true,
		"loadNdjson":                                  true,
		"mutate":                                      true,
		"mutateIfGraphCommit":                         true,
		"queriesList":                                 true,
		"queryInvoke":                                 true,
		"queryInvokeIfGraphCommit":                    true,
		"query":                                       true,
		"schemaShow":                                  true,
		"snapshot":                                    true,
		"health":                                      true,
		"health_head":                                 true,
		"readiness":                                   true,
	}

	for path, methods := range doc.Paths {
		for method, op := range methods {
			if op.OperationID == "" {
				continue
			}
			if !covered[op.OperationID] {
				t.Errorf("uncovered OpenAPI operation %s %s (%s)", method, path, op.OperationID)
			}
		}
	}
}
