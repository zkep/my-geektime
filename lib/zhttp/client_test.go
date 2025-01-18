package zhttp_test

import (
	"testing"

	"github.com/zkep/mygeektime/lib/zhttp"
)

func TestClient(t *testing.T) {
	err := zhttp.NewRequest().Do("GET", "https://github.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
}
