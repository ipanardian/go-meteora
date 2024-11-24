package meteora

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
)

type Meteora struct {
	BaseUrl string
	pool    sync.Pool
}

type MeteoraRequest struct {
	Data        *string
	QueryParams *string
	Headers     map[string]string
}

type MeteoraQuoteRequest struct {
	PairID string
}

type MeteoraData struct {
	Address               string  `json:"address"`
	Name                  string  `json:"name"`
	MintX                 string  `json:"mint_x"`
	MintY                 string  `json:"mint_y"`
	ReserveX              string  `json:"reserve_x"`
	ReserveY              string  `json:"reserve_y"`
	ReserveXAmount        int64   `json:"reserve_x_amount"`
	ReserveYAmount        int64   `json:"reserve_y_amount"`
	BinStep               int     `json:"bin_step"`
	BaseFeePercentage     string  `json:"base_fee_percentage"`
	MaxFeePercentage      string  `json:"max_fee_percentage"`
	ProtocolFeePercentage string  `json:"protocol_fee_percentage"`
	Liquidity             string  `json:"liquidity"`
	RewardMintX           string  `json:"reward_mint_x"`
	RewardMintY           string  `json:"reward_mint_y"`
	Fees24H               float64 `json:"fees_24h"`
	TodayFees             float64 `json:"today_fees"`
	TradeVolume24H        float64 `json:"trade_volume_24h"`
	CumulativeTradeVolume string  `json:"cumulative_trade_volume"`
	CumulativeFeeVolume   string  `json:"cumulative_fee_volume"`
	CurrentPrice          float64 `json:"current_price"`
	Apr                   float64 `json:"apr"`
	Apy                   float64 `json:"apy"`
	FarmApr               float64 `json:"farm_apr"`
	FarmApy               float64 `json:"farm_apy"`
	Hide                  bool    `json:"hide"`
}

func NewMeteora(baseUrl string) *Meteora {
	return &Meteora{
		BaseUrl: baseUrl,
		pool: sync.Pool{
			New: func() interface{} {
				timeout := 5000 * time.Millisecond
				return httpclient.NewClient(
					httpclient.WithHTTPTimeout(timeout),
				)
			},
		},
	}
}

func (m *Meteora) Get(res any, path string, data MeteoraRequest) error {
	return m.getAndUnmarshalJson(res, path, data)
}

func (m *Meteora) Post(res any, path string, data MeteoraRequest) error {
	return m.postAndUnmarshalJson(res, path, data)
}

func (m *Meteora) SwapQuote(res any, headers map[string]string, data MeteoraQuoteRequest) error {
	headers["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
	headers["Cache-Content"] = "no-cache"
	return m.getAndUnmarshalJson(res, "/clmm-api/pair/", MeteoraRequest{
		Headers: headers,
		Data:    &data.PairID,
	})
}

func (m *Meteora) getAndUnmarshalJson(res any, path string, data MeteoraRequest) error {
	client := m.pool.Get().(*httpclient.Client)
	url := fmt.Sprintf("%s%s%s", m.BaseUrl, path, *data.Data)

	if data.QueryParams != nil {
		url = url + "?" + *data.QueryParams
	}

	jB, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dataReader := strings.NewReader("")
	if data.Data != nil {
		dataReader = strings.NewReader(string(jB))
	}

	req, err := http.NewRequest(http.MethodGet, url, dataReader)
	if err != nil {
		return err
	}

	if data.Headers != nil {
		for key, value := range data.Headers {
			req.Header.Set(key, value)
		}
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	return nil
}

func (m *Meteora) postAndUnmarshalJson(res any, path string, data MeteoraRequest) error {
	client := m.pool.Get().(*httpclient.Client)
	url := fmt.Sprintf("%s%s", m.BaseUrl, path)

	if data.QueryParams != nil {
		url = url + "?" + *data.QueryParams
	}

	jB, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dataReader := strings.NewReader("")
	if data.Data != nil {
		dataReader = strings.NewReader(string(jB))
	}

	req, err := http.NewRequest(http.MethodPost, url, dataReader)
	if err != nil {
		return err
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("Cache-Content", "no-cache")
	if data.Headers != nil {
		for key, value := range data.Headers {
			req.Header.Set(key, value)
		}
	}

	rcv := getPointer(res)

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer rsp.Body.Close()

	if err := json.NewDecoder(rsp.Body).Decode(rcv); err != nil {
		return err
	}

	return nil
}

func getPointer(v interface{}) interface{} {
	vv := valueOf(v)
	if vv.Kind() == reflect.Ptr {
		return v
	}
	return reflect.New(vv.Type()).Interface()
}

func valueOf(i interface{}) reflect.Value {
	return reflect.ValueOf(i)
}
