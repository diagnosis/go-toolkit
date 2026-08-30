package mailer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/diagnosis/go-toolkit/v2/apperr"
)

func TestResendMailer_SendsAuthAndPayload(t *testing.T) {
	var gotAuth string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	m := NewResendMailer("test-key", "muster@safadev.app")
	m.baseURL = srv.URL

	err := m.Send(context.Background(), []string{"a@b.com"}, "subj", "<b>hi</b>")
	if err != nil {
		t.Fatalf("expected no error got error %v", err)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf(`expected "Bearer test-key" got %v`, gotAuth)
	}
	if gotBody["from"] != "muster@safadev.app" {
		t.Errorf(`expected "muster@safadev.app" got %v`, gotBody["from"])
	}
	if gotBody["html"] != "<b>hi</b>" {
		t.Errorf(`expected "<b>hi</b>" got %v`, gotBody["html"])
	}
	if gotBody["subject"] != "subj" {
		t.Errorf(`expected "subj" got %v`, gotBody["subject"])
	}

	toSlice, ok := gotBody["to"].([]any)
	if ok {
		if len(toSlice) != 1 {
			t.Errorf("expected slice len 1 got %d", len(toSlice))
		}
		if toSlice[0] != "a@b.com" {
			t.Errorf(`expected "a@b.com" got %v`, toSlice[0])
		}
	} else {
		t.Errorf("expected to receive slice of receiver")
	}

}

func TestNewResendMailer_422Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		_, _ = w.Write([]byte(`{"message":"invalid from"}`))
	}))
	defer srv.Close()

	m := NewResendMailer("test-key", "muster@safadev.app")
	m.baseURL = srv.URL

	err := m.Send(context.Background(), []string{"safa@safa.com"}, "subj", "<p>hello</p>")
	if err == nil {
		t.Fatal("expected error but got no error")
	}

	if se, ok := apperr.AsStatusErr(err); ok {
		if se.HTTPStatus != 500 {
			t.Errorf("expected %d got %d", 500, se.HTTPStatus)
		}
		if se.Message != "email delivery failed" {
			t.Errorf("expected %s got %s", "email delivery failed", se.Message)
		}

		if se.InternalMessage != "resend status=422: invalid from" {
			t.Errorf("expected resend status=422 got %s", se.InternalMessage)
		}
	} else {
		t.Error("expected to receive status error")
	}

}

func TestResendMailer_SendWithContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		_, _ = w.Write([]byte(`{"message":"invalid from"}`))
	}))
	defer srv.Close()

	m := NewResendMailer("test-key", "muster@safadev.app")
	m.baseURL = srv.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := m.Send(ctx, []string{"safa@safa.com"}, "subj", "<p>hello</p>")
	if err == nil {
		t.Fatal("expected error got nil")
	}
	if se, ok := apperr.AsStatusErr(err); ok {
		if se.HTTPStatus != 500 {
			t.Errorf("expected 500 got %d", se.HTTPStatus)
		}
		if se.Message != "failed to deliver email" {
			t.Errorf("expected failed to deliver email got %s", se.Message)
		}
		if se.InternalMessage != "http client request failed" {
			t.Errorf("expected http client request failed got %s", se.InternalMessage)
		}
		errText := se.Error()
		if !strings.Contains(errText, "context canceled") {
			t.Error("expected to contain context failed")
		}
	} else {
		t.Error("expected to receive status error")
	}
}

func TestResendMailer_DoubleSend(t *testing.T) {

	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(200)
	}))
	defer srv.Close()
	m := NewResendMailer("test-key", "muster@safadev.app")
	m.baseURL = srv.URL
	err := m.Send(context.Background(), []string{"a@b.com"}, "subj", "<b>hi</b>")
	if err != nil {
		t.Fatalf("expected to deliver got %v", err)
	}
	err = m.Send(context.Background(), []string{"a@a.com"}, "subj23", "<b>hello</b>")
	if err != nil {
		t.Fatalf("expected to deliver got %v", err)
	}
	if hits != 2 {
		t.Errorf("expected req counts 2 got %d", hits)
	}
}
