# Go Meteora SDK

## Usage
```go
import "github.com/ipanardian/go-meteora"

var res meteora.MeteoraData
cl := meteora.NewMeteoraAPI(MeteoraURI)
headers := make(map[string]string)
err = cl.SwapQuote(&res, headers, meteora.MeteoraQuoteRequest{
    PairID: {{pair_id}}
})
```