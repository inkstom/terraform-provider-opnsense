// Package mockopnsense is a small in-memory imitation of the OPNsense MVC API used to exercise
// provider CRUD/import/plan-stability logic without a real firewall. It models the generic
// ApiMutableModelControllerBase behaviour (add/get/set/del + reconfigure) and the response
// shapes OPNsense uses (option fields as {key:{value,selected}} maps, booleans as "0"/"1").
//
// It does NOT reproduce OPNsense's validation or side effects; real acceptance tests are
// still required against a live instance.
package mockopnsense

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
)

// Kind describes how a field is rendered on GET.
type Kind int

const (
	Plain  Kind = iota // "value"
	Select             // {"k": {"value": "k", "selected": 1}, ...}
	Multi              // as Select, several keys may be selected; written as "a,b"
)

// Model describes one MVC model as served by a controller.
type Model struct {
	Monad    string
	Fields   map[string]Kind
	Defaults map[string]string
	// Options are extra unselected keys rendered for Select fields (mimics real option lists).
	Options map[string][]string
	// NewID returns the identifier for a new item; defaults to a random UUID.
	NewID func(s *store) string
	// OnAdd lets a model compute server-side fields (e.g. auto numbers).
	OnAdd func(s *store, id string, item map[string]string)
	// Volatile fields are accepted on write but never stored.
	Volatile []string
}

type store struct {
	model Model
	items map[string]map[string]string
	seq   int
}

// Route binds endpoints to a model.
type Route struct {
	Add, Get, Set, Del, Reconfigure string
	Model                           Model
}

// Server is a running mock.
type Server struct {
	*httptest.Server
	mu     sync.Mutex
	stores map[string]*store // keyed by Route.Get
	routes []Route
	Calls  []string
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func randKey() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// Start launches the mock and points the provider at it via environment variables.
func Start(t *testing.T, routes ...Route) *Server {
	t.Helper()
	s := &Server{stores: map[string]*store{}, routes: routes}
	for _, r := range routes {
		s.stores[r.Get] = &store{model: r.Model, items: map[string]map[string]string{}}
	}
	s.Server = httptest.NewTLSServer(http.HandlerFunc(s.handle))
	t.Cleanup(s.Close)

	t.Setenv("OPNSENSE_URI", s.URL)
	t.Setenv("OPNSENSE_API_KEY", "key")
	t.Setenv("OPNSENSE_API_SECRET", "secret")
	t.Setenv("OPNSENSE_ALLOW_INSECURE", "true")
	return s
}

// Items returns a copy of the stored items for the route whose get endpoint is getEndpoint.
func (s *Server) Items(getEndpoint string) map[string]map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := map[string]map[string]string{}
	for id, it := range s.stores[getEndpoint].items {
		cp := map[string]string{}
		for k, v := range it {
			cp[k] = v
		}
		out[id] = cp
	}
	return out
}

