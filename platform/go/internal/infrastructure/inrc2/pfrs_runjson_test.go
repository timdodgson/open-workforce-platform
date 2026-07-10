package inrc2

import "testing"

func TestPFRSStandardRunJSONFormat(t *testing.T) {
	got := formatPFRSStandardRunJSON(PFRSStandardRunJSONParams{
		InstanceName: "n012w8", WorkerMode: "sa", BestPenalty: 3465, RunLabel: "my-run",
	})
	want := `{
  "instance": "n012w8",
  "problemType": "nrp",
  "mode": "sa",
  "bestObjective": 3465,
  "totalPenalty": 3465,
  "runLabel": "my-run"
}`
	if got != want {
		t.Fatalf("PFRS standard run.json format changed:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

