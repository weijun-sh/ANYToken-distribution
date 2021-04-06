package callapi

import (
	"math/big"
	"time"

	"github.com/anyswap/ANYToken-distribution/log"
	"github.com/anyswap/ANYToken-distribution/params"
	ethereum "github.com/fsn-dev/fsn-go-sdk/efsn"
	"github.com/fsn-dev/fsn-go-sdk/efsn/common"
	"github.com/fsn-dev/fsn-go-sdk/efsn/core/types"
)

var (
	methodFactoryTokenCount []string = []string{"0x9f181b5e", "0x574f2ba3"}
	methodFactoryExchange   []string = []string{"0x06f2bf62", "0x1e3dd18b"}
	methodExchangeFactory   []string = []string{"0x966dae0e", "0xc45a0155"}
	methodExchangeToken0    string   = "0x0dfe1681"
	methodExchangeToken1    string   = "0xd21220a7"
)

var (
	method_FactoryTokenCount string
	method_FactoryExchange   string
	method_ExchangeToken0    string
	method_ExchangeToken1    string
	method_ExchangeFactory   string
)

func Init() {
	count := 0
	if params.AnyswapV2 {
		count = 1
	}
	method_FactoryTokenCount = methodFactoryTokenCount[count]
	method_FactoryExchange = methodFactoryExchange[count]
	method_ExchangeFactory = methodExchangeFactory[count]
	method_ExchangeToken0 = methodExchangeToken0
	method_ExchangeToken1 = methodExchangeToken0
}

// LoopGetBlockHeader loop get block header
func (c *APICaller) LoopGetBlockHeader(blockNumber *big.Int) *types.Header {
	for {
		header, err := c.client.HeaderByNumber(c.context, blockNumber)
		if err == nil {
			return header
		}
		log.Error("[callapi] get block header failed.", "blockNumber", blockNumber, "err", err)
		time.Sleep(c.rpcRetryInterval)
	}
}

// LoopGetLatestBlockHeader loop get latest block header
func (c *APICaller) LoopGetLatestBlockHeader() *types.Header {
	for {
		header, err := c.client.HeaderByNumber(c.context, nil)
		if err == nil {
			log.Info("[callapi] get latest block header succeed.",
				"number", header.Number,
				"hash", header.Hash().String(),
				"timestamp", header.Time,
			)
			return header
		}
		log.Error("[callapi] get latest block header failed.", "err", err)
		time.Sleep(c.rpcRetryInterval)
	}
}

// LoopGetExchangeLiquidity get exchange liquidity
func (c *APICaller) LoopGetExchangeLiquidity(exchangeAddr common.Address, blockNumber *big.Int) *big.Int {
	return c.LoopGetTokenTotalSupply(exchangeAddr, blockNumber)
}

// LoopGetTokenTotalSupply get token total supply
func (c *APICaller) LoopGetTokenTotalSupply(address common.Address, blockNumber *big.Int) *big.Int {
	var totalSupply *big.Int
	var err error
	for {
		totalSupply, err = c.GetTokenTotalSupply(address, blockNumber)
		if err == nil {
			break
		}
		log.Error("[callapi] GetTokenTotalSupply error", "address", address.String(), "err", err)
		time.Sleep(c.rpcRetryInterval)
	}
	return totalSupply
}

// LoopGetCoinBalance get coin balance
func (c *APICaller) LoopGetCoinBalance(address common.Address, blockNumber *big.Int) *big.Int {
	var fsnBalance *big.Int
	var err error
	for {
		fsnBalance, err = c.GetCoinBalance(address, blockNumber)
		if err == nil {
			break
		}
		log.Error("[callapi] GetCoinBalance error", "address", address.String(), "err", err)
		time.Sleep(c.rpcRetryInterval)
	}
	return fsnBalance
}

// LoopGetExchangeTokenBalance get account token balance
func (c *APICaller) LoopGetExchangeTokenBalance(exchange, token common.Address, blockNumber *big.Int) *big.Int {
	return c.LoopGetTokenBalance(token, exchange, blockNumber)
}

// LoopGetLiquidityBalance get account token balance
func (c *APICaller) LoopGetLiquidityBalance(exchange, account common.Address, blockNumber *big.Int) *big.Int {
	return c.LoopGetTokenBalance(exchange, account, blockNumber)
}

// LoopGetTokenBalance get account token balance
func (c *APICaller) LoopGetTokenBalance(tokenAddr, account common.Address, blockNumber *big.Int) *big.Int {
	var tokenBalance *big.Int
	var err error
	for {
		tokenBalance, err = c.GetTokenBalance(tokenAddr, account, blockNumber)
		if err == nil {
			break
		}
		log.Error("[callapi] GetTokenBalance error", "token", tokenAddr.String(), "account", account.String(), "err", err)
		time.Sleep(c.rpcRetryInterval)
	}
	return tokenBalance
}

