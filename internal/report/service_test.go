package report

import (
	"context"
	"strings"
	"testing"
	"time"
)

type fakeSource struct {
	rows          map[Market][]Flow
	tickers       map[Market][]Ticker
	afterHours    map[Market][]AfterHoursStock
	intraday      map[string][]IntradayPrice
	rowsByEnd     map[string]map[Market][]Flow
	tickersByDate map[string]map[Market][]Ticker
}

func (f fakeSource) MarketFlows(ctx context.Context, market Market, period QueryPeriod) ([]Flow, error) {
	if f.rowsByEnd != nil {
		return f.rowsByEnd[period.End.Format("20060102")][market], nil
	}
	return f.rows[market], nil
}

func (f fakeSource) MarketTickers(ctx context.Context, market Market, date time.Time) ([]Ticker, error) {
	if f.tickersByDate != nil {
		return f.tickersByDate[date.Format("20060102")][market], nil
	}
	return f.tickers[market], nil
}

func (f fakeSource) AfterHoursGainers(ctx context.Context, market Market) ([]AfterHoursStock, error) {
	return f.afterHours[market], nil
}

func (f fakeSource) IntradayPrices(ctx context.Context, code string, date time.Time) ([]IntradayPrice, error) {
	return f.intraday[code], nil
}

