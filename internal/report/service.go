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

const fallbackBusinessDays = 7

type flowDataResult struct {
	Query    QueryPeriod
	Rows     map[Market][]Flow
	Fallback bool
}

type tickerDataResult struct {
	Query    QueryPeriod
	Rows     map[Market][]Ticker
	Fallback bool
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
	case CommandInterest:
		return s.InterestReport(ctx, cmd.Period, cmd.Limit, now)
	case CommandTradingValue:
		return s.TradingValueReport(ctx, cmd.Period, cmd.Limit, now)
	case CommandFlowTop:
		return s.FlowTopReport(ctx, cmd.Period, cmd.Limit, now)
	case CommandAfterHours:
		return s.AfterHoursReport(ctx, cmd.Limit, now)
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
	data, err := s.loadFlowData(ctx, period, now)
	if err != nil {
		return "", err
	}
	if !hasAnyFlow(data.Rows) {
		return noDataMessage("연기금등 수급 리포트", BuildQueryPeriod(period, dateOnly(now.In(s.loc)))), nil
	}
	tickers, _ := s.fetchTickers(ctx, data.Query.End)

	var b strings.Builder
	appendReportHeader(&b, "연기금등 수급 리포트", data.Query, data.Fallback)
	b.WriteString("연기금등은 국민연금 단독 매매가 아니라 KRX 투자자 분류상 연기금등 집계입니다.\n\n")

	for _, market := range []Market{MarketKOSPI, MarketKOSDAQ} {
		appendMarketReport(&b, market, data.Rows[market], tickersByCode(tickers[market]), limit)
	}
	return strings.TrimSpace(b.String()), nil
}

func (s *Service) StockReport(ctx context.Context, code string, period Period, now time.Time) (string, error) {
	data, err := s.loadFlowData(ctx, period, now)
	if err != nil {
		return "", err
	}
	if !hasAnyFlow(data.Rows) {
		return noDataMessage("종목 수급 리포트", BuildQueryPeriod(period, dateOnly(now.In(s.loc)))), nil
	}
	tickers, _ := s.fetchTickers(ctx, data.Query.End)

	matches := make([]stockMatch, 0)
	for _, market := range []Market{MarketKOSPI, MarketKOSDAQ} {
		for _, row := range data.Rows[market] {
			if score, ok := stockMatchScore(row, code); ok {
				matches = append(matches, stockMatch{Market: market, Row: row, Score: score})
			}
		}
	}
	if len(matches) == 1 || hasSingleBestMatch(matches) {
		sort.SliceStable(matches, func(i, j int) bool {
			return matches[i].Score < matches[j].Score
		})
		best := matches[0]
		return formatStockReport(best.Market, best.Row, data.Query, data.Fallback, tickersByCode(tickers[best.Market])), nil
	}
	if len(matches) > 1 {
		return formatAmbiguousStockMatches(code, matches, data.Query, data.Fallback), nil
	}
	return fmt.Sprintf("%s 종목을 %s 연기금등 수급 데이터에서 찾지 못했습니다.", code, data.Query.Label), nil
}

func (s *Service) TradingValueReport(ctx context.Context, period Period, limit int, now time.Time) (string, error) {
	data, err := s.loadTickerData(ctx, period, now)
	if err != nil {
		return "", err
	}
	if !hasAnyTicker(data.Rows) {
		return noDataMessage("거래대금 상위 리포트", BuildQueryPeriod(period, dateOnly(now.In(s.loc)))), nil
	}

	var b strings.Builder
	appendReportHeader(&b, "거래대금 상위 리포트", data.Query, data.Fallback)
	for _, market := range []Market{MarketKOSPI, MarketKOSDAQ} {
		appendTradingValueRanking(&b, market, data.Rows[market], limit)
	}
	return strings.TrimSpace(b.String()), nil
}

