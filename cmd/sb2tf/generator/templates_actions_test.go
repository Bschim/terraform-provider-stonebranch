package generator

import (
	"bytes"
	"strings"
	"testing"
)

// minimalTaskData returns the minimal top-level fields required to execute
// a task_* template without erroring on unrelated required fields
// (task_file_monitor and task_unix only unconditionally require "name",
// plus the "_terraformResource"/"_resourceName" fields every template
// references).
func minimalTaskData(resourceType, resourceName string) map[string]interface{} {
	return map[string]interface{}{
		"name":               "test-task",
		"_terraformResource": resourceType,
		"_resourceName":      resourceName,
	}
}

func renderTemplate(t *testing.T, resourceType string, data map[string]interface{}) string {
	t.Helper()
	tmpl := templates[resourceType]
	if tmpl == nil {
		t.Fatalf("no template registered for resource type %q", resourceType)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template %q: %v", resourceType, err)
	}
	return buf.String()
}

// TestActionsEmailNotificationRegression mirrors the shape of task that
// originally motivated this fix (see CLAUDE.md's "actions" gotcha): a task
// with a configured emailNotifications entry that was silently dropped from
// generated HCL because none of the 13 task templates referenced .actions.
// Confirms the rendered HCL now contains the email notification's fields
// instead of silently dropping the whole `actions` block.
func TestActionsEmailNotificationRegression(t *testing.T) {
	data := minimalTaskData("stonebranch_task_file_monitor", "task_file_monitor_001")
	data["actions"] = map[string]interface{}{
		"emailNotifications": []interface{}{
			map[string]interface{}{
				"status":           "Failed",
				"emailConnection":  "example_email_connection",
				"emailTemplate":    "example_email_template",
				"listReportFormat": "PDF",
				"fileNumLines":     100,
			},
		},
	}

	out := renderTemplate(t, "task_file_monitor", data)

	for _, want := range []string{
		"email_notifications = [",
		`email_connection = "example_email_connection"`,
		`email_template = "example_email_template"`,
		`list_report_format = "PDF"`,
		"file_num_lines = 100",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered output missing %q\n--- output ---\n%s", want, out)
		}
	}
}

// TestActionsAllSubTypes covers all 5 action sub-types in a single fixture,
// including the nested report{} (under an email notification) and
// vertices[]/variables[] (under a system operation), and confirms every
// sub-type's list header and a representative nested field renders.
func TestActionsAllSubTypes(t *testing.T) {
	data := minimalTaskData("stonebranch_task_unix", "task_unix_001")
	data["actions"] = map[string]interface{}{
		"systemOperations": []interface{}{
			map[string]interface{}{
				"status":    "Success",
				"operation": "Launch Task",
				"task":      "downstream-task",
				"vertices": []interface{}{
					map[string]interface{}{
						"taskName":   "downstream-task",
						"vertexName": "vertex-1",
					},
				},
				"variables": []interface{}{
					map[string]interface{}{
						"name":  "MY_VAR",
						"value": "my-value",
					},
				},
			},
		},
		"emailNotifications": []interface{}{
			map[string]interface{}{
				"status":  "Failed",
				"subject": "Something failed",
				"report": map[string]interface{}{
					"title":    "My Report",
					"userName": "admin",
				},
			},
		},
		"abortActions": []interface{}{
			map[string]interface{}{
				"status":        "Running/Problems",
				"cancelProcess": true,
			},
		},
		"setVariableActions": []interface{}{
			map[string]interface{}{
				"status":        "Success",
				"variableName":  "RESULT",
				"variableValue": "ok-value",
			},
		},
		"snmpNotifications": []interface{}{
			map[string]interface{}{
				"status":      "Failed",
				"snmpManager": "my-snmp-manager",
				"severity":    "Critical",
			},
		},
	}

	out := renderTemplate(t, "task_unix", data)

	for _, want := range []string{
		"system_operations = [",
		"email_notifications = [",
		"abort_actions = [",
		"set_variable_actions = [",
		"snmp_notifications = [",
		`vertex_name = "vertex-1"`,
		`variable_value = "ok-value"`,
		`title = "My Report"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered output missing %q\n--- output ---\n%s", want, out)
		}
	}
}

// TestActionsEmptyOmitted confirms that an `actions` object whose sub-lists
// are all empty (as returned by the UAC API for tasks with no configured
// actions), or a data map that has no "actions" key at all, does not
// produce an `actions = {` stub in the generated HCL.
func TestActionsEmptyOmitted(t *testing.T) {
	t.Run("empty sub-lists", func(t *testing.T) {
		data := minimalTaskData("stonebranch_task_unix", "task_unix_001")
		data["actions"] = map[string]interface{}{
			"systemOperations":   []interface{}{},
			"emailNotifications": []interface{}{},
			"abortActions":       []interface{}{},
			"setVariableActions": []interface{}{},
			"snmpNotifications":  []interface{}{},
		}

		out := renderTemplate(t, "task_unix", data)
		if strings.Contains(out, "actions = {") {
			t.Errorf("expected no actions block for all-empty sub-lists, got:\n%s", out)
		}
	})

	t.Run("actions key absent", func(t *testing.T) {
		data := minimalTaskData("stonebranch_task_unix", "task_unix_001")

		out := renderTemplate(t, "task_unix", data)
		if strings.Contains(out, "actions = {") {
			t.Errorf("expected no actions block when actions key is absent, got:\n%s", out)
		}
	})
}