func TestPensionReportShowsBothMarketsAndDisclaimer(t *testing.T) {
	svc := NewService(fakeSource{rows: map[Market][]Flow{
		MarketKOSPI: {
			{Code: "005930", Name: "삼성전자", NetValue: 12_300_000_000},
			{Code: "000660", Name: "SK하이닉스", NetValue: -5_000_000_000},
		},
		MarketKOSDAQ: {
			{Code: "091990", Name: "셀트리온헬스케어", NetValue: 3_300_000_000},
			{Code: "035900", Name: "JYP Ent.", NetValue: -2_200_000_000},
		},
	}}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.PensionReport(context.Background(), PeriodToday, 10, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("PensionReport() error = %v", err)
	}

	for _, want := range []string{"KOSPI", "KOSDAQ", "순매수 TOP 10", "순매도 TOP 10", "국민연금 단독", "삼성전자", "JYP Ent."} {
		if !strings.Contains(msg, want) {
			t.Fatalf("report missing %q:\n%s", want, msg)
		}
	}
}

func TestStockReportFindsCodeAcrossMarkets(t *testing.T) {
	svc := NewService(fakeSource{rows: map[Market][]Flow{
		MarketKOSPI: {
			{Code: "005930", Name: "삼성전자", BuyValue: 20_000_000_000, SellValue: 7_700_000_000, NetValue: 12_300_000_000, NetVolume: 1000},
		},
		MarketKOSDAQ: {},
	}}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.StockReport(context.Background(), "005930", Period20D, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("StockReport() error = %v", err)
	}

	for _, want := range []string{"005930", "삼성전자", "KOSPI", "+123.0억원", "최근 20거래일", "매수: 200.0억원 / 매도: 77.0억원"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("stock report missing %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "매수: +") || strings.Contains(msg, "매도: +") {
		t.Fatalf("stock report should not sign gross buy/sell values:\n%s", msg)
	}
}

func TestStockReportFindsNameAcrossMarkets(t *testing.T) {
	svc := NewService(fakeSource{rows: map[Market][]Flow{
		MarketKOSPI: {
			{Code: "005930", Name: "삼성전자", BuyValue: 20_000_000_000, SellValue: 7_700_000_000, NetValue: 12_300_000_000, NetVolume: 1000},
		},
		MarketKOSDAQ: {},
	}}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.StockReport(context.Background(), "삼성전자", PeriodToday, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("StockReport() error = %v", err)
	}

	for _, want := range []string{"005930", "삼성전자", "KOSPI", "+123.0억원"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("stock report missing %q:\n%s", want, msg)
		}
	}
}

func TestStockReportFindsCommonAliases(t *testing.T) {
	svc := NewService(fakeSource{rows: map[Market][]Flow{
		MarketKOSPI: {
			{Code: "005930", Name: "삼성전자", NetValue: 12_300_000_000},
			{Code: "000660", Name: "SK하이닉스", NetValue: 7_000_000_000},
		},
		MarketKOSDAQ: {},
	}}, time.FixedZone("KST", 9*60*60))

	tests := []struct {
		query string
		want  string
	}{
		{"삼전", "삼성전자(005930)"},
		{"하이닉스", "SK하이닉스(000660)"},
		{"sk hynix", "SK하이닉스(000660)"},
	}

	for _, tt := range tests {
		msg, err := svc.StockReport(context.Background(), tt.query, PeriodToday, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("StockReport(%q) error = %v", tt.query, err)
		}
		if !strings.Contains(msg, tt.want) {
			t.Fatalf("StockReport(%q) missing %q:\n%s", tt.query, tt.want, msg)
		}
	}
}

func TestStockReportFindsExpandedAliases(t *testing.T) {
	svc := NewService(fakeSource{rows: map[Market][]Flow{
		MarketKOSPI: {
			{Code: "034020", Name: "두산에너빌리티", NetValue: 12_300_000_000},
			{Code: "323410", Name: "카카오뱅크", NetValue: 7_000_000_000},
		},
		MarketKOSDAQ: {
			{Code: "035900", Name: "JYP Ent.", NetValue: 3_000_000_000},
		},
	}}, time.FixedZone("KST", 9*60*60))

	tests := []struct {
		query string
		want  string
	}{
		{"두산에너", "두산에너빌리티(034020)"},
		{"카뱅", "카카오뱅크(323410)"},
		{"jyp", "JYP Ent.(035900)"},
	}

	for _, tt := range tests {
		msg, err := svc.StockReport(context.Background(), tt.query, PeriodToday, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("StockReport(%q) error = %v", tt.query, err)
		}
		if !strings.Contains(msg, tt.want) {
			t.Fatalf("StockReport(%q) missing %q:\n%s", tt.query, tt.want, msg)
		}
	}
}

func TestStockReportShowsCandidatesForAmbiguousName(t *testing.T) {
	svc := NewService(fakeSource{rows: map[Market][]Flow{
		MarketKOSPI: {
			{Code: "005930", Name: "삼성전자", NetValue: 12_300_000_000},
			{Code: "006400", Name: "삼성SDI", NetValue: 7_000_000_000},
			{Code: "018260", Name: "삼성에스디에스", NetValue: 2_000_000_000},
		},
		MarketKOSDAQ: {},
	}}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.StockReport(context.Background(), "삼성", PeriodToday, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("StockReport() error = %v", err)
	}

	for _, want := range []string{"여러 종목", "삼성전자(005930)", "삼성SDI(006400)", "/종목 005930"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("ambiguous report missing %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "순매수 금액") {
		t.Fatalf("ambiguous report should not return a single stock report:\n%s", msg)
	}
}

func TestPensionReportFallsBackToPreviousBusinessDayWhenTodayIsEmpty(t *testing.T) {
	svc := NewService(fakeSource{rowsByEnd: map[string]map[Market][]Flow{
		"20260526": {
			MarketKOSPI:  {},
			MarketKOSDAQ: {},
		},
		"20260525": {
			MarketKOSPI: {
				{Code: "005930", Name: "삼성전자", NetValue: 12_300_000_000},
			},
			MarketKOSDAQ: {},
		},
	}}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.PensionReport(context.Background(), PeriodToday, 10, time.Date(2026, 5, 26, 10, 0, 0, 0, time.FixedZone("KST", 9*60*60)))
	if err != nil {
		t.Fatalf("PensionReport() error = %v", err)
	}

	for _, want := range []string{"기준일: 2026-05-25", "최근 조회 가능 거래일", "삼성전자"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("fallback report missing %q:\n%s", want, msg)
		}
	}
}

func TestPensionReportShowsNoDataMessageAfterFallbackAttempts(t *testing.T) {
	svc := NewService(fakeSource{}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.PensionReport(context.Background(), PeriodToday, 10, time.Date(2026, 5, 26, 10, 0, 0, 0, time.FixedZone("KST", 9*60*60)))
	if err != nil {
		t.Fatalf("PensionReport() error = %v", err)
	}

	for _, want := range []string{"최근 조회 가능한 KRX 데이터가 없습니다", "장전", "휴장일"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("no data report missing %q:\n%s", want, msg)
		}
	}
}

func TestBuildQueryPeriodSupports10TradingDays(t *testing.T) {
	got := BuildQueryPeriod(Period10D, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if got.Label != "최근 10거래일" {
		t.Fatalf("Label = %q, want 최근 10거래일", got.Label)
	}
	if got.Start.Format("2006-01-02") != "2026-05-13" {
		t.Fatalf("Start = %s, want 2026-05-13", got.Start.Format("2006-01-02"))
	}
}

func TestFlowTopReportShowsTradingValueRatio(t *testing.T) {
	svc := NewService(fakeSource{
		rows: map[Market][]Flow{
			MarketKOSPI: {
				{Code: "005930", Name: "삼성전자", NetValue: 10_000_000_000},
			},
			MarketKOSDAQ: {},
		},
		tickers: map[Market][]Ticker{
			MarketKOSPI: {
				{Code: "005930", Name: "삼성전자", TradeValue: 200_000_000_000},
			},
			MarketKOSDAQ: {},
		},
	}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.FlowTopReport(context.Background(), PeriodToday, 10, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("FlowTopReport() error = %v", err)
	}

	for _, want := range []string{"삼성전자", "거래대금 대비 +5.00%"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("flow top report missing %q:\n%s", want, msg)
		}
	}
}

func TestTradingValueReportShowsBothMarkets(t *testing.T) {
	svc := NewService(fakeSource{tickers: map[Market][]Ticker{
		MarketKOSPI: {
			{Code: "005930", Name: "삼성전자", TradeValue: 2_000_000_000_000},
		},
		MarketKOSDAQ: {
			{Code: "035900", Name: "JYP Ent.", TradeValue: 120_000_000_000},
		},
	}}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.TradingValueReport(context.Background(), PeriodToday, 10, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("TradingValueReport() error = %v", err)
	}

	for _, want := range []string{"거래대금 상위", "KOSPI", "KOSDAQ", "삼성전자", "JYP Ent.", "2조원"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("trading value report missing %q:\n%s", want, msg)
		}
	}
}

func TestFormatWonUsesReadableKoreanUnits(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"under trillion", formatUnsignedWon(152_580_000_000), "1,525.8억원"},
		{"exact trillion", formatUnsignedWon(2_000_000_000_000), "2조원"},
		{"trillion with remainder", formatUnsignedWon(9_864_470_000_000), "9조 8,644.7억원"},
		{"signed trillion", formatWon(-1_234_560_000_000), "-1조 2,345.6억원"},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Fatalf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestFlowTopReportShowsNetBuyOnly(t *testing.T) {
	svc := NewService(fakeSource{rows: map[Market][]Flow{
		MarketKOSPI: {
			{Code: "005930", Name: "삼성전자", NetValue: 12_300_000_000},
			{Code: "000660", Name: "SK하이닉스", NetValue: -5_000_000_000},
		},
		MarketKOSDAQ: {},
	}}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.FlowTopReport(context.Background(), PeriodToday, 10, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("FlowTopReport() error = %v", err)
	}

	if !strings.Contains(msg, "연기금등 순매수 상위") || !strings.Contains(msg, "삼성전자") {
		t.Fatalf("flow top report missing expected content:\n%s", msg)
	}
	if strings.Contains(msg, "SK하이닉스") {
		t.Fatalf("flow top report should not include net sell row:\n%s", msg)
	}
}

func TestInterestReportCombinesTradingValueAndPensionFlow(t *testing.T) {
	svc := NewService(fakeSource{
		rows: map[Market][]Flow{
			MarketKOSPI: {
				{Code: "005930", Name: "삼성전자", NetValue: 12_300_000_000},
			},
			MarketKOSDAQ: {},
		},
		tickers: map[Market][]Ticker{
			MarketKOSPI: {
				{Code: "005930", Name: "삼성전자", TradeValue: 2_000_000_000_000},
			},
			MarketKOSDAQ: {},
		},
	}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.InterestReport(context.Background(), PeriodToday, 10, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("InterestReport() error = %v", err)
	}

	for _, want := range []string{"관심 종목", "거래대금", "연기금등 순매수", "삼성전자"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("interest report missing %q:\n%s", want, msg)
		}
	}
}

func TestAfterHoursReportShowsBothMarkets(t *testing.T) {
	svc := NewService(fakeSource{afterHours: map[Market][]AfterHoursStock{
		MarketKOSPI: {
			{Code: "011070", Name: "LG이노텍", AfterPrice: 1_123_000, AfterChange: 259_000, AfterChangeRate: 29.98},
			{Code: "005930", Name: "삼성전자", AfterPrice: 85_000, AfterChange: 1_000, AfterChangeRate: 1.19},
		},
		MarketKOSDAQ: {
			{Code: "356680", Name: "엑스게이트", AfterPrice: 27_550, AfterChange: 6_350, AfterChangeRate: 29.95},
		},
	}}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.AfterHoursReport(context.Background(), 10, time.Date(2026, 5, 26, 18, 10, 0, 0, time.FixedZone("KST", 9*60*60)))
	if err != nil {
		t.Fatalf("AfterHoursReport() error = %v", err)
	}

	for _, want := range []string{"시간외 급등", "KOSPI", "KOSDAQ", "LG이노텍(011070) +29.98%", "시간외가: 1,123,000원", "네이버페이 증권"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("after-hours report missing %q:\n%s", want, msg)
		}
	}
}

func TestSignalReportRanksByCompositeSignal(t *testing.T) {
	svc := NewService(fakeSource{
		rows: map[Market][]Flow{
			MarketKOSPI: {
				{Code: "005930", Name: "삼성전자", NetValue: 12_000_000_000},
				{Code: "000660", Name: "SK하이닉스", NetValue: 3_000_000_000},
			},
			MarketKOSDAQ: {},
		},
		tickers: map[Market][]Ticker{
			MarketKOSPI: {
				{Code: "005930", Name: "삼성전자", TradeValue: 200_000_000_000, ChangeRate: 2.4},
				{Code: "000660", Name: "SK하이닉스", TradeValue: 100_000_000_000, ChangeRate: -1.2},
			},
			MarketKOSDAQ: {},
		},
		afterHours: map[Market][]AfterHoursStock{
			MarketKOSPI: {
				{Code: "005930", Name: "삼성전자", AfterChangeRate: 1.5},
			},
			MarketKOSDAQ: {},
		},
	}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.SignalReport(context.Background(), PeriodToday, 10, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("SignalReport() error = %v", err)
	}

	for _, want := range []string{"관심 신호", "KOSPI", "삼성전자(005930)", "점", "거래대금 대비 +6.00%", "시간외 +1.50%", "매매 추천이 아닙니다"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("signal report missing %q:\n%s", want, msg)
		}
	}
}

func TestSignalStockReportFindsName(t *testing.T) {
	svc := NewService(fakeSource{
		rows: map[Market][]Flow{
			MarketKOSPI: {
				{Code: "005930", Name: "삼성전자", NetValue: 12_000_000_000},
			},
			MarketKOSDAQ: {},
		},
		tickers: map[Market][]Ticker{
			MarketKOSPI: {
				{Code: "005930", Name: "삼성전자", TradeValue: 200_000_000_000, ChangeRate: 2.4},
			},
			MarketKOSDAQ: {},
		},
	}, time.FixedZone("KST", 9*60*60))

	msg, err := svc.SignalStockReport(context.Background(), "삼성전자", time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("SignalStockReport() error = %v", err)
	}

	for _, want := range []string{"삼성전자(005930) 관심 신호", "KOSPI", "점수", "연기금등"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("signal stock report missing %q:\n%s", want, msg)
		}
	}
}

func TestSessionPerformanceReportUsesIntradayPrices(t *testing.T) {
	loc := time.FixedZone("KST", 9*60*60)
	date := time.Date(2026, 5, 26, 0, 0, 0, 0, loc)
	svc := NewService(fakeSource{
		rows: map[Market][]Flow{
			MarketKOSPI: {
				{Code: "005930", Name: "삼성전자", NetValue: 12_000_000_000},
				{Code: "000660", Name: "SK하이닉스", NetValue: 3_000_000_000},
			},
			MarketKOSDAQ: {},
		},
		tickers: map[Market][]Ticker{
			MarketKOSPI: {
				{Code: "005930", Name: "삼성전자", TradeValue: 200_000_000_000, ChangeRate: 2.4},
				{Code: "000660", Name: "SK하이닉스", TradeValue: 100_000_000_000, ChangeRate: 1.0},
			},
			MarketKOSDAQ: {},
		},
		intraday: map[string][]IntradayPrice{
			"005930": {
				{Time: date.Add(9 * time.Hour), Close: 100_000},
				{Time: date.Add(11*time.Hour + 30*time.Minute), Close: 103_000},
				{Time: date.Add(12 * time.Hour), Close: 102_000},
				{Time: date.Add(15*time.Hour + 30*time.Minute), Close: 101_000},
			},
			"000660": {
				{Time: date.Add(9 * time.Hour), Close: 200_000},
				{Time: date.Add(11*time.Hour + 30*time.Minute), Close: 198_000},
				{Time: date.Add(12 * time.Hour), Close: 198_000},
				{Time: date.Add(15*time.Hour + 30*time.Minute), Close: 202_000},
			},
		},
	}, loc)

	msg, err := svc.SessionPerformanceReport(context.Background(), SessionMorning, PeriodToday, 10, date.Add(16*time.Hour))
	if err != nil {
		t.Fatalf("SessionPerformanceReport() error = %v", err)
	}

	for _, want := range []string{"오전 실적", "09:00 -> 11:30", "삼성전자(005930) +3.00%", "100,000원 -> 103,000원", "가격 기준"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("session report missing %q:\n%s", want, msg)
		}
	}
}
