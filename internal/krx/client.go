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
	netBuyByInvestorBLD  = "dbms/MDC_OUT/STAT/standard/MDCSTAT02401_OUT"
	pensionInvestorCode  = "6000"
	defaultClientTimeout = 10 * time.Second
)

type Client struct {
	baseURL string
	http    *http.Client
	cache   *cache.Cache[string, []report.Flow]
}

func NewClient(baseURL string, ttl time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: defaultClientTimeout},
		cache:   cache.New[string, []report.Flow](ttl, time.Now),
	}
}

func (c *Client) MarketFlows(ctx context.Context, market report.Market, period report.QueryPeriod) ([]report.Flow, error) {
	key := fmt.Sprintf("%s:%s:%s", market, period.Start.Format("20060102"), period.End.Format("20060102"))
	if rows, ok := c.cache.Get(key); ok {
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
	c.cache.Set(key, rows)
	return rows, nil
}

func marketCode(market report.Market) string {
	switch market {
	case report.MarketKOSDAQ:
		return "KSQ"
	default:
		return "STK"
	}
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
