package scripts

import "testing"

func TestExtractVariableNames(t *testing.T) {
	names := ExtractVariableNames(
		"const host = '{{ host }}'; const port = '{{port}}';",
		"again {{host}} and {{user_id}} {{1bad}} {{bad-name}}",
	)
	want := []string{"host", "port", "user_id"}
	if len(names) != len(want) {
		t.Fatalf("names = %#v", names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("names = %#v", names)
		}
	}
}

func TestNormalizeVariables(t *testing.T) {
	got := NormalizeVariables([]Variable{
		{Name: "host", Value: "a"},
		{Name: "host", Value: "b"},
		{Name: "bad-name", Value: "x"},
		{Name: "1port", Value: "x"},
		{Name: "port", Value: "80"},
	})
	if len(got) != 2 || got[0].Name != "host" || got[0].Value != "a" || got[1].Name != "port" || got[1].Value != "80" {
		t.Fatalf("got = %#v", got)
	}
	if NormalizeVariables(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestExtractInlineDefaults(t *testing.T) {
	names := ExtractVariableNames(
		`{{ host = local host }} {{port=80}} {{host=ignored}} {{user_id}} {{name=}}`,
	)
	want := []string{"host", "port", "user_id", "name"}
	if len(names) != len(want) {
		t.Fatalf("names = %#v", names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("names = %#v", names)
		}
	}
}

func TestApplyInlineDefault(t *testing.T) {
	got := ApplyVariables(`{{host=localhost}} {{ port = 80 }} {{host}} {{missing}}`, map[string]string{
		"port": "saved",
	}, map[string]string{
		"missing": "over",
	})
	want := `localhost saved localhost over`
	if got != want {
		t.Fatalf("got %q", got)
	}
	cleared := ApplyVariables(`{{host=localhost}}`, map[string]string{"host": ""}, nil)
	if cleared != "" {
		t.Fatalf("form empty should win, got %q", cleared)
	}
	spaced := ApplyVariables(`{{ host = local host }}`, nil, nil)
	if spaced != "local host" {
		t.Fatalf("spaced = %q", spaced)
	}
}

func TestApplyVariables(t *testing.T) {
	content := `print("{{host}}", "{{ port }}", "{{missing}}")`
	got := ApplyVariables(content, map[string]string{
		"host":    "localhost",
		"port":    "80",
		"missing": "keep",
	}, map[string]string{
		"port": "",
	})
	want := `print("localhost", "", "keep")`
	if got != want {
		t.Fatalf("got %q", got)
	}

	cleared := ApplyVariables("{{host}}", nil, nil)
	if cleared != "" {
		t.Fatalf("cleared = %q", cleared)
	}
}

func TestVariablesRoundTrip(t *testing.T) {
	s := newTestService(t)
	created, err := s.Create(CreateInput{
		Name:     "vars",
		Language: "javascript",
		Content:  "const host = '{{host}}';",
		Variables: []Variable{
			{Name: "host", Value: "localhost"},
			{Name: "bad-name", Value: "nope"},
			{Name: "host", Value: "ignored"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(created.Variables) != 1 || created.Variables[0].Name != "host" || created.Variables[0].Value != "localhost" {
		t.Fatalf("created variables = %#v", created.Variables)
	}
	updated, err := s.Update(created.ID, UpdateInput{
		Name:     "vars",
		Language: "javascript",
		Content:  "const host = '{{host}}';",
		Variables: []Variable{
			{Name: "port", Value: "443"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Variables) != 1 || updated.Variables[0].Name != "port" || updated.Variables[0].Value != "443" {
		t.Fatalf("updated variables = %#v", updated.Variables)
	}
	items, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || len(items[0].Variables) != 1 || items[0].Variables[0].Value != "443" {
		t.Fatalf("list variables = %#v", items[0].Variables)
	}
}
