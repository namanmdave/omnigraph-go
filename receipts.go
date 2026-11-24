package omnigraph

func validMergeReceipt(out *BranchMergeOutput, in BranchMergeInput) bool {
	if out == nil || out.Source != in.Source {
		return false
	}
	target := in.Target
	if target == "" {
		target = "main"
	}
	if out.Target != target {
		return false
	}
	if in.DryRun {
		return out.DryRun && out.Commit == nil &&
			(out.Outcome == "already_up_to_date" || out.Outcome == "fast_forward" || out.Outcome == "merged")
	}
	if out.DryRun {
		return false
	}
	if out.Outcome == "already_up_to_date" {
		return out.Commit == nil
	}
	if out.Outcome != "fast_forward" && out.Outcome != "merged" {
		return false
	}
	if !validCommitReceipt(out.Commit, target) {
		return false
	}
	if out.Commit.ParentCommitID == nil || *out.Commit.ParentCommitID == "" ||
		out.Commit.MergedParentCommitID == nil || *out.Commit.MergedParentCommitID == "" {
		return false
	}
	if in.AsCommit {
		return out.Commit.Kind == "unpublished" && !out.Commit.Published
	}
	return out.Commit.Kind == "published"
}

func validMutationReceipt(out *ChangeOutput) bool {
	if out == nil {
		return false
	}
	if out.Commit == nil {
		return true
	}
	return out.Commit.GraphCommitID != "" && out.Commit.GraphManifestVersion > 0
}

func validCommitReceipt(c *CommitOutput, expectedBranch string) bool {
	if c == nil || c.GraphCommitID == "" || c.GraphManifestVersion <= 0 {
		return false
	}
	branch := "main"
	if c.GraphBranch != nil && *c.GraphBranch != "" {
		branch = *c.GraphBranch
	}
	return branch == expectedBranch
}
