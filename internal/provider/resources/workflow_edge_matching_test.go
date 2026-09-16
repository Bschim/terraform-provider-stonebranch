package resources

import "testing"

func edg(sourceId, sourceTask, targetId, targetTask string) WorkflowEdgeResponseModel {
	return WorkflowEdgeResponseModel{
		SourceId: &EdgeVertexRefResp{Value: sourceId, TaskName: sourceTask},
		TargetId: &EdgeVertexRefResp{Value: targetId, TaskName: targetTask},
	}
}

func TestMatchEdge_ByVertexId(t *testing.T) {
	edges := []WorkflowEdgeResponseModel{
		edg("10", "START", "11", "MIDDLE"),
		edg("11", "MIDDLE", "12", "END"),
	}

	got := matchEdge(edges, "11", "12", "", "")
	if got == nil || got.SourceId.Value != "11" || got.TargetId.Value != "12" {
		t.Fatalf("expected edge 11->12, got %+v", got)
	}
}

func TestMatchEdge_FallsBackToTaskNames(t *testing.T) {
	edges := []WorkflowEdgeResponseModel{
		// Vertex IDs have shifted since state was written (21->12 renumbered
		// elsewhere), but the task-name pair still identifies the same edge.
		edg("20", "START", "21", "MIDDLE"),
	}

	got := matchEdge(edges, "10", "12", "START", "MIDDLE")
	if got == nil || got.SourceId.Value != "20" || got.TargetId.Value != "21" {
		t.Fatalf("expected fallback match on task names, got %+v", got)
	}
}

func TestMatchEdge_IdMatchRejectedWhenTaskNamesDisagree(t *testing.T) {
	edges := []WorkflowEdgeResponseModel{
		// Vertices got renumbered elsewhere in the workflow: IDs 4/5 now
		// belong to a completely different task pair than what state
		// recorded them for. The (4,5) pair still exists live, so a pure
		// ID match would silently pick the wrong edge unless task names
		// (already known from a prior Read) are cross-checked.
		edg("4", "PGECHLOI", "5", "PGECHIPR"),
		edg("2", "SRECHIPS", "3", "PGTRTLO2"),
	}

	got := matchEdge(edges, "4", "5", "SRECHIPS", "PGTRTLO2")
	if got == nil || got.SourceId.Value != "2" || got.TargetId.Value != "3" {
		t.Fatalf("expected fallback match on task names (2->3), got %+v", got)
	}
}

func TestMatchEdge_NoMatch(t *testing.T) {
	edges := []WorkflowEdgeResponseModel{
		edg("20", "START", "21", "MIDDLE"),
	}

	got := matchEdge(edges, "10", "12", "OTHER_SOURCE", "OTHER_TARGET")
	if got != nil {
		t.Fatalf("expected no match, got %+v", got)
	}
}

func TestMatchEdge_NoTaskNamesProvided_NoFallback(t *testing.T) {
	edges := []WorkflowEdgeResponseModel{
		edg("20", "START", "21", "MIDDLE"),
	}

	// IDs don't match and no task names are given (e.g. right after Create,
	// or ImportState by vertex ID) - must not fall back blindly.
	got := matchEdge(edges, "10", "12", "", "")
	if got != nil {
		t.Fatalf("expected no match without task names, got %+v", got)
	}
}

func TestCanConfirmEdgeDeleted(t *testing.T) {
	cases := []struct {
		name                           string
		sourceTaskName, targetTaskName string
		want                           bool
	}{
		{"both known - fallback was attempted, no match is trustworthy", "START", "MIDDLE", true},
		{"source missing - fallback was skipped, can't trust", "", "MIDDLE", false},
		{"target missing - fallback was skipped, can't trust", "START", "", false},
		{"both missing - never backfilled, can't trust", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := canConfirmEdgeDeleted(tc.sourceTaskName, tc.targetTaskName)
			if got != tc.want {
				t.Fatalf("canConfirmEdgeDeleted(%q, %q) = %v, want %v", tc.sourceTaskName, tc.targetTaskName, got, tc.want)
			}
		})
	}
}
