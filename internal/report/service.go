package report

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Service struct {
	source Source
	loc    *time.Location
}

func NewService(source Source, loc *time.Location) *Service {
	if loc == nil {
		loc = time.FixedZone("KST", 9*60*60)
	}
	return &Service{source: source, loc: loc}
}

func (s *Service) HandleText(ctx context.Context, text string) (string, error) {
	now := time.Now()
	cmd := ParseCommand(text)
	switch cmd.Kind {
	case CommandStart:
		return startMessage(), nil
	case CommandHelp:
		return helpMessage(), nil
	case CommandPension:
		return s.PensionReport(ctx, cmd.Period, cmd.Limit, now)
	case CommandStock:
		return s.StockReport(ctx, cmd.StockQuery(), cmd.Period, now)
	default:
		return unknownMessage(), nil
	}
}

func (s *Service) CheckKOSPI(ctx context.Context) (int, error) {
	query := BuildQueryPeriod(PeriodToday, time.Now().In(s.loc))
	rows, err := s.source.MarketFlows(ctx, MarketKOSPI, query)
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func (s *Service) PensionReport(ctx context.Context, period Period, limit int, now time.Time) (string, error) {
	query := BuildQueryPeriod(period, now.In(s.loc))
	var b strings.Builder
	fmt.Fprintf(&b, "연기금등 수급 리포트 (%s)\n", query.Label)
	fmt.Fprintf(&b, "기준일: %s\n\n", query.End.Format("2006-01-02"))
	b.WriteString("연기금등은 국민연금 단독 매매가 아니라 KRX 투자자 분류상 연기금등 집계입니다.\n\n")

	for _, market := range []Market{MarketKOSPI, MarketKOSDAQ} {
		rows, err := s.source.MarketFlows(ctx, market, query)
		if err != nil {
			return "", err
		}
		appendMarketReport(&b, market, rows, limit)
	}
	return strings.TrimSpace(b.String()), nil
}

func (s *Service) StockReport(ctx context.Context, code string, period Period, now time.Time) (string, error) {
	query := BuildQueryPeriod(period, now.In(s.loc))
	for _, market := range []Market{MarketKOSPI, MarketKOSDAQ} {
		rows, err := s.source.MarketFlows(ctx, market, query)
		if err != nil {
			return "", err
		}
		for _, row := range rows {
			if stockMatches(row, code) {
				return formatStockReport(market, row, query), nil
			}
		}
	}
	return fmt.Sprintf("%s 종목을 %s 연기금등 수급 데이터에서 찾지 못했습니다.", code, query.Label), nil
}

func (c Command) StockQuery() string {
	if c.Query != "" {
		return c.Query
	}
	return c.Code
}

func stockMatches(row Flow, query string) bool {
	normalizedQuery := normalizeStockQuery(query)
	if normalizedQuery == "" {
		return false
	}
	return normalizeStockQuery(row.Code) == normalizedQuery || normalizeStockQuery(row.Name) == normalizedQuery
}

func normalizeStockQuery(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), ""))
}

func BuildQueryPeriod(period Period, now time.Time) QueryPeriod {
	end := dateOnly(now)
	switch period {
	case Period5D:
		return QueryPeriod{Start: businessDaysBefore(end, 4), End: end, Label: "최근 5거래일"}
	case Period20D:
		return QueryPeriod{Start: businessDaysBefore(end, 19), End: end, Label: "최근 20거래일"}
	default:
		return QueryPeriod{Start: end, End: end, Label: "오늘"}
	}
}

func appendMarketReport(b *strings.Builder, market Market, rows []Flow, limit int) {
	fmt.Fprintf(b, "[%s]\n", market)
	if len(rows) == 0 {
		b.WriteString("조회된 데이터가 없습니다.\n\n")
		return
	}

	buys := append([]Flow(nil), rows...)
	sort.SliceStable(buys, func(i, j int) bool {
		return buys[i].NetValue > buys[j].NetValue
	})
	appendRanking(b, "순매수", buys, limit, true)

	sells := append([]Flow(nil), rows...)
	sort.SliceStable(sells, func(i, j int) bool {
		return sells[i].NetValue < sells[j].NetValue
	})
	appendRanking(b, "순매도", sells, limit, false)
	b.WriteString("\n")
}

func appendRanking(b *strings.Builder, title string, rows []Flow, limit int, positive bool) {
	fmt.Fprintf(b, "%s TOP %d\n", title, limit)
	count := 0
	for _, row := range rows {
		if positive && row.NetValue <= 0 {
			continue
		}
		if !positive && row.NetValue >= 0 {
			continue
		}
		count++
		fmt.Fprintf(b, "%d. %s(%s) %s\n", count, row.Name, row.Code, formatWon(row.NetValue))
		if count >= limit {
			break
		}
	}
	if count == 0 {
		b.WriteString("- 없음\n")
	}
}

func formatStockReport(market Market, row Flow, query QueryPeriod) string {
	return fmt.Sprintf(
		"%s(%s) 연기금등 수급 (%s)\n시장: %s\n순매수 금액: %s\n매수: %s / 매도: %s\n순매수 수량: %s주\n\n연기금등은 국민연금 단독 매매가 아니라 KRX 투자자 분류상 연기금등 집계입니다.",
		row.Name,
		row.Code,
		query.Label,
		market,
		formatWon(row.NetValue),
		formatUnsignedWon(row.BuyValue),
		formatUnsignedWon(row.SellValue),
		formatInt(row.NetVolume),
	)
}

func formatWon(value int64) string {
	sign := ""
	if value > 0 {
		sign = "+"
	}
	if value < 0 {
		sign = "-"
		value = -value
	}
	return fmt.Sprintf("%s%.1f억원", sign, float64(value)/100_000_000)
}

func formatUnsignedWon(value int64) string {
	if value < 0 {
		value = -value
	}
	return fmt.Sprintf("%.1f억원", float64(value)/100_000_000)
}

func formatInt(value int64) string {
	sign := ""
	if value > 0 {
		sign = "+"
	}
	if value < 0 {
		sign = "-"
		value = -value
	}
	digits := fmt.Sprintf("%d", value)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + "," + digits[i:]
	}
	return sign + digits
}

func dateOnly(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func businessDaysBefore(end time.Time, daysBefore int) time.Time {
	current := end
	for daysBefore > 0 {
		current = current.AddDate(0, 0, -1)
		if current.Weekday() == time.Saturday || current.Weekday() == time.Sunday {
			continue
		}
		daysBefore--
	}
	return current
}

func startMessage() string {
	return "KRX 주식 봇입니다.\n현재는 KRX 연기금등 수급 리포트를 제공합니다.\n/help 명령어로 사용법을 확인하세요.\n\n연기금등은 국민연금 단독 매매가 아니라 KRX 투자자 분류상 연기금등 집계입니다."
}

func helpMessage() string {
	return strings.TrimSpace(`사용 가능한 명령어
/연기금 오늘
/연기금 5일
/연기금 20일
/연기금 오늘 20
/종목 005930
/종목 삼성전자
/종목 005930 20일
삼성전자

영문 명령어도 사용할 수 있습니다.
/pension today
/pension 5d
/pension 20d
/pension today 20
/stock 005930
/stock 005930 20d

조회 결과는 투자 참고용이며 매매 추천이 아닙니다.`)
}

func unknownMessage() string {
	return "알 수 없는 명령어입니다. /help 를 입력해 사용 가능한 명령어를 확인하세요."
}
