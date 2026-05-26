package report

import "testing"

func TestParseCommand(t *testing.T) {
	tests := []struct {
		text string
		want Command
	}{
		{"/start", Command{Kind: CommandStart}},
		{"/시작", Command{Kind: CommandStart}},
		{"/help", Command{Kind: CommandHelp}},
		{"/도움말", Command{Kind: CommandHelp}},
		{"/pension today", Command{Kind: CommandPension, Period: PeriodToday, Limit: 10}},
		{"/pension@MyBot 5d", Command{Kind: CommandPension, Period: Period5D, Limit: 10}},
		{"/pension 10d", Command{Kind: CommandPension, Period: Period10D, Limit: 10}},
		{"/pension today 20", Command{Kind: CommandPension, Period: PeriodToday, Limit: 20}},
		{"/연기금 오늘", Command{Kind: CommandPension, Period: PeriodToday, Limit: 10}},
		{"/연기금 5일", Command{Kind: CommandPension, Period: Period5D, Limit: 10}},
		{"/연기금 10일", Command{Kind: CommandPension, Period: Period10D, Limit: 10}},
		{"/연기금 20일 20", Command{Kind: CommandPension, Period: Period20D, Limit: 20}},
		{"/stock 005930", Command{Kind: CommandStock, Code: "005930", Period: PeriodToday}},
		{"/stock 005930 20d", Command{Kind: CommandStock, Code: "005930", Period: Period20D}},
		{"/종목 005930", Command{Kind: CommandStock, Code: "005930", Period: PeriodToday}},
		{"/종목 005930 10일", Command{Kind: CommandStock, Code: "005930", Period: Period10D}},
		{"/종목 005930 20일", Command{Kind: CommandStock, Code: "005930", Period: Period20D}},
		{"/종목 삼성전자", Command{Kind: CommandStock, Query: "삼성전자", Period: PeriodToday}},
		{"/종목 삼성전자 20일", Command{Kind: CommandStock, Query: "삼성전자", Period: Period20D}},
		{"삼성전자", Command{Kind: CommandStock, Query: "삼성전자", Period: PeriodToday}},
	}

	for _, tt := range tests {
		got := ParseCommand(tt.text)
		if got != tt.want {
			t.Fatalf("ParseCommand(%q) = %#v, want %#v", tt.text, got, tt.want)
		}
	}
}

func TestParseCommandRejectsUnknownInput(t *testing.T) {
	got := ParseCommand("/pension yesterday")
	if got.Kind != CommandUnknown {
		t.Fatalf("Kind = %v, want CommandUnknown", got.Kind)
	}
}
