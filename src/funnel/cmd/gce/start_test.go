package gce

import (
	"funnel/config"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func loadTestData(name string) []byte {
	b, err := ioutil.ReadFile("testdata/" + name + ".json")
	if err != nil {
		panic(err)
	}
	return b
}

func testServer(f http.HandlerFunc) *httptest.Server {
	// Start test server
	lis, err := net.Listen("tcp", ":20002")
	if err != nil {
		panic(err)
	}
	// Set up test server response
	mux := http.NewServeMux()
	mux.HandleFunc("/computeMetadata/v1/", f)
	ts := httptest.NewUnstartedServer(mux)
	ts.Listener = lis
	ts.Start()
	return ts
}

// Tests that the code can correctly get metadata from
// a GCE metadata server and merge it with a config.Config instance.
func TestGetMetadata(t *testing.T) {
	ts := testServer(func(w http.ResponseWriter, r *http.Request) {
		if v, ok := r.URL.Query()["recursive"]; !ok || v[0] != "true" {
			t.Fatal("Expected recursive query")
		}
		w.Write(loadTestData("metadata1"))
	})
	defer ts.Close()

	var cerr error
	conf := config.DefaultConfig()
	conf, cerr = WithGCEConfig(conf, "http://localhost:20002")
	if cerr != nil {
		t.Fatal(cerr)
	}

	if conf.Backends.GCE.Project != "funnel-proj-id" {
		t.Fatal("Unexpected project id")
	}

	if conf.Scheduler != "gce" {
		t.Fatal("Unexpected scheduler")
	}

	// When meta.instance.attributes.funnelWorker != ""
	// conf.Worker.ID == meta.instance.name
	if conf.Worker.ID != "funnel-worker-1492486244" {
		t.Fatal("Unexpected worker ID")
	}

	if conf.Worker.Metadata["gce"] != "yes" {
		t.Fatal("Expected gce tag in worker metadata")
	}

	if conf.Backends.GCE.Zone != "us-west1-b" {
		t.Fatal("Unexpected zone")
	}
}