func (s *Service) FlowTopReport(ctx context.Context, period Period, limit int, now time.Time) (string, error) {
	data, err := s.loadFlowData(ctx, period, now)
	if err != nil {
		return "", err
	}
	if !hasAnyFlow(data.Rows) {
		return noDataMessage("연기금등 순매수 상위 리포트", BuildQueryPeriod(period, dateOnly(now.In(s.loc)))), nil
	}
	tickers, _ := s.fetchTickers(ctx, data.Query.End)

	var b strings.Builder
	appendReportHeader(&b, "연기금등 순매수 상위 리포트", data.Query, data.Fallback)
	b.WriteString("연기금등은 국민연금 단독 매매가 아니라 KRX 투자자 분류상 연기금등 집계입니다.\n\n")
	for _, market := range []Market{MarketKOSPI, MarketKOSDAQ} {
		appendFlowTopRanking(&b, market, data.Rows[market], tickersByCode(tickers[market]), limit)
	}
	return strings.TrimSpace(b.String()), nil
}

func (s *Service) InterestReport(ctx context.Context, period Period, limit int, now time.Time) (string, error) {
	data, err := s.loadFlowData(ctx, period, now)
	if err != nil {
		return "", err
	}
	if !hasAnyFlow(data.Rows) {
		return noDataMessage("관심 종목 리포트", BuildQueryPeriod(period, dateOnly(now.In(s.loc)))), nil
	}
	tickers, _ := s.fetchTickers(ctx, data.Query.End)

	var b strings.Builder
	appendReportHeader(&b, "관심 종목 리포트", data.Query, data.Fallback)
	b.WriteString("거래대금 상위와 연기금등 순매수 상위를 함께 보여줍니다.\n\n")
	for _, market := range []Market{MarketKOSPI, MarketKOSDAQ} {
		fmt.Fprintf(&b, "[%s]\n", market)
		appendTradingValueItems(&b, "거래대금", tickers[market], min(limit, 5))
		appendFlowTopItems(&b, "연기금등 순매수", data.Rows[market], tickersByCode(tickers[market]), min(limit, 5))
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String()), nil
}

func (s *Service) AfterHoursReport(ctx context.Context, limit int, now time.Time) (string, error) {
	rowsByMarket := make(map[Market][]AfterHoursStock, 2)
	for _, market := range []Market{MarketKOSPI, MarketKOSDAQ} {
		rows, err := s.source.AfterHoursGainers(ctx, market)
		if err != nil {
			return "", fmt.Errorf("시간외 급등 데이터 조회 실패: %w", err)
		}
		rowsByMarket[market] = rows
	}

	var b strings.Builder
	b.WriteString("시간외 급등 리포트\n")
	fmt.Fprintf(&b, "조회 시각: %s\n\n", now.In(s.loc).Format("2006-01-02 15:04"))
	b.WriteString("네이버페이 증권 공개 시세의 KRX 시간외 가격 기준입니다.\n\n")
	for _, market := range []Market{MarketKOSPI, MarketKOSDAQ} {
		appendAfterHoursMarketReport(&b, market, rowsByMarket[market], limit)
	}
	b.WriteString("시간외 단일가는 정규장 종가 대비 변동률 기준이며, 장중에는 지연되거나 변동될 수 있습니다.")
	return strings.TrimSpace(b.String()), nil
}

func (s *Service) loadFlowData(ctx context.Context, period Period, now time.Time) (flowDataResult, error) {
	originalEnd := dateOnly(now.In(s.loc))
	end := originalEnd
	var last flowDataResult
	for attempt := 0; attempt <= fallbackBusinessDays; attempt++ {
		query := BuildQueryPeriod(period, end)
		rows, err := s.fetchFlows(ctx, query)
		last = flowDataResult{Query: query, Rows: rows, Fallback: !sameDate(query.End, originalEnd)}
		if err != nil {
			return last, err
		}
		if hasAnyFlow(rows) {
			return last, nil
		}
		end = previousBusinessDay(end)
	}
	return last, nil
}

