package csolve

import "testing"

func TestLooksLikePatientDescription(t *testing.T) {
	cases := []struct {
		problem string
		want    bool
	}{
		{"방광경검사를 받은 환자의 진술 중 우선적인 간호 중재가 필요한 것은?", true},
		{"퇴원 교육 내용을 잘 이해한 표현은?", true},
		{"환자가 '소변에 핏덩이가 보여요.'라고 말했다.", true},
		{"급성 심근경색증으로 의심되는 환자가 응급실에 왔다.", false},
	}
	for _, item := range cases {
		if got := looks_like_patient_description(item.problem); got != item.want {
			t.Fatalf("looks_like_patient_description(%q) = %v, want %v", item.problem, got, item.want)
		}
	}
}
