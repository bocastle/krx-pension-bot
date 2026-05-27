package report

import (
	"regexp"
	"strconv"
	"strings"
)

var stockCodePattern = regexp.MustCompile(`^\d{6}$`)

func ParseCommand(text string) Command {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return Command{Kind: CommandUnknown}
	}

	name := strings.ToLower(fields[0])
	if at := strings.IndexByte(name, '@'); at >= 0 {
		name = name[:at]
	}

	if !strings.HasPrefix(name, "/") {
		if isCommandKeyword(name) {
			name = "/" + name
		} else if len(fields) == 1 {
			return stockCommand(fields[0], PeriodToday)
		}
	}

	if !strings.HasPrefix(name, "/") && len(fields) == 1 {
		return stockCommand(fields[0], PeriodToday)
	}

	switch name {
	case "/start", "/시작":
		if len(fields) == 1 {
			return Command{Kind: CommandStart}
		}
	case "/help", "/도움말":
		if len(fields) == 1 {
			return Command{Kind: CommandHelp}
		}
	case "/pension", "/연기금":
		return parsePension(fields[1:])
	case "/stock", "/종목":
		return parseStock(fields[1:])
	case "/관심":
		return parseMarketRanking(fields[1:], CommandInterest)
	case "/거래대금":
		return parseMarketRanking(fields[1:], CommandTradingValue)
	case "/수급상위":
		return parseMarketRanking(fields[1:], CommandFlowTop)
	case "/시간외", "/afterhours":
		return parseAfterHours(fields[1:])
	case "/신호", "/signal":
		return parseSignal(fields[1:])
	case "/오전실적", "/morning":
		return parseSessionPerformance(fields[1:], CommandMorningPerformance)
	case "/오후실적", "/afternoon":
		return parseSessionPerformance(fields[1:], CommandAfternoonPerformance)
	}
	return Command{Kind: CommandUnknown}
}

func isCommandKeyword(name string) bool {
	switch name {
	case "start", "시작",
		"help", "도움말",
		"pension", "연기금",
		"stock", "종목",
		"관심",
		"거래대금",
		"수급상위",
		"시간외", "afterhours",
		"신호", "signal",
		"오전실적", "morning",
		"오후실적", "afternoon":
		return true
	default:
		return false
	}
}

func parsePension(args []string) Command {
	if len(args) < 1 || len(args) > 2 {
		return Command{Kind: CommandUnknown}
	}
	period, ok := parsePeriod(args[0])
	if !ok {
		return Command{Kind: CommandUnknown}
	}
	limit := 10
	if len(args) == 2 {
		parsed, err := strconv.Atoi(args[1])
		if err != nil || parsed < 1 || parsed > 50 {
			return Command{Kind: CommandUnknown}
		}
		limit = parsed
	}
	return Command{Kind: CommandPension, Period: period, Limit: limit}
}

func parseStock(args []string) Command {
	if len(args) < 1 {
		return Command{Kind: CommandUnknown}
	}
	period := PeriodToday
	queryArgs := args
	if len(args) >= 2 {
		if parsed, ok := parsePeriod(args[len(args)-1]); ok {
			period = parsed
			queryArgs = args[:len(args)-1]
		}
	}
	query := strings.TrimSpace(strings.Join(queryArgs, " "))
	if query == "" {
		return Command{Kind: CommandUnknown}
	}
	return stockCommand(query, period)
}

func parseMarketRanking(args []string, kind CommandKind) Command {
	if len(args) < 1 || len(args) > 2 {
		return Command{Kind: CommandUnknown}
	}
	period, ok := parsePeriod(args[0])
	if !ok || period != PeriodToday {
		return Command{Kind: CommandUnknown}
	}
	limit := 10
	if len(args) == 2 {
		parsed, err := strconv.Atoi(args[1])
		if err != nil || parsed < 1 || parsed > 50 {
			return Command{Kind: CommandUnknown}
		}
		limit = parsed
	}
	return Command{Kind: kind, Period: period, Limit: limit}
}

func parseAfterHours(args []string) Command {
	if len(args) < 1 || len(args) > 2 {
		return Command{Kind: CommandUnknown}
	}

	limit := 10
	switch strings.ToLower(args[0]) {
	case "급등", "오늘", "up", "gainers":
	case "":
		return Command{Kind: CommandUnknown}
	default:
		parsed, err := strconv.Atoi(args[0])
		if err != nil || parsed < 1 || parsed > 50 || len(args) != 1 {
			return Command{Kind: CommandUnknown}
		}
		return Command{Kind: CommandAfterHours, Period: PeriodToday, Limit: parsed}
	}

	if len(args) == 2 {
		parsed, err := strconv.Atoi(args[1])
		if err != nil || parsed < 1 || parsed > 50 {
			return Command{Kind: CommandUnknown}
		}
		limit = parsed
	}
	return Command{Kind: CommandAfterHours, Period: PeriodToday, Limit: limit}
}

func parseSignal(args []string) Command {
	if len(args) < 1 {
		return Command{Kind: CommandUnknown}
	}
	if period, ok := parsePeriod(args[0]); ok {
		if len(args) > 2 {
			return Command{Kind: CommandUnknown}
		}
		if period != PeriodToday {
			return Command{Kind: CommandUnknown}
		}
		limit := 10
		if len(args) == 2 {
			parsed, err := strconv.Atoi(args[1])
			if err != nil || parsed < 1 || parsed > 50 {
				return Command{Kind: CommandUnknown}
			}
			limit = parsed
		}
		return Command{Kind: CommandSignal, Period: PeriodToday, Limit: limit}
	}
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return Command{Kind: CommandUnknown}
	}
	cmd := stockCommand(query, PeriodToday)
	cmd.Kind = CommandSignal
	return cmd
}

func parseSessionPerformance(args []string, kind CommandKind) Command {
	if len(args) < 1 || len(args) > 2 {
		return Command{Kind: CommandUnknown}
	}
	period, ok := parsePeriod(args[0])
	if !ok || period != PeriodToday {
		return Command{Kind: CommandUnknown}
	}
	limit := 10
	if len(args) == 2 {
		parsed, err := strconv.Atoi(args[1])
		if err != nil || parsed < 1 || parsed > 20 {
			return Command{Kind: CommandUnknown}
		}
		limit = parsed
	}
	return Command{Kind: kind, Period: PeriodToday, Limit: limit}
}

func stockCommand(query string, period Period) Command {
	if stockCodePattern.MatchString(query) {
		return Command{Kind: CommandStock, Code: query, Period: period}
	}
	return Command{Kind: CommandStock, Query: query, Period: period}
}

func parsePeriod(raw string) (Period, bool) {
	switch strings.ToLower(raw) {
	case "today", "오늘":
		return PeriodToday, true
	case "5d", "5일":
		return Period5D, true
	case "10d", "10일":
		return Period10D, true
	case "20d", "20일":
		return Period20D, true
	default:
		return "", false
	}
}