func (s *Service) loadTickerData(ctx context.Context, period Period, now time.Time) (tickerDataResult, error) {
	originalEnd := dateOnly(now.In(s.loc))
	end := originalEnd
	var last tickerDataResult
	for attempt := 0; attempt <= fallbackBusinessDays; attempt++ {
		query := BuildQueryPeriod(period, end)
		rows, err := s.fetchTickers(ctx, query.End)
		last = tickerDataResult{Query: query, Rows: rows, Fallback: !sameDate(query.End, originalEnd)}
		if err != nil {
			return last, err
		}
		if hasAnyTicker(rows) {
			return last, nil
		}
		end = previousBusinessDay(end)
	}
	return last, nil
}

func (s *Service) fetchFlows(ctx context.Context, query QueryPeriod) (map[Market][]Flow, error) {
	rows := make(map[Market][]Flow, 2)
	for _, market := range []Market{MarketKOSPI, MarketKOSDAQ} {
		marketRows, err := s.source.MarketFlows(ctx, market, query)
		if err != nil {
			return rows, fmt.Errorf("KRX 연기금등 데이터 조회 실패: %w", err)
		}
		rows[market] = marketRows
	}
	return rows, nil
}

func (s *Service) fetchTickers(ctx context.Context, date time.Time) (map[Market][]Ticker, error) {
	rows := make(map[Market][]Ticker, 2)
	for _, market := range []Market{MarketKOSPI, MarketKOSDAQ} {
		marketRows, err := s.source.MarketTickers(ctx, market, date)
		if err != nil {
			return rows, fmt.Errorf("KRX 거래대금 데이터 조회 실패: %w", err)
		}
		rows[market] = marketRows
	}
	return rows, nil
}

func hasAnyFlow(rows map[Market][]Flow) bool {
	for _, marketRows := range rows {
		if len(marketRows) > 0 {
			return true
		}
	}
	return false
}

func hasAnyTicker(rows map[Market][]Ticker) bool {
	for _, marketRows := range rows {
		if len(marketRows) > 0 {
			return true
		}
	}
	return false
}

func (c Command) StockQuery() string {
	if c.Query != "" {
		return c.Query
	}
	return c.Code
}

func normalizeStockQuery(value string) string {
	replacer := strings.NewReplacer(" ", "", ".", "", "-", "", "_", "", "(", "", ")", "")
	return strings.ToLower(replacer.Replace(value))
}

type stockMatch struct {
	Market Market
	Row    Flow
	Score  int
}

func stockMatchScore(row Flow, query string) (int, bool) {
	normalizedQuery := normalizeStockQuery(query)
	if normalizedQuery == "" {
		return 0, false
	}
	code := normalizeStockQuery(row.Code)
	name := normalizeStockQuery(row.Name)
	if code == normalizedQuery || name == normalizedQuery {
		return 0, true
	}
	for _, alias := range stockAliases[name] {
		if normalizeStockQuery(alias) == normalizedQuery {
			return 1, true
		}
	}
	if len([]rune(normalizedQuery)) >= 2 && strings.Contains(name, normalizedQuery) {
		return 2, true
	}
	return 0, false
}

var stockAliases = map[string][]string{
	"삼성전자":      {"삼전", "삼성전자보통주"},
	"sk하이닉스":    {"하이닉스", "하닉", "skhynix", "sk hynix"},
	"lg에너지솔루션":  {"엘지에너지솔루션", "lg엔솔", "엘지엔솔", "엔솔"},
	"lg디스플레이":   {"엘지디스플레이", "엘디플"},
	"lg전자":      {"엘지전자"},
	"lg화학":      {"엘지화학"},
	"lg이노텍":     {"엘지이노텍"},
	"현대차":       {"현차", "현대자동차"},
	"기아":        {"기아차"},
	"naver":     {"네이버"},
	"posco홀딩스":  {"포스코홀딩스", "포홀"},
	"카카오뱅크":     {"카뱅"},
	"카카오페이":     {"카페이"},
	"셀트리온":      {"셀트"},
	"두산에너빌리티":   {"두산에너", "두빌"},
	"한화오션":      {"한화조선"},
	"한화에어로스페이스": {"한화에어로"},
	"hd현대중공업":   {"현대중공업", "현중"},
	"hd한국조선해양":  {"한국조선해양", "한조해", "hd현대조선"},
	"에코프로비엠":    {"에코비"},
	"에코프로":      {"에코"},
	"리가켐바이오":    {"리가켐"},
	"jypent":    {"jyp", "제왑", "제이와이피"},
}

