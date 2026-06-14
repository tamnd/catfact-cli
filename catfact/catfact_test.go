package catfact_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/catfact-cli/catfact"
)

const fakeFactsJSON = `{"current_page":1,"data":[{"fact":"Unlike dogs, cats do not have a sweet tooth.","length":44},{"fact":"Cats sleep 12-16 hours per day.","length":31}],"total":332}`

const fakeFactJSON = `{"fact":"Cats have 32 muscles in each ear.","length":34}`

func newTestClient(ts *httptest.Server) *catfact.Client {
	cfg := catfact.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return catfact.NewClient(cfg)
}

func TestFactsSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeFactsJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Facts(context.Background(), 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

func TestFactsParsesItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeFactsJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Facts(context.Background(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Rank != 1 {
		t.Errorf("items[0].Rank = %d, want 1", items[0].Rank)
	}
	if items[0].Fact != "Unlike dogs, cats do not have a sweet tooth." {
		t.Errorf("items[0].Fact = %q", items[0].Fact)
	}
	if items[0].Length != 44 {
		t.Errorf("items[0].Length = %d, want 44", items[0].Length)
	}
	if items[1].Rank != 2 {
		t.Errorf("items[1].Rank = %d, want 2", items[1].Rank)
	}
}

func TestFactsPageOffsetRank(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify page param forwarded
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("page query param = %q, want 2", r.URL.Query().Get("page"))
		}
		_, _ = fmt.Fprint(w, fakeFactsJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Facts(context.Background(), 10, 2)
	if err != nil {
		t.Fatal(err)
	}
	// page 2 with limit 10 → ranks 11, 12
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Rank != 11 {
		t.Errorf("items[0].Rank = %d, want 11", items[0].Rank)
	}
	if items[1].Rank != 12 {
		t.Errorf("items[1].Rank = %d, want 12", items[1].Rank)
	}
}

func TestFactsRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeFactsJSON)
	}))
	defer ts.Close()

	cfg := catfact.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := catfact.NewClient(cfg)

	_, err := c.Facts(context.Background(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

func TestFactParsesOne(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeFactJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	item, err := c.Fact(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if item.Rank != 1 {
		t.Errorf("item.Rank = %d, want 1", item.Rank)
	}
	if item.Fact != "Cats have 32 muscles in each ear." {
		t.Errorf("item.Fact = %q", item.Fact)
	}
	if item.Length != 34 {
		t.Errorf("item.Length = %d, want 34", item.Length)
	}
}