// Seed inserts an item directly (e.g. pre-existing wan/lan assignments).
func (s *Server) Seed(getEndpoint, id string, item map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.stores[getEndpoint]
	full := map[string]string{}
	for k, v := range st.model.Defaults {
		full[k] = v
	}
	for k, v := range item {
		full[k] = v
	}
	st.items[id] = full
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := strings.TrimPrefix(r.URL.Path, "/api")
	s.Calls = append(s.Calls, r.Method+" "+p)

	switch p {
	case "/wireguard/server/keyPair":
		writeJSON(w, map[string]string{"status": "ok", "privkey": randKey(), "pubkey": randKey()})
		return
	case "/wireguard/client/psk":
		writeJSON(w, map[string]string{"status": "ok", "psk": randKey()})
		return
	}

	for _, rt := range s.routes {
		st := s.stores[rt.Get]
		switch {
		case p == rt.Reconfigure:
			writeJSON(w, map[string]string{"status": "ok"})
			return
		case p == rt.Add && r.Method == http.MethodPost:
			fields, err := decode(r, st.model.Monad)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			id := newUUID()
			if st.model.NewID != nil {
				id = st.model.NewID(st)
			}
			item := map[string]string{}
			for k, v := range st.model.Defaults {
				item[k] = v
			}
			st.merge(item, fields)
			if st.model.OnAdd != nil {
				st.model.OnAdd(st, id, item)
			}
			st.items[id] = item
			writeJSON(w, map[string]string{"result": "saved", "uuid": id})
			return
		case strings.HasPrefix(p, rt.Set+"/") && r.Method == http.MethodPost:
			id := strings.TrimPrefix(p, rt.Set+"/")
			item, ok := st.items[id]
			if !ok {
				writeJSON(w, map[string]any{"result": "failed", "validations": map[string]string{"id": "not found"}})
				return
			}
			fields, err := decode(r, st.model.Monad)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			st.merge(item, fields)
			writeJSON(w, map[string]string{"result": "saved"})
			return
		case strings.HasPrefix(p, rt.Del+"/") && r.Method == http.MethodPost:
			id := strings.TrimPrefix(p, rt.Del+"/")
			if _, ok := st.items[id]; !ok {
				writeJSON(w, map[string]string{"result": "not found"})
				return
			}
			delete(st.items, id)
			writeJSON(w, map[string]string{"result": "deleted"})
			return
		case strings.HasPrefix(p, rt.Get+"/") && r.Method == http.MethodGet:
			id := strings.TrimPrefix(p, rt.Get+"/")
			item, ok := st.items[id]
			if !ok {
				// OPNsense returns an empty array for unknown items.
				writeJSON(w, []any{})
				return
			}
			writeJSON(w, map[string]any{st.model.Monad: st.render(item)})
			return
		}
	}
	http.NotFound(w, r)
}

func decode(r *http.Request, monad string) (map[string]string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var wrapped map[string]map[string]any
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, fmt.Errorf("bad body %q: %w", body, err)
	}
	raw, ok := wrapped[monad]
	if !ok {
		return nil, fmt.Errorf("missing monad %q in %q", monad, body)
	}
	out := map[string]string{}
	for k, v := range raw {
		switch tv := v.(type) {
		case string:
			out[k] = tv
		case nil:
			out[k] = ""
		default:
			b, _ := json.Marshal(tv)
			out[k] = string(b)
		}
	}
	return out, nil
}

func (st *store) merge(item, fields map[string]string) {
	volatile := map[string]bool{}
	for _, v := range st.model.Volatile {
		volatile[v] = true
	}
	for k, v := range fields {
		if _, known := st.model.Fields[k]; !known || volatile[k] {
			continue // OPNsense ignores unknown keys
		}
		if st.model.Fields[k] == Multi {
			parts := strings.Split(v, ",")
			kept := parts[:0]
			for _, p := range parts {
				if p != "" {
					kept = append(kept, p)
				}
			}
			sort.Strings(kept)
			v = strings.Join(kept, ",")
		}
		item[k] = v
	}
}

func (st *store) render(item map[string]string) map[string]any {
	out := map[string]any{}
	for name, kind := range st.model.Fields {
		val := item[name]
		switch kind {
		case Plain:
			out[name] = val
		case Select, Multi:
			selected := map[string]bool{}
			if kind == Multi {
				for _, k := range strings.Split(val, ",") {
					if k != "" {
						selected[k] = true
					}
				}
			} else {
				selected[val] = true
			}
			opts := map[string]any{}
			for _, k := range st.model.Options[name] {
				opts[k] = map[string]any{"value": k, "selected": 0}
			}
			for k := range selected {
				if k == "" && kind == Select {
					// blank option is rendered by OPNsense for optional option fields
					opts[""] = map[string]any{"value": "none", "selected": 1}
					continue
				}
				opts[k] = map[string]any{"value": k, "selected": 1}
			}
			out[name] = opts
		}
	}
	return out
}