func hasSingleBestMatch(matches []stockMatch) bool {
	if len(matches) == 0 {
		return false
	}
	best := matches[0].Score
	for _, match := range matches[1:] {
		if match.Score < best {
			best = match.Score
		}
	}
	count := 0
	for _, match := range matches {
		if match.Score == best {
			count++
		}
	}
	return count == 1 && best < 2
}

func formatAmbiguousStockMatches(query string, matches []stockMatch, period QueryPeriod, fallback bool) string {
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score < matches[j].Score
		}
		return matches[i].Row.NetValue > matches[j].Row.NetValue
	})

	var b strings.Builder
	fmt.Fprintf(&b, "%q 검색 결과가 여러 종목에 해당합니다. 종목명이나 코드를 더 구체적으로 입력해 주세요.\n", query)
	fmt.Fprintf(&b, "기준: %s / %s", period.Label, period.End.Format("2006-01-02"))
	if fallback {
		b.WriteString(" (최근 조회 가능 거래일)")
	}
	b.WriteString("\n\n")
	limit := min(len(matches), 10)
	for i := 0; i < limit; i++ {
		row := matches[i].Row
		fmt.Fprintf(&b, "%d. %s(%s) %s\n", i+1, row.Name, row.Code, formatWon(row.NetValue))
	}
	if limit > 0 {
		fmt.Fprintf(&b, "\n예: /종목 %s", matches[0].Row.Code)
	}
	return strings.TrimSpace(b.String())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func appendReportHeader(b *strings.Builder, title string, query QueryPeriod, fallback bool) {
	fmt.Fprintf(b, "%s (%s)\n", title, query.Label)
	fmt.Fprintf(b, "기준일: %s", query.End.Format("2006-01-02"))
	if fallback {
		b.WriteString(" (최근 조회 가능 거래일)")
	}
	b.WriteString("\n\n")
}

func noDataMessage(title string, query QueryPeriod) string {
	var b strings.Builder
	appendReportHeader(&b, title, query, false)
	b.WriteString("최근 조회 가능한 KRX 데이터가 없습니다.\n")
	b.WriteString("장전, 휴장일, KRX 데이터 지연일 수 있습니다. 잠시 후 다시 시도해 주세요.")
	return strings.TrimSpace(b.String())
}

func BuildQueryPeriod(period Period, now time.Time) QueryPeriod {
	end := dateOnly(now)
	switch period {
	case Period5D:
		return QueryPeriod{Start: businessDaysBefore(end, 4), End: end, Label: "최근 5거래일"}
	case Period10D:
		return QueryPeriod{Start: businessDaysBefore(end, 9), End: end, Label: "최근 10거래일"}
	case Period20D:
		return QueryPeriod{Start: businessDaysBefore(end, 19), End: end, Label: "최근 20거래일"}
	default:
		return QueryPeriod{Start: end, End: end, Label: "오늘"}
	}
}

func appendMarketReport(b *strings.Builder, market Market, rows []Flow, tickers map[string]Ticker, limit int) {
	fmt.Fprintf(b, "[%s]\n", market)
	if len(rows) == 0 {
		b.WriteString("조회된 데이터가 없습니다.\n\n")
		return
	}

	buys := append([]Flow(nil), rows...)
	sort.SliceStable(buys, func(i, j int) bool {
		return buys[i].NetValue > buys[j].NetValue
	})
	appendRanking(b, "순매수", buys, tickers, limit, true)

	sells := append([]Flow(nil), rows...)
	sort.SliceStable(sells, func(i, j int) bool {
		return sells[i].NetValue < sells[j].NetValue
	})
	appendRanking(b, "순매도", sells, tickers, limit, false)
	b.WriteString("\n")
}

