package ctlstatus

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/appengine/aetest"
)

func Test_dummy(t *testing.T) {
	if false {
		t.Error("the sky is falling")
	}
}

func Test_root(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("couldn't create instance: %v", err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to make a GET request: %v", err)
	}

	response := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		t.Fatalf("response did not contain expected 200. instead: %v", response.Code)
	}

}
