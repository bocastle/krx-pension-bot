package report

import (
	"context"
	"time"
)

type Market string

const (
	MarketKOSPI  Market = "KOSPI"
	MarketKOSDAQ Market = "KOSDAQ"
)

type Period string

const (
	PeriodToday Period = "today"
	Period5D    Period = "5d"
	Period10D   Period = "10d"
	Period20D   Period = "20d"
)

type CommandKind int

const (
	CommandUnknown CommandKind = iota
	CommandStart
	CommandHelp
	CommandPension
	CommandStock
	CommandInterest
	CommandTradingValue
	CommandFlowTop
	CommandAfterHours
	CommandSignal
	CommandMorningPerformance
	CommandAfternoonPerformance
)

type Command struct {
	Kind   CommandKind
	Period Period
	Code   string
	Query  string
	Limit  int
}

type Flow struct {
	Code       string
	Name       string
	SellVolume int64
	BuyVolume  int64
	NetVolume  int64
	SellValue  int64
	BuyValue   int64
	NetValue   int64
}

type Ticker struct {
	Code         string
	Name         string
	TradeVolume  int64
	TradeValue   int64
	ChangeRate   float64
	ClosingPrice int64
}

type AfterHoursStock struct {
	Code            string
	Name            string
	AfterPrice      int64
	AfterChange     int64
	AfterChangeRate float64
}

type IntradayPrice struct {
	Time   time.Time
	Close  int64
	Volume int64
}

type Session string

const (
	SessionMorning   Session = "morning"
	SessionAfternoon Session = "afternoon"
)

type QueryPeriod struct {
	Start time.Time
	End   time.Time
	Label string
}

type Source interface {
	MarketFlows(ctx context.Context, market Market, period QueryPeriod) ([]Flow, error)
	MarketTickers(ctx context.Context, market Market, date time.Time) ([]Ticker, error)
	AfterHoursGainers(ctx context.Context, market Market) ([]AfterHoursStock, error)
	IntradayPrices(ctx context.Context, code string, date time.Time) ([]IntradayPrice, error)
}