func appendRanking(b *strings.Builder, title string, rows []Flow, tickers map[string]Ticker, limit int, positive bool) {
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
		fmt.Fprintf(b, "%d. %s(%s) %s%s\n", count, row.Name, row.Code, formatWon(row.NetValue), flowRatioSuffix(row, tickers))
		if count >= limit {
			break
		}
	}
	if count == 0 {
		b.WriteString("- 없음\n")
	}
}

func appendTradingValueRanking(b *strings.Builder, market Market, rows []Ticker, limit int) {
	fmt.Fprintf(b, "[%s]\n", market)
	appendTradingValueItems(b, fmt.Sprintf("거래대금 TOP %d", limit), rows, limit)
	b.WriteString("\n")
}

func appendTradingValueItems(b *strings.Builder, title string, rows []Ticker, limit int) {
	fmt.Fprintf(b, "%s\n", title)
	if len(rows) == 0 {
		b.WriteString("- 없음\n")
		return
	}
	sorted := append([]Ticker(nil), rows...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].TradeValue > sorted[j].TradeValue
	})
	count := 0
	for _, row := range sorted {
		if row.TradeValue <= 0 {
			continue
		}
		count++
		fmt.Fprintf(b, "%d. %s(%s) %s\n", count, row.Name, row.Code, formatUnsignedWon(row.TradeValue))
		if count >= limit {
			break
		}
	}
	if count == 0 {
		b.WriteString("- 없음\n")
	}
}

func appendFlowTopRanking(b *strings.Builder, market Market, rows []Flow, tickers map[string]Ticker, limit int) {
	fmt.Fprintf(b, "[%s]\n", market)
	appendFlowTopItems(b, fmt.Sprintf("연기금등 순매수 TOP %d", limit), rows, tickers, limit)
	b.WriteString("\n")
}

func appendFlowTopItems(b *strings.Builder, title string, rows []Flow, tickers map[string]Ticker, limit int) {
	fmt.Fprintf(b, "%s\n", title)
	if len(rows) == 0 {
		b.WriteString("- 없음\n")
		return
	}
	sorted := append([]Flow(nil), rows...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].NetValue > sorted[j].NetValue
	})
	count := 0
	for _, row := range sorted {
		if row.NetValue <= 0 {
			continue
		}
		count++
		fmt.Fprintf(b, "%d. %s(%s) %s%s\n", count, row.Name, row.Code, formatWon(row.NetValue), flowRatioSuffix(row, tickers))
		if count >= limit {
			break
		}
	}
	if count == 0 {
		b.WriteString("- 없음\n")
	}
}

func appendAfterHoursMarketReport(b *strings.Builder, market Market, rows []AfterHoursStock, limit int) {
	fmt.Fprintf(b, "[%s]\n", market)
	fmt.Fprintf(b, "시간외 급등 TOP %d\n", limit)
	if len(rows) == 0 {
		b.WriteString("- 없음\n\n")
		return
	}
	sorted := append([]AfterHoursStock(nil), rows...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].AfterChangeRate > sorted[j].AfterChangeRate
	})
	count := 0
	for _, row := range sorted {
		if row.AfterChangeRate <= 0 {
			continue
		}
		count++
		fmt.Fprintf(b, "%d. %s(%s) %+.2f%%\n", count, row.Name, row.Code, row.AfterChangeRate)
		fmt.Fprintf(b, "   시간외가: %s원 / 대비: %s원\n", formatUnsignedInt(row.AfterPrice), formatInt(row.AfterChange))
		if count >= limit {
			break
		}
	}
	if count == 0 {
		b.WriteString("- 없음\n")
	}
	b.WriteString("\n")
}

func formatStockReport(market Market, row Flow, query QueryPeriod, fallback bool, tickers map[string]Ticker) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s(%s) 연기금등 수급 (%s)\n", row.Name, row.Code, query.Label)
	fmt.Fprintf(&b, "기준일: %s", query.End.Format("2006-01-02"))
	if fallback {
		b.WriteString(" (최근 조회 가능 거래일)")
	}
	fmt.Fprintf(&b, "\n시장: %s\n", market)
	fmt.Fprintf(&b, "순매수 금액: %s\n", formatWon(row.NetValue))
	if ratio, ok := flowRatio(row, tickers); ok {
		fmt.Fprintf(&b, "거래대금 대비: %+.2f%%\n", ratio)
	}
	fmt.Fprintf(&b, "매수: %s / 매도: %s\n", formatUnsignedWon(row.BuyValue), formatUnsignedWon(row.SellValue))
	fmt.Fprintf(&b, "순매수 수량: %s주\n\n", formatInt(row.NetVolume))
	b.WriteString("연기금등은 국민연금 단독 매매가 아니라 KRX 투자자 분류상 연기금등 집계입니다.")
	return strings.TrimSpace(b.String())
}

