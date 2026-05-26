package krx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bocastle/krx-pension-bot/internal/report"
)

func TestClientFetchesAndParsesMarketFlows(t *testing.T) {
	var formBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/comm/bldAttendant/getJsonData.cmd" {
			t.Fatalf("path = %s, want KRX JSON endpoint", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.Form.Get("mktId"); got != "STK" {
			t.Fatalf("mktId = %q, want STK", got)
		}
		if got := r.Form.Get("invstTpCd"); got != "6000" {
			t.Fatalf("invstTpCd = %q, want 6000", got)
		}
		formBody = r.Form.Encode()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"output": [
				{
					"ISU_SRT_CD": "005930",
					"ISU_ABBRV": "삼성전자",
					"ASK_TRDVOL": "1,000",
					"BID_TRDVOL": "2,000",
					"NETBID_TRDVOL": "1,000",
					"ASK_TRDVAL": "7,700,000,000",
					"BID_TRDVAL": "20,000,000,000",
					"NETBID_TRDVAL": "12,300,000,000"
				}
			]
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 10*time.Minute)
	rows, err := client.MarketFlows(context.Background(), report.MarketKOSPI, report.QueryPeriod{
		Start: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("MarketFlows() error = %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(rows))
	}
	if rows[0].Code != "005930" || rows[0].Name != "삼성전자" || rows[0].NetValue != 12_300_000_000 {
		t.Fatalf("row = %#v", rows[0])
	}
	if !strings.Contains(formBody, "bld=dbms%2FMDC_OUT%2FSTAT%2Fstandard%2FMDCSTAT02401_OUT") {
		t.Fatalf("form body missing bld: %s", formBody)
	}
}
