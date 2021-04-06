package syncer

import (
	"fmt"
	"github.com/anyswap/ANYToken-distribution/log"
	"github.com/anyswap/ANYToken-distribution/params"
	"github.com/fsn-dev/fsn-go-sdk/efsn/common"
	"github.com/fsn-dev/fsn-go-sdk/efsn/core/types"
)

func addExchanges(rlog *types.Log) {
	topics := rlog.Topics
	if len(topics) != 2 {
		return
	}
	token := common.BytesToAddress(topics[1].Bytes())
	exchange := common.BytesToAddress(topics[2].Bytes())
	params.AddTokenAndExchange(token, exchange)
}

func InitAllExchanges() {
	log.Info("InitAllExchanges", "params.CheckExchanges", params.CheckExchanges)
	if params.GetExchanges {
		if params.AnyswapV2 {
			initAllExchangesV2()
		} else {
			initAllExchanges()
		}
	}
}

func initAllExchanges() {
	for _, factory := range params.GetFactories() {
		initExchangesInFactory(factory)
	}
}

func initExchangesInFactory(factory common.Address) {
	tokenCount := capi.LoopGetFactoryTokenCount(factory)
	for i := uint64(1); i <= tokenCount; i++ {
		token := capi.LoopGetFactoryTokenWithID(factory, i)
		exchange := capi.LoopGetFactoryExchange(factory, token)
		params.AddTokenAndExchange(token, exchange)
	}
	log.Info("initExchangesInFactory success", "factory", factory.String(), "tokenCount", tokenCount, "added", len(params.AllExchanges))
}

func addExchangesV2(rlog *types.Log) {
	topics := rlog.Topics
	if len(topics) != 3 {
		return
	}
	token0 := common.BytesToAddress(topics[1].Bytes())
	token1 := common.BytesToAddress(topics[2].Bytes())
	data := rlog.Data
	exchange := common.BytesToAddress(data[:32])
	//index := data[32:]
	pair, err := getExchangeV2Symbol(token0, token1)
	if err != nil {
		return
	}
	params.AddTokenAndExchangeV2(token0, token1, exchange, pair)
}

func initAllExchangesV2() {
	for _, factory := range params.GetFactories() {
		initExchangesV2InFactory(factory)
	}
}

func initExchangesV2InFactory(factory common.Address) {
	tokenCount := capi.LoopGetFactoryTokenCount(factory)
	for i := uint64(0); i <= tokenCount; i++ {
		exchange := capi.LoopGetFactoryExchangeV2(factory, i)
		token0 := capi.GetExchangeV2Token0Address(exchange)
		token1 := capi.GetExchangeV2Token1Address(exchange)
		params.AddTokenAndExchange(token0, exchange)
		params.AddTokenAndExchange(token1, exchange)
		pair, err := getExchangeV2Symbol(token0, token1)
		if err != nil {
			log.Info("initExchangesV2InFactory", "pairsIndex", i, "pairsAll", tokenCount, "exchange", exchange.String(), "token0", token0.String(), "token1", token1.String(), "GetErc20Symbol:err", err)
			continue
		}
		params.AddTokenAndExchangeV2(token0, token1, exchange, pair)
		printConfigExchanges(i+1, token0, token1, exchange, pair)
	}
	log.Info("initExchangesInFactory success", "factory", factory.String(), "tokenCount", tokenCount, "added", len(params.AllExchanges))
}

func printConfigExchanges(index uint64, token0, token1, exchange common.Address, pair string) {
	fmt.Printf("[[Exchanges]]\n")
	fmt.Printf("Pairs = \"%v\"\n", pair)
	fmt.Printf("Exchange = \"%v\"\n", exchange.String())
	fmt.Printf("Token0 = \"%v\"\n", token0.String())
	fmt.Printf("Token1 = \"%v\"\n", token1.String())
	fmt.Printf("CreationHeight = %v\n", index)
	fmt.Printf("LiquidWeight = 0\n")
	fmt.Printf("TradeWeight = 0\n\n")
}

func getExchangeV2Symbol(token0, token1 common.Address) (string, error) {
	pair0, err := capi.GetErc20Symbol(token0)
	if err != nil {
		log.Info("getExchangeV2Symbol", "token0", token0.String(), "GetErc20Symbol:err", err)
		return "", err
	}
	pair1, err := capi.GetErc20Symbol(token1)
	if err != nil {
		log.Info("[initExchangesV2InFactory]", "token1", token1.String(), "GetErc20Symbol:err", err)
		return "", err
	}
	return pair0 + "/" + pair1, nil
}
