package auth

import (
	"context"

	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOPAFailsClosed(t *testing.T) {
	for _, test := range []struct {
		body   string
		status int
		allow  bool
	}{{`{"result":true}`, 200, true}, {`{"result":false}`, 200, false}, {`{}`, 200, false}, {`{"result":"true"}`, 200, false}, {`oops`, 200, false}, {`{"result":true}`, 500, false}} {
		t.Run(test.body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Error(r.Method)
				}
				w.WriteHeader(test.status)
				w.Write([]byte(test.body))
			}))
			defer server.Close()
			err := (&OPAClient{URL: server.URL}).Authorize(context.Background(), Decision{Path: "Query.customer"})
			if (err == nil) != test.allow {
				t.Fatalf("err=%v allow=%v", err, test.allow)
			}
		})
	}
}