// LoopGetFactoryExchange get factory exchange
func (c *APICaller) LoopGetFactoryExchange(factory, tokenAddr common.Address) common.Address {
	return c.loopGetFactoryExcahngeOrToken(factory, tokenAddr, true)
}

// LoopGetFactoryToken get factory token
func (c *APICaller) LoopGetFactoryToken(factory, exchange common.Address) common.Address {
	return c.loopGetFactoryExcahngeOrToken(factory, exchange, false)
}

func (c *APICaller) loopGetFactoryExcahngeOrToken(factory, address common.Address, isGetExchange bool) common.Address {
	var (
		res []byte
		err error

		getExchangeFuncHash = common.FromHex(method_FactoryExchange)
		getTokenFuncHash    = common.FromHex("0x59770438")
	)
	data := make([]byte, 36)
	if isGetExchange {
		copy(data[:4], getExchangeFuncHash)
	} else {
		copy(data[:4], getTokenFuncHash)
	}
	copy(data[4:], address.Hash().Bytes())
	msg := ethereum.CallMsg{
		To:   &factory,
		Data: data,
	}
	for {
		res, err = c.client.CallContract(c.context, msg, nil)
		if err == nil {
			break
		}
		log.Error("[callapi] GetFactoryExcahngeOrToken error", "factory", factory.String(), "address", address.String(), "isGetExchange", isGetExchange, "err", err)
		time.Sleep(c.rpcRetryInterval)
	}
	return common.BytesToAddress(common.GetData(res, 0, 32))
}

// LoopGetFactoryExchangeV2 get factory exchange
func (c *APICaller) LoopGetFactoryExchangeV2(factory common.Address, pairsIndex uint64) common.Address {
	return c.loopGetFactoryExchangeV2(factory, pairsIndex)
}

func (c *APICaller) loopGetFactoryExchangeV2(factory common.Address, pairsIndex uint64) common.Address {
	var (
		res []byte
		err error

		getExchangeFuncHash = common.FromHex(method_FactoryExchange)
	)
	data := make([]byte, 36)
	copy(data[:4], getExchangeFuncHash)
	copy(data[4:], common.BigToHash(new(big.Int).SetUint64(pairsIndex)).Bytes())
	msg := ethereum.CallMsg{
		To:   &factory,
		Data: data,
	}
	for {
		res, err = c.client.CallContract(c.context, msg, nil)
		if err == nil {
			break
		}
		log.Error("[callapi] GetFactoryExcahngeV2 error", "factory", factory.String(), "pairsIndex", pairsIndex, "err", err)
		time.Sleep(c.rpcRetryInterval)
	}
	return common.BytesToAddress(common.GetData(res, 0, 32))
}

// LoopGetFactoryTokenCount get token count of factory
func (c *APICaller) LoopGetFactoryTokenCount(factory common.Address) uint64 {
	var (
		res []byte
		err error

		getTokenCountFuncHash = common.FromHex(method_FactoryTokenCount)
	)
	msg := ethereum.CallMsg{
		To:   &factory,
		Data: getTokenCountFuncHash,
	}
	for i := 0; i < c.rpcRetryCount; i++ {
		res, err = c.client.CallContract(c.context, msg, nil)
		if err == nil {
			break
		}
		log.Error("[callapi] GetFactoryTokenCount error", "factory", factory.String(), "err", err)
		time.Sleep(c.rpcRetryInterval)
	}
	return new(big.Int).SetBytes(common.GetData(res, 0, 32)).Uint64()
}

// LoopGetFactoryTokenWithID get token with id
func (c *APICaller) LoopGetFactoryTokenWithID(factory common.Address, id uint64) common.Address {
	var (
		res []byte
		err error

		getTokenWithIDFuncHash = common.FromHex("0xaa65a6c0")
	)
	data := make([]byte, 36)
	copy(data[:4], getTokenWithIDFuncHash)
	copy(data[4:], common.BigToHash(new(big.Int).SetUint64(id)).Bytes())
	msg := ethereum.CallMsg{
		To:   &factory,
		Data: data,
	}
	for i := 0; i < c.rpcRetryCount; i++ {
		res, err = c.client.CallContract(c.context, msg, nil)
		if err == nil {
			break
		}
		log.Error("[callapi] GetFactoryTokenWithID error", "factory", factory.String(), "id", id, "err", err)
		time.Sleep(c.rpcRetryInterval)
	}
	return common.BytesToAddress(common.GetData(res, 0, 32))
}
