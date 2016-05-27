package ctlstatus

import (
	"fmt"
	"testing"
)

func Test_Path(t *testing.T) {
	i := Incident{Id: 666}
	expected := "/incident/666/"
	r := i.Path()
	if r != expected {
		t.Errorf("expected '%v', got '%v'", expected, r)
	}
}

func Test_StatusOptions(t *testing.T) {
	var table = []struct {
		in  string
		out []string
	}{
		{in: "investigating", out: []string{"outage", "resolved"}},
		{in: "outage", out: []string{"resolved"}},
		{in: "resolved", out: []string{"investigating", "outage"}},
	}
	for _, tt := range table {
		i := Incident{Status: tt.in}
		// quick and dirty approach to comparing slices
		r := fmt.Sprintf("%v", i.StatusOptions())
		expected := fmt.Sprintf("%v", tt.out)
		if r != expected {
			t.Errorf("expected '%v', got '%v'", expected, r)
		}
	}
}

func Test_BootstrapClass(t *testing.T) {
	var table = []struct {
		in  string
		out string
	}{
		{"investigating", "warning"},
		{"outage", "danger"},
		{"resolved", "success"},
	}
	for _, tt := range table {
		i := Incident{Status: tt.in}
		r := i.BootstrapClass()
		expected := tt.out
		if r != expected {
			t.Errorf("expected '%v', got '%v'", expected, r)
		}
	}
}
