package resources

import (
	"strings"
	"testing"
)

func vtx(taskName, vertexId, x, y string) WorkflowVertexResponseModel {
	return WorkflowVertexResponseModel{
		Task:     &TaskRefResponse{Value: taskName},
		VertexId: vertexId,
		VertexX:  x,
		VertexY:  y,
	}
}

func TestMatchVertex_UniqueTaskName(t *testing.T) {
	vertices := []WorkflowVertexResponseModel{
		vtx("TASK_A", "10", "100", "100"),
		vtx("TASK_B", "11", "200", "100"),
	}

	got, err := matchVertex(vertices, "TASK_B", "999", "0", "0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.VertexId != "11" {
		t.Fatalf("expected vertex 11, got %+v", got)
	}
}

func TestMatchVertex_NoMatch(t *testing.T) {
	vertices := []WorkflowVertexResponseModel{
		vtx("TASK_A", "10", "100", "100"),
	}

	got, err := matchVertex(vertices, "TASK_MISSING", "10", "100", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil match, got %+v", got)
	}
}

func TestMatchVertex_DuplicateTaskName_PriorIdTiebreak(t *testing.T) {
	vertices := []WorkflowVertexResponseModel{
		vtx("PGOUTSCT", "20", "100", "100"),
		vtx("PGOUTSCT", "21", "200", "100"),
		vtx("PGOUTSCT", "22", "300", "100"),
	}

	got, err := matchVertex(vertices, "PGOUTSCT", "21", "0", "0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.VertexId != "21" {
		t.Fatalf("expected vertex 21 (matched by prior vertex_id), got %+v", got)
	}
}

func TestMatchVertex_DuplicateTaskName_PositionTiebreak(t *testing.T) {
	vertices := []WorkflowVertexResponseModel{
		vtx("PGOUTSCT", "30", "100", "100"),
		vtx("PGOUTSCT", "31", "200", "100"),
		vtx("PGOUTSCT", "32", "300", "100"),
	}

	// Prior vertex_id (99) no longer exists among the candidates - UAC
	// renumbered it - but the position still uniquely identifies the vertex.
	got, err := matchVertex(vertices, "PGOUTSCT", "99", "200", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.VertexId != "31" {
		t.Fatalf("expected vertex 31 (matched by position), got %+v", got)
	}
}

func TestMatchVertex_DuplicateTaskName_TrulyAmbiguous(t *testing.T) {
	vertices := []WorkflowVertexResponseModel{
		vtx("PGOUTSCT", "40", "100", "100"),
		vtx("PGOUTSCT", "41", "200", "100"),
	}

	// Neither prior vertex_id nor position narrows this to one candidate.
	got, err := matchVertex(vertices, "PGOUTSCT", "99", "999", "999")
	if err == nil {
		t.Fatalf("expected an ambiguity error, got vertex %+v", got)
	}
	if !strings.Contains(err.Error(), "PGOUTSCT") {
		t.Fatalf("expected error to mention the task name, got: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil vertex alongside the error, got %+v", got)
	}
}
