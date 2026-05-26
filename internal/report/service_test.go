package report

import (
	"context"
	"strings"
	"testing"
	"time"
)

type fakeSource struct {
	rows map[Market][]Flow
}

func (f fakeSource) MarketFlows(ctx context.Context, market Market, period QueryPeriod) ([]Flow, error) {
	return f.rows[market], nil
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
