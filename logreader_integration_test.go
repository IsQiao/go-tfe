//go:build integration
// +build integration

package tfe

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// checkedWrite writes message to w and fails the test if there's an error.
func checkedWrite(t *testing.T, w io.Writer, message []byte) {
	_, err := w.Write(message)
	if err != nil {
		t.Fatalf("error writing response: %s", err)
	}
}

func testLogReader(t *testing.T, h http.HandlerFunc) (*httptest.Server, *LogReader) {
	skipIfNotCINode(t)

	ts := httptest.NewServer(h)

	cfg := &Config{
		Address:    ts.URL,
		Token:      "dummy-token",
		HTTPClient: ts.Client(),
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	logURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	lr := &LogReader{
		client: client,
		ctx:    context.Background(),
		logURL: logURL,
	}

	return ts, lr
}

func TestLogReader_withMarkersSingle(t *testing.T) {
	skipIfNotCINode(t)
	t.Parallel()

	logReads := 0
	ts, lr := testLogReader(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logReads++
		switch {
		case logReads == 2:
			checkedWrite(t, w, []byte("\x02Terraform run started - logs - Terraform run finished\x03"))
		}
	}))
	defer ts.Close()

	doneReads := 0
	lr.done = func() (bool, error) {
		doneReads++
		if logReads >= 2 {
			return true, nil
		}
		return false, nil
	}

	logs, err := io.ReadAll(lr)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Terraform run started - logs - Terraform run finished"
	if string(logs) != expected {
		t.Fatalf("expected %s, got: %s", expected, string(logs))
	}
	if doneReads != 1 {
		t.Fatalf("expected 1 done reads, got %d reads", doneReads)
	}
	if logReads != 3 {
		t.Fatalf("expected 3 log reads, got %d reads", logReads)
	}
}

func TestLogReader_withMarkersDouble(t *testing.T) {
	skipIfNotCINode(t)
	t.Parallel()

	logReads := 0
	ts, lr := testLogReader(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logReads++
		switch {
		case logReads == 2:
			checkedWrite(t, w, []byte("\x02Terraform run started"))
		case logReads == 3:
			checkedWrite(t, w, []byte(" - logs - Terraform run finished\x03"))
		}
	}))
	defer ts.Close()

	doneReads := 0
	lr.done = func() (bool, error) {
		doneReads++
		if logReads >= 3 {
			return true, nil
		}
		return false, nil
	}

	logs, err := io.ReadAll(lr)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Terraform run started - logs - Terraform run finished"
	if string(logs) != expected {
		t.Fatalf("expected %s, got: %s", expected, string(logs))
	}
	if doneReads != 1 {
		t.Fatalf("expected 1 done reads, got %d reads", doneReads)
	}
	if logReads != 4 {
		t.Fatalf("expected 4 log reads, got %d reads", logReads)
	}
}

func TestLogReader_withMarkersMulti(t *testing.T) {
	skipIfNotCINode(t)
	t.Parallel()

	logReads := 0
	ts, lr := testLogReader(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logReads++
		switch {
		case logReads == 2:
			checkedWrite(t, w, []byte("\x02"))
		case logReads == 3:
			checkedWrite(t, w, []byte("Terraform run started"))
		case logReads == 16:
			checkedWrite(t, w, []byte(" - logs - "))
		case logReads == 30:
			checkedWrite(t, w, []byte("Terraform run finished"))
		case logReads == 31:
			checkedWrite(t, w, []byte("\x03"))
		}
	}))
	defer ts.Close()

	doneReads := 0
	lr.done = func() (bool, error) {
		doneReads++
		if logReads >= 31 {
			return true, nil
		}
		return false, nil
	}

	logs, err := io.ReadAll(lr)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Terraform run started - logs - Terraform run finished"
	if string(logs) != expected {
		t.Fatalf("expected %s, got: %s", expected, string(logs))
	}
	if doneReads != 3 {
		t.Fatalf("expected 3 done reads, got %d reads", doneReads)
	}
	if logReads != 31 {
		t.Fatalf("expected 31 log reads, got %d reads", logReads)
	}
}

func TestLogReader_withoutMarkers(t *testing.T) {
	skipIfNotCINode(t)
	t.Parallel()

	logReads := 0
	ts, lr := testLogReader(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logReads++
		switch {
		case logReads == 2:
			checkedWrite(t, w, []byte("Terraform run started"))
		case logReads == 16:
			checkedWrite(t, w, []byte(" - logs - "))
		case logReads == 31:
			checkedWrite(t, w, []byte("Terraform run finished"))
		}
	}))
	defer ts.Close()

	doneReads := 0
	lr.done = func() (bool, error) {
		doneReads++
		if logReads >= 31 {
			return true, nil
		}
		return false, nil
	}

	logs, err := io.ReadAll(lr)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Terraform run started - logs - Terraform run finished"
	if string(logs) != expected {
		t.Fatalf("expected %s, got: %s", expected, string(logs))
	}
	if doneReads != 25 {
		t.Fatalf("expected 14 done reads, got %d reads", doneReads)
	}
	if logReads != 32 {
		t.Fatalf("expected 32 log reads, got %d reads", logReads)
	}
}

func TestLogReader_withoutEndOfTextMarker(t *testing.T) {
	skipIfNotCINode(t)
	t.Parallel()

	logReads := 0
	ts, lr := testLogReader(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logReads++
		switch {
		case logReads == 2:
			checkedWrite(t, w, []byte("\x02"))
		case logReads == 3:
			checkedWrite(t, w, []byte("Terraform run started"))
		case logReads == 16:
			checkedWrite(t, w, []byte(" - logs - "))
		case logReads == 31:
			checkedWrite(t, w, []byte("Terraform run finished"))
		}
	}))
	defer ts.Close()

	doneReads := 0
	lr.done = func() (bool, error) {
		doneReads++
		if logReads >= 31 {
			return true, nil
		}
		return false, nil
	}

	logs, err := io.ReadAll(lr)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Terraform run started - logs - Terraform run finished"
	if string(logs) != expected {
		t.Fatalf("expected %s, got: %s", expected, string(logs))
	}
	if doneReads != 3 {
		t.Fatalf("expected 3 done reads, got %d reads", doneReads)
	}
	if logReads != 42 {
		t.Fatalf("expected 42 log reads, got %d reads", logReads)
	}
}
