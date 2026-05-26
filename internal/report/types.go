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
	Period20D   Period = "20d"
)

type CommandKind int

const (
	CommandUnknown CommandKind = iota
	CommandStart
	CommandHelp
	CommandPension
	CommandStock
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

type QueryPeriod struct {
	Start time.Time
	End   time.Time
	Label string
}

type Source interface {
	MarketFlows(ctx context.Context, market Market, period QueryPeriod) ([]Flow, error)
}
