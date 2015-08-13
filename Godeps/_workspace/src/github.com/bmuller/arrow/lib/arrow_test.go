package arrow

import (
	"testing"
	"time"
)

func TestCParse(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	a, _ := CParseInLocation("%Y-%m-%d", "1980-06-19", loc)
	if a.Format("2006-01-02") != "1980-06-19" {
		t.Error("CParseInLocation")
	}
}
