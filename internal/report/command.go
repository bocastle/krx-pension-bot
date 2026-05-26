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

	switch name {
	case "/start":
		if len(fields) == 1 {
			return Command{Kind: CommandStart}
		}
	case "/help":
		if len(fields) == 1 {
			return Command{Kind: CommandHelp}
		}
	case "/pension":
		return parsePension(fields[1:])
	case "/stock":
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
	code := args[0]
	if !stockCodePattern.MatchString(code) {
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
	return Command{Kind: CommandStock, Code: code, Period: period}
}

func parsePeriod(raw string) (Period, bool) {
	switch strings.ToLower(raw) {
	case "today":
		return PeriodToday, true
	case "5d":
		return Period5D, true
	case "20d":
		return Period20D, true
	default:
		return "", false
	}
}