func tickersByCode(rows []Ticker) map[string]Ticker {
	byCode := make(map[string]Ticker, len(rows))
	for _, row := range rows {
		byCode[row.Code] = row
	}
	return byCode
}

func flowRatioSuffix(row Flow, tickers map[string]Ticker) string {
	ratio, ok := flowRatio(row, tickers)
	if !ok {
		return ""
	}
	return fmt.Sprintf(" (거래대금 대비 %+.2f%%)", ratio)
}

func flowRatio(row Flow, tickers map[string]Ticker) (float64, bool) {
	if row.NetValue == 0 || len(tickers) == 0 {
		return 0, false
	}
	ticker, ok := tickers[row.Code]
	if !ok || ticker.TradeValue <= 0 {
		return 0, false
	}
	return float64(row.NetValue) / float64(ticker.TradeValue) * 100, true
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
	return sign + formatKoreanWon(value)
}

func formatUnsignedWon(value int64) string {
	if value < 0 {
		value = -value
	}
	return formatKoreanWon(value)
}

func formatKoreanWon(value int64) string {
	const trillion = int64(1_000_000_000_000)

	jo := value / trillion
	remainder := value % trillion
	if jo == 0 {
		return formatEok(remainder) + "억원"
	}
	if remainder == 0 {
		return fmt.Sprintf("%d조원", jo)
	}
	return fmt.Sprintf("%d조 %s억원", jo, formatEok(remainder))
}

func formatEok(value int64) string {
	return addCommaToDecimal(fmt.Sprintf("%.1f", float64(value)/100_000_000))
}

func addCommaToDecimal(value string) string {
	parts := strings.SplitN(value, ".", 2)
	integer := parts[0]
	for i := len(integer) - 3; i > 0; i -= 3 {
		integer = integer[:i] + "," + integer[i:]
	}
	if len(parts) == 1 {
		return integer
	}
	return integer + "." + parts[1]
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

func formatUnsignedInt(value int64) string {
	if value < 0 {
		value = -value
	}
	digits := fmt.Sprintf("%d", value)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + "," + digits[i:]
	}
	return digits
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

func previousBusinessDay(end time.Time) time.Time {
	return businessDaysBefore(end, 1)
}

func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func startMessage() string {
	return "KRX 주식 봇입니다.\n현재는 KRX 연기금등 수급 리포트를 제공합니다.\n/help 명령어로 사용법을 확인하세요.\n\n연기금등은 국민연금 단독 매매가 아니라 KRX 투자자 분류상 연기금등 집계입니다."
}

func helpMessage() string {
	return strings.TrimSpace(`사용 가능한 명령어
/연기금 오늘
/연기금 5일
/연기금 10일
/연기금 20일
/연기금 오늘 20
/관심 오늘
/거래대금 오늘
/수급상위 오늘
/시간외 급등
/시간외 급등 20
/종목 005930
/종목 삼성전자
/종목 삼전
/종목 하이닉스
/종목 카뱅
/종목 두산에너
/종목 005930 10일
/종목 005930 20일
삼성전자

영문 명령어도 사용할 수 있습니다.
/pension today
/pension 5d
/pension 10d
/pension 20d
/pension today 20
/stock 005930
/stock 005930 20d
/afterhours up

조회 결과는 투자 참고용이며 매매 추천이 아닙니다.`)
}

func unknownMessage() string {
	return "알 수 없는 명령어입니다. /help 를 입력해 사용 가능한 명령어를 확인하세요."
}
