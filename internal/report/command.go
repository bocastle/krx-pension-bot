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
	}
	return Command{Kind: CommandUnknown}
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
	if len(args) < 1 || len(args) > 2 {
		return Command{Kind: CommandUnknown}
	}
	query := strings.TrimSpace(args[0])
	if query == "" {
		return Command{Kind: CommandUnknown}
	}
	period := PeriodToday
	if len(args) == 2 {
		parsed, ok := parsePeriod(args[1])
		if !ok {
			return Command{Kind: CommandUnknown}
		}
		period = parsed
	}
	return stockCommand(query, period)
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
	case "20d", "20일":
		return Period20D, true
	default:
		return "", false
	}
}
