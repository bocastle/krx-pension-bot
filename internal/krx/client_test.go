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

func TestClientFetchesAndParsesMarketTickers(t *testing.T) {
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
		if got := r.Form.Get("strtDd"); got != "20260526" {
			t.Fatalf("strtDd = %q, want 20260526", got)
		}
		if got := r.Form.Get("endDd"); got != "20260526" {
			t.Fatalf("endDd = %q, want 20260526", got)
		}
		if got := r.Form.Get("itmTpCd2"); got != "1" {
			t.Fatalf("itmTpCd2 = %q, want 1", got)
		}
		if got := r.Form.Get("itmTpCd3"); got != "2" {
			t.Fatalf("itmTpCd3 = %q, want 2", got)
		}
		formBody = r.Form.Encode()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"OutBlock_1": [
				{
					"ISU_CD": "005930",
					"ISU_ABBRV": "삼성전자",
					"ACC_TRDVOL": "31,234,567",
					"ACC_TRDVAL": "2,000,000,000,000",
					"CLSPRC": "64,900",
					"FLUC_RT": "1.25"
				}
			]
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 10*time.Minute)
	rows, err := client.MarketTickers(context.Background(), report.MarketKOSPI, time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("MarketTickers() error = %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(rows))
	}
	if rows[0].Code != "005930" || rows[0].Name != "삼성전자" || rows[0].TradeValue != 2_000_000_000_000 {
		t.Fatalf("row = %#v", rows[0])
	}
	if rows[0].TradeVolume != 31_234_567 || rows[0].ClosingPrice != 64_900 || rows[0].ChangeRate != 1.25 {
		t.Fatalf("parsed numeric fields = %#v", rows[0])
	}
	if !strings.Contains(formBody, "bld=dbms%2FMDC_OUT%2FEASY%2Franking%2FMDCEASY01601_OUT") {
		t.Fatalf("form body missing bld: %s", formBody)
	}
}

func TestClientTreatsStringOutputAsEmptyRows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":"","CURRENT_DATETIME":"2026.05.26 PM 03:28:49"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 10*time.Minute)
	rows, err := client.MarketFlows(context.Background(), report.MarketKOSPI, report.QueryPeriod{
		Start: time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("MarketFlows() error = %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("len(rows) = %d, want 0", len(rows))
	}
}

func TestClientFetchesAndParsesAfterHoursGainers(t *testing.T) {
	var requestPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.String()
		if r.URL.Path != "/front-api/domestic/stock/list" {
			t.Fatalf("path = %s, want Naver stock list endpoint", r.URL.Path)
		}
		if got := r.URL.Query().Get("sortType"); got != "up" {
			t.Fatalf("sortType = %q, want up", got)
		}
		if got := r.URL.Query().Get("category"); got != "KOSPI" {
			t.Fatalf("category = %q, want KOSPI", got)
		}
		if got := r.URL.Query().Get("domesticStockExchangeType"); got != "KRX" {
			t.Fatalf("domesticStockExchangeType = %q, want KRX", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"isSuccess": true,
			"result": {
				"stocks": [
					{
						"itemCode": "011070",
						"name": "LG이노텍",
						"overMarketPriceInfo": {
							"tradingSessionType": "AFTER_MARKET",
							"overPrice": "1,123,000",
							"fluctuations": "259,000",
							"fluctuationsType": "RISING",
							"fluctuationsRatio": "29.98"
						}
					},
					{
						"itemCode": "005930",
						"name": "삼성전자"
					}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 10*time.Minute)
	rows, err := client.AfterHoursGainers(context.Background(), report.MarketKOSPI)
	if err != nil {
		t.Fatalf("AfterHoursGainers() error = %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1, path=%s", len(rows), requestPath)
	}
	if rows[0].Code != "011070" || rows[0].Name != "LG이노텍" || rows[0].AfterPrice != 1_123_000 || rows[0].AfterChangeRate != 29.98 {
		t.Fatalf("row = %#v", rows[0])
	}
}

func TestClientFetchesAndParsesIntradayPrices(t *testing.T) {
	var requestPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.String()
		if r.URL.Path != "/siseJson.naver" {
			t.Fatalf("path = %s, want Naver intraday endpoint", r.URL.Path)
		}
		if got := r.URL.Query().Get("symbol"); got != "005930" {
			t.Fatalf("symbol = %q, want 005930", got)
		}
		if got := r.URL.Query().Get("timeframe"); got != "minute" {
			t.Fatalf("timeframe = %q, want minute", got)
		}
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		_, _ = w.Write([]byte(`
 [['날짜', '시가', '고가', '저가', '종가', '거래량', '외국인소진율'],
["202605261130", null, null, null, 103000, 2000, null],
["202605260900", null, null, null, 100000, 1000, null]
]
`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 10*time.Minute)
	rows, err := client.IntradayPrices(context.Background(), "005930", time.Date(2026, 5, 26, 0, 0, 0, 0, time.FixedZone("KST", 9*60*60)))
	if err != nil {
		t.Fatalf("IntradayPrices() error = %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2, path=%s", len(rows), requestPath)
	}
	if rows[0].Time.Format("200601021504") != "202605260900" || rows[0].Close != 100_000 || rows[0].Volume != 1_000 {
		t.Fatalf("first row = %#v", rows[0])
	}
	if rows[1].Time.Format("200601021504") != "202605261130" || rows[1].Close != 103_000 || rows[1].Volume != 2_000 {
		t.Fatalf("second row = %#v", rows[1])
	}
}
