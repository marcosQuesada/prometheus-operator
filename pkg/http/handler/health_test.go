package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestItReturnsCommitVersionAndDateOnHealthzHandlerRequest(t *testing.T) {
	v := "fake-version"
	date := "2022-05-03"
	ch := NewChecker(v, date)
	req, _ := http.NewRequest("GET", "healthz", nil)

	w := httptest.NewRecorder()
	ch.healthHandler(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code response, expected http.StatusOK, got %d", w.Result().StatusCode)
	}
	defer w.Result().Body.Close()
	raw, err := ioutil.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatalf("unable to read body, error %v", err)
	}

	res := map[string]string{}
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatalf("unable to unmarshall, error %v", err)
	}

	resv, ok := res["version"]
	if !ok {
		t.Fatal("unable to find response version")
	}
	resd, ok := res["date"]
	if !ok {
		t.Fatal("unable to find response date")
	}

	if expected, got := v, resv; expected != got {
		t.Errorf("version do not match, expected %s got %s", expected, got)
	}
	if expected, got := date, resd; expected != got {
		t.Errorf("date do not match, expected %s got %s", expected, got)
	}
}
