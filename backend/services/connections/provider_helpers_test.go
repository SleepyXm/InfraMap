package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSelectedServiceIDsPreservesClearedSelection(t *testing.T) {
	for _, test := range []struct {
		name, metadata string
		want           map[string]bool
	}{
		{"legacy", `{"projects":[{"ref":"old"}]}`, map[string]bool{"old": true}},
		{"cleared", `{"services":[],"projects":[{"ref":"old"}]}`, map[string]bool{}},
		{"null", `{"services":null,"projects":[{"ref":"old"}]}`, map[string]bool{}},
		{"current", `{"services":[{"ref":"new"}],"projects":[{"ref":"old"}]}`, map[string]bool{"new": true}},
		{"missing", `{}`, map[string]bool{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := selectedServiceIDs[SupabaseService]([]byte(test.metadata))
			if err != nil || !reflect.DeepEqual(got, test.want) {
				t.Fatalf("selected = %v, error = %v; want %v", got, err, test.want)
			}
		})
	}
	if _, err := selectedServiceIDs[SupabaseService]([]byte(`{"services":{}}`)); err == nil {
		t.Fatal("malformed service selection was accepted")
	}
	selected, err := selectedServiceIDs[VercelService]([]byte(`{"projects":[{"id":"legacy"}]}`))
	if err != nil || !selected["legacy"] {
		t.Fatalf("legacy Vercel selection lost: %v, %v", selected, err)
	}
}

func TestSelectServicesValidatesAndDeduplicates(t *testing.T) {
	available := []VercelService{{ID: "one"}, {ID: "two"}}
	selected, err := selectServices(available, []string{"two", "two", "one"})
	if err != nil || len(selected) != 2 || selected[0].ID != "two" || !selected[0].Selected || !selected[1].Selected {
		t.Fatalf("selection = %+v, error = %v", selected, err)
	}
	if available[0].Selected || available[1].Selected {
		t.Fatal("selection mutated provider results")
	}
	if selected, err := selectServices(available, []string{"one", "inaccessible"}); err == nil || selected != nil {
		t.Fatalf("partially accepted inaccessible selection: %+v, %v", selected, err)
	}
	empty, err := selectServices(available, nil)
	body, _ := json.Marshal(empty)
	if err != nil || string(body) != "[]" {
		t.Fatalf("cleared selection must serialize as []: %s, %v", body, err)
	}
}

type providerTestTransport func(*http.Request) (*http.Response, error)

func (transport providerTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func useProviderServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	previous := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: providerTestTransport(func(request *http.Request) (*http.Response, error) {
		copy := request.Clone(request.Context())
		copy.URL.Scheme, copy.URL.Host = target.Scheme, target.Host
		return server.Client().Transport.RoundTrip(copy)
	})}
	t.Cleanup(func() { http.DefaultClient = previous })
}

func TestFetchVercelServicesDecodesProjectsAndPaginates(t *testing.T) {
	calls := 0
	useProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/v9/projects" || r.URL.Query().Get("teamId") != "team-one" || r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected provider request: %s", r.URL)
		}
		if calls == 1 {
			_, _ = w.Write([]byte(`{"projects":[{"id":"one","name":"Frontend"}],"pagination":{"next":123}}`))
			return
		}
		if r.URL.Query().Get("until") != "123" {
			t.Errorf("missing pagination cursor: %s", r.URL)
		}
		_, _ = w.Write([]byte(`{"projects":[{"id":"two","name":"Backend"}],"pagination":{"next":null}}`))
	})
	services, err := FetchVercelServices(context.Background(), "test-token", "team-one")
	if err != nil || calls != 2 || len(services) != 2 || services[0].ID != "one" || services[1].ID != "two" {
		t.Fatalf("services = %+v, calls = %d, error = %v", services, calls, err)
	}
}

func TestFetchVercelServicesDoesNotReturnFalseEmptyOrPartialResults(t *testing.T) {
	for _, test := range []struct {
		name, response string
		wantError      bool
	}{
		{"empty", `{"projects":[],"pagination":{"next":null}}`, false},
		{"wrong-field", `{"services":[{"id":"one"}]}`, true},
		{"truncated", `{"projects":[{"id":"one"}],"pagination":{"next":123}}`, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			useProviderServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(test.response)) })
			services, err := FetchVercelServices(context.Background(), "test-token", "")
			if (err != nil) != test.wantError || (test.wantError && services != nil) {
				t.Fatalf("services = %+v, error = %v", services, err)
			}
		})
	}
}

func TestProviderRequestReportsErrorsAndCancellation(t *testing.T) {
	for _, body := range []string{
		`{"message":"permission denied"}`, `{"error":"permission denied"}`,
		`{"error":{"message":"permission denied"}}`, "not json",
	} {
		t.Run(body, func(t *testing.T) {
			useProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(body))
			})
			var result any
			err := providerRequest(context.Background(), http.MethodGet, "https://api.vercel.com/test", "test-token", nil, &result, time.Second)
			if err == nil || !strings.Contains(err.Error(), "403") {
				t.Fatalf("provider error was lost: %v", err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var result any
	err := providerRequest(ctx, http.MethodGet, "https://api.vercel.com/test", "", nil, &result, time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation was lost: %v", err)
	}
}

func TestProviderRequestSendsJSONAndRejectsMalformedSuccess(t *testing.T) {
	useProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Accept") != "application/json" {
			t.Error("missing JSON request headers or method")
		}
		var input map[string]string
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input["name"] != "service" {
			t.Errorf("invalid request body: %v, %v", input, err)
		}
		_, _ = w.Write([]byte("invalid json"))
	})
	var result any
	err := providerRequest(context.Background(), http.MethodPost, "https://api.supabase.com/test", "test-token", map[string]string{"name": "service"}, &result, time.Second)
	if err == nil {
		t.Fatal("invalid success response was accepted")
	}
}

func TestFetchVercelAccountName(t *testing.T) {
	for _, test := range []struct{ name, team, body, want string }{
		{"personal", "", `{"user":{"username":"dave"}}`, "dave"},
		{"team", "team-one", `{"name":"InfraMap","slug":"inframap"}`, "InfraMap"},
	} {
		t.Run(test.name, func(t *testing.T) {
			useProviderServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(test.body)) })
			if got := FetchVercelAccountName(context.Background(), "test-token", test.team); got != test.want {
				t.Fatalf("name = %q, want %q", got, test.want)
			}
		})
	}
}
