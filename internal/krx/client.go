package krx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bocastle/krx-pension-bot/internal/cache"
	"github.com/bocastle/krx-pension-bot/internal/report"
)

const (
	jsonPath             = "/comm/bldAttendant/getJsonData.cmd"
	naverFrontAPIPath    = "/front-api/domestic/stock/list"
	marketTickerBLD      = "dbms/MDC_OUT/EASY/ranking/MDCEASY01601_OUT"
	netBuyByInvestorBLD  = "dbms/MDC_OUT/STAT/standard/MDCSTAT02401_OUT"
	naverStockBaseURL    = "https://m.stock.naver.com"
	pensionInvestorCode  = "6000"
	defaultClientTimeout = 10 * time.Second
	afterHoursMaxPages   = 5
	afterHoursPageSize   = 50
)

type Client struct {
	baseURL     string
	naverURL    string
	http        *http.Client
	flowCache   *cache.Cache[string, []report.Flow]
	tickerCache *cache.Cache[string, []report.Ticker]
	afterCache  *cache.Cache[string, []report.AfterHoursStock]
}

func NewClient(baseURL string, ttl time.Duration) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL:     baseURL,
		naverURL:    defaultNaverURL(baseURL),
		http:        &http.Client{Timeout: defaultClientTimeout},
		flowCache:   cache.New[string, []report.Flow](ttl, time.Now),
		tickerCache: cache.New[string, []report.Ticker](ttl, time.Now),
		afterCache:  cache.New[string, []report.AfterHoursStock](ttl, time.Now),
	}
}

func (c *Client) MarketFlows(ctx context.Context, market report.Market, period report.QueryPeriod) ([]report.Flow, error) {
	key := fmt.Sprintf("%s:%s:%s", market, period.Start.Format("20060102"), period.End.Format("20060102"))
	if rows, ok := c.flowCache.Get(key); ok {
		return rows, nil
	}

	form := url.Values{}
	form.Set("bld", netBuyByInvestorBLD)
	form.Set("strtDd", period.Start.Format("20060102"))
	form.Set("endDd", period.End.Format("20060102"))
	form.Set("mktId", marketCode(market))
	form.Set("invstTpCd", pensionInvestorCode)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+jsonPath, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Referer", "https://data.krx.co.kr/contents/MDC/MDI/mdiLoader/index.cmd?menuId=MDC0201020303")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("krx request failed: status %d", resp.StatusCode)
	}

	var payload map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	rows := parseRowsPayload(payload["output"])
	if len(rows) == 0 {
		rows = parseRowsPayload(payload["block1"])
	}
	c.flowCache.Set(key, rows)
	return rows, nil
}

func (c *Client) MarketTickers(ctx context.Context, market report.Market, date time.Time) ([]report.Ticker, error) {
	key := fmt.Sprintf("ticker:%s:%s", market, date.Format("20060102"))
	if rows, ok := c.tickerCache.Get(key); ok {
		return rows, nil
	}

	form := url.Values{}
	form.Set("bld", marketTickerBLD)
	form.Set("locale", "ko_KR")
	form.Set("mktId", marketCode(market))
	if market == report.MarketKOSDAQ {
		form.Set("segTpCd", "ALL")
	}
	form.Set("itmTpCd1", "N")
	form.Set("itmTpCd2", "1")
	form.Set("itmTpCd3", "2")
	form.Set("stkprcTpCd", "Y")
	form.Set("strtDd", date.Format("20060102"))
	form.Set("endDd", date.Format("20060102"))
	form.Set("share", "1")
	form.Set("money", "1")
	form.Set("csvxls_isNo", "false")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+jsonPath, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Referer", "https://data.krx.co.kr/contents/MDC/MDI/outerLoader/index.cmd?screenId=MDCEASY016&locale=ko_KR")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("krx request failed: status %d", resp.StatusCode)
	}

	var payload map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	rows := parseTickerRowsPayload(payload["output"])
	if len(rows) == 0 {
		rows = parseTickerRowsPayload(payload["block1"])
	}
	if len(rows) == 0 {
		rows = parseTickerRowsPayload(payload["OutBlock_1"])
	}
	c.tickerCache.Set(key, rows)
	return rows, nil
}

func (c *Client) AfterHoursGainers(ctx context.Context, market report.Market) ([]report.AfterHoursStock, error) {
	key := fmt.Sprintf("after-hours:%s", market)
	if rows, ok := c.afterCache.Get(key); ok {
		return rows, nil
	}

	rowsByCode := make(map[string]report.AfterHoursStock)
	for page := 1; page <= afterHoursMaxPages; page++ {
		values := url.Values{}
		values.Set("sortType", "up")
		values.Set("category", afterHoursCategory(market))
		values.Set("domesticStockExchangeType", "KRX")
		values.Set("page", strconv.Itoa(page))
		values.Set("pageSize", strconv.Itoa(afterHoursPageSize))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.naverURL+naverFrontAPIPath+"?"+values.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Referer", "https://m.stock.naver.com/domestic/home/upper/"+afterHoursCategory(market))
		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			resp.Body.Close()
			return nil, fmt.Errorf("naver stock request failed: status %d", resp.StatusCode)
		}

		var payload naverAfterHoursPayload
		err = json.NewDecoder(resp.Body).Decode(&payload)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		rows := parseAfterHoursRows(payload.Result.Stocks)
		for _, row := range rows {
			rowsByCode[row.Code] = row
		}
		if len(payload.Result.Stocks) < afterHoursPageSize {
			break
		}
	}

	rows := make([]report.AfterHoursStock, 0, len(rowsByCode))
	for _, row := range rowsByCode {
		rows = append(rows, row)
	}
	c.afterCache.Set(key, rows)
	return rows, nil
}

