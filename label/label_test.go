package centitylabel

import "testing"

func TestNodeLabelCauseStatus(t *testing.T) {
	got := NodeLabel("cause_status", `{"status_id":3350}`)
	want := "cause_status:3350"
	if got != want {
		t.Fatalf("NodeLabel = %q, want %q", got, want)
	}
}

func TestNodeLabelReceivedMedicalProcedure(t *testing.T) {
	got := NodeLabel("received_medical_procedure", `{"medical_procedure":{"kind":"Transplant","transplant":{"target":{"name":"간"}}}}`)
	want := "received_medical_procedure:간 이식 받음"
	if got != want {
		t.Fatalf("NodeLabel = %q, want %q", got, want)
	}
}

func TestNodeLabelReceivedMedicalProcedureWithPeriod(t *testing.T) {
	got := NodeLabel("received_medical_procedure", `{"medical_procedure":{"kind":"Transplant","transplant":{"target":{"name":"간"}}},"post_procedure_period":"수일에서 수주 이내"}`)
	want := "received_medical_procedure:간 이식 받음 (수일에서 수주 이내)"
	if got != want {
		t.Fatalf("NodeLabel = %q, want %q", got, want)
	}
}

func TestStatusLabelIndex(t *testing.T) {
	tests := []struct {
		kind string
		data string
		want string
	}{
		{"decreased_index", `{"index":{"name":"담즙 배액량"}}`, "담즙 배액량 감소"},
		{"increased_index", `{"index":{"name":"AST"}}`, "AST 증가"},
		{"decreased_index", `{"index":"index(담즙 배액량)"}`, "담즙 배액량 감소"},
		{"decreased_index", `{"index":{"name":"소변량 감소"}}`, "소변량 감소"},
	}
	for _, test := range tests {
		if got := StatusLabel(test.kind, test.data); got != test.want {
			t.Fatalf("StatusLabel(%q) = %q, want %q", test.kind, got, test.want)
		}
	}
}

func TestStatusLabelReceivedMedicalProcedure(t *testing.T) {
	got := StatusLabel("received_medical_procedure", `{"medical_procedure":{"kind":"Infuse","infuse":{"material":{"kind":"other_actor","name":"고용량 스테로이드 정맥 투여"},"route":"iv"}}}`)
	want := "고용량 스테로이드 정맥 투여 받음"
	if got != want {
		t.Fatalf("StatusLabel = %q, want %q", got, want)
	}
}

func TestStatusLabelProcedureWithoutTechnicalPrefix(t *testing.T) {
	tests := []struct {
		data string
		want string
	}{
		{`{"medical_procedure":{"kind":"DiagnosticTest","diagnostic_test":{"name":"소변배양검사"}}}`, "소변배양검사 받음"},
		{`{"medical_procedure":{"kind":"Unknown","unknown":{"name":"체중 측정"}}}`, "체중 측정 받음"},
		{`{"medical_procedure":{"kind":"Cleaning","cleaning":{"target":{"kind":"unknown","name":"S상결장루"}}}}`, "세척 (S상결장루) 받음"},
	}
	for _, test := range tests {
		if got := StatusLabel("received_medical_procedure", test.data); got != test.want {
			t.Fatalf("StatusLabel = %q, want %q", got, test.want)
		}
	}
}

func TestStatusLabelPain(t *testing.T) {
	got := StatusLabel("have_symptom", `{"symptom":{"kind":"pain","aspects":[{"kind":"site","site":{"kind":"organ","name":"우상복부"}},{"kind":"quality","name":"날카로운"}]}}`)
	want := "우상복부 통증 (날카로운)"
	if got != want {
		t.Fatalf("StatusLabel = %q, want %q", got, want)
	}
}

func TestStatusLabelCauseStatusFallback(t *testing.T) {
	got := StatusLabel("cause_status", `{"statusKey":"s3"}`)
	if got != "원인 상태 정보 없음" {
		t.Fatalf("StatusLabel = %q, want 원인 상태 정보 없음", got)
	}
}