type naverAfterHoursPayload struct {
	Result struct {
		Stocks []naverAfterHoursRow `json:"stocks"`
	} `json:"result"`
}

type naverAfterHoursRow struct {
	Code                string                    `json:"itemCode"`
	Name                string                    `json:"name"`
	OverMarketPriceInfo *naverOverMarketPriceInfo `json:"overMarketPriceInfo"`
}

type naverOverMarketPriceInfo struct {
	TradingSessionType string `json:"tradingSessionType"`
	OverPrice          string `json:"overPrice"`
	Fluctuations       string `json:"fluctuations"`
	FluctuationsRatio  string `json:"fluctuationsRatio"`
}

func marketCode(market report.Market) string {
	switch market {
	case report.MarketKOSDAQ:
		return "KSQ"
	default:
		return "STK"
	}
}

func afterHoursCategory(market report.Market) string {
	switch market {
	case report.MarketKOSDAQ:
		return "KOSDAQ"
	default:
		return "KOSPI"
	}
}

func defaultNaverURL(baseURL string) string {
	if strings.Contains(baseURL, "data.krx.co.kr") {
		return naverStockBaseURL
	}
	return baseURL
}

func parseRowsPayload(raw json.RawMessage) []report.Flow {
	if len(raw) == 0 || string(raw) == `""` {
		return nil
	}

	var rows []map[string]string
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil
	}
	return parseRows(rows)
}

func parseTickerRowsPayload(raw json.RawMessage) []report.Ticker {
	if len(raw) == 0 || string(raw) == `""` {
		return nil
	}

	var rows []map[string]string
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil
	}
	return parseTickerRows(rows)
}

func parseRows(raw []map[string]string) []report.Flow {
	rows := make([]report.Flow, 0, len(raw))
	for _, item := range raw {
		row := report.Flow{
			Code:       pick(item, "ISU_SRT_CD", "ISU_CD", "isuSrtCd"),
			Name:       pick(item, "ISU_ABBRV", "ISU_NM", "isuAbbrv"),
			SellVolume: parseNumber(pick(item, "ASK_TRDVOL", "ASK_TRD_VOL")),
			BuyVolume:  parseNumber(pick(item, "BID_TRDVOL", "BID_TRD_VOL")),
			NetVolume:  parseNumber(pick(item, "NETBID_TRDVOL", "NETBID_TRD_VOL")),
			SellValue:  parseNumber(pick(item, "ASK_TRDVAL", "ASK_TRD_VAL")),
			BuyValue:   parseNumber(pick(item, "BID_TRDVAL", "BID_TRD_VAL")),
			NetValue:   parseNumber(pick(item, "NETBID_TRDVAL", "NETBID_TRD_VAL")),
		}
		if row.Code != "" || row.Name != "" {
			rows = append(rows, row)
		}
	}
	return rows
}

func parseTickerRows(raw []map[string]string) []report.Ticker {
	rows := make([]report.Ticker, 0, len(raw))
	for _, item := range raw {
		row := report.Ticker{
			Code:         pick(item, "ISU_SRT_CD", "ISU_CD", "isuSrtCd"),
			Name:         pick(item, "ISU_ABBRV", "ISU_NM", "isuAbbrv"),
			TradeVolume:  parseNumber(pick(item, "ACC_TRDVOL", "ACC_TRD_VOL")),
			TradeValue:   parseNumber(pick(item, "ACC_TRDVAL", "ACC_TRD_VAL")),
			ChangeRate:   parseDecimal(pick(item, "FLUC_RT", "FLUCT_RT")),
			ClosingPrice: parseNumber(pick(item, "TDD_CLSPRC", "CLSPRC")),
		}
		if row.Code != "" || row.Name != "" {
			rows = append(rows, row)
		}
	}
	return rows
}

func parseAfterHoursRows(raw []naverAfterHoursRow) []report.AfterHoursStock {
	rows := make([]report.AfterHoursStock, 0, len(raw))
	for _, item := range raw {
		if item.OverMarketPriceInfo == nil || item.OverMarketPriceInfo.TradingSessionType != "AFTER_MARKET" {
			continue
		}
		row := report.AfterHoursStock{
			Code:            strings.TrimSpace(item.Code),
			Name:            strings.TrimSpace(item.Name),
			AfterPrice:      parseNumber(item.OverMarketPriceInfo.OverPrice),
			AfterChange:     parseNumber(item.OverMarketPriceInfo.Fluctuations),
			AfterChangeRate: parseDecimal(item.OverMarketPriceInfo.FluctuationsRatio),
		}
		if row.Code != "" || row.Name != "" {
			rows = append(rows, row)
		}
	}
	return rows
}

func pick(row map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(row[key]); value != "" {
			return value
		}
	}
	return ""
}

func parseNumber(raw string) int64 {
	cleaned := strings.NewReplacer(",", "", "+", "", " ", "").Replace(strings.TrimSpace(raw))
	if cleaned == "" || cleaned == "-" {
		return 0
	}
	value, err := strconv.ParseInt(cleaned, 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func parseDecimal(raw string) float64 {
	cleaned := strings.NewReplacer(",", "", "+", "", "%", "", " ", "").Replace(strings.TrimSpace(raw))
	if cleaned == "" || cleaned == "-" {
		return 0
	}
	value, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	return value
}
