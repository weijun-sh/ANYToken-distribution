package utils

import (
	"fmt"
	"time"

	"github.com/anyswap/ANYToken-distribution/callapi"
	"github.com/anyswap/ANYToken-distribution/log"
	"github.com/anyswap/ANYToken-distribution/mongodb"
	"github.com/anyswap/ANYToken-distribution/params"
	"github.com/fsn-dev/fsn-go-sdk/efsn/common"
	"github.com/urfave/cli/v2"
)

// InitApp init app (remember close client in the caller)
func InitApp(ctx *cli.Context, withConfigFile bool) *callapi.APICaller {
	return initApp(ctx, withConfigFile, "")
}

// InitAppWithURL init app for library use (remember close client in the caller)
func InitAppWithURL(ctx *cli.Context, serverURL string, withConfigFile bool) *callapi.APICaller {
	return initApp(ctx, withConfigFile, serverURL)
}

func initApp(ctx *cli.Context, withConfigFile bool, serverURL string) *callapi.APICaller {
	callapi.Init()
	SetLogger(ctx)

	if !withConfigFile {
		return DialServer(serverURL)
	}

	InitSyncArguments(ctx)

	configFile := GetConfigFilePath(ctx)
	params.LoadConfig(configFile)

	InitMongodb()

	if serverURL == "" {
		serverURL = params.GetConfig().Gateway.APIAddress
	}

	capi := DialServer(serverURL)

	if err := VerifyConfig(capi); err != nil {
		log.Fatalf("verifyConfig error. %v", err)
	}

	return capi
}

// DialServer connect to serverURL
func DialServer(serverURL string) *callapi.APICaller {
	capi := callapi.NewDefaultAPICaller()
	for {
		err := capi.DialServer(serverURL)
		if err == nil {
			break
		}
		time.Sleep(3 * time.Second)
	}
	return capi
}

// InitMongodb init mongodb by config
func InitMongodb() {
	config := params.GetConfig()
	dbConfig := config.MongoDB
	mongodb.MongoServerInit([]string{dbConfig.DBURL}, dbConfig.DBName, dbConfig.UserName, dbConfig.Password)
}

func VerifyConfig(capi *callapi.APICaller) error {
	if !params.CheckExchanges {
		return nil
	}
	if params.AnyswapV2 {
		return verifyConfigV2(capi)
	} else {
		return verifyConfig(capi)
	}
}

func verifyConfig(capi *callapi.APICaller) error {
	config := params.GetConfig()
	for _, ex := range config.Exchanges {
		exchange := common.HexToAddress(ex.Exchange)
		token := common.HexToAddress(ex.Token)
		wantToken := capi.GetExchangeTokenAddress(exchange)
		if token != wantToken {
			return fmt.Errorf("exchange token mismatch. exchange %v want token %v, but have %v", ex.Exchange, wantToken.String(), ex.Token)
		}
		factory := capi.GetExchangeFactoryAddress(exchange)
		if !params.IsConfigedFactory(factory) {
			return fmt.Errorf("exchange %v 's factory %v is not configed", ex.Exchange, factory.String())
		}
		log.Info("verify exchange token success", "exchange", ex.Exchange, "token", ex.Token, "factory", factory.String())
	}
	return nil
}

func verifyConfigV2(capi *callapi.APICaller) error {
	config := params.GetConfig()
	for _, ex := range config.Exchanges {
		exchange := common.HexToAddress(ex.Exchange)
		token0 := common.HexToAddress(ex.Token0)
		token1 := common.HexToAddress(ex.Token1)
		wantToken0 := capi.GetExchangeV2Token0Address(exchange)
		wantToken1 := capi.GetExchangeV2Token1Address(exchange)
		if token0 != wantToken0 && token1 != wantToken1 {
			return fmt.Errorf("exchange token mismatch. exchange %v want token0 %v, but have %v, want token0 %v, but have %v", ex.Exchange, wantToken0.String(), ex.Token0, wantToken1.String(), ex.Token1)
		}
		factory := capi.GetExchangeV2FactoryAddress(exchange)
		if !params.IsConfigedFactory(factory) {
			return fmt.Errorf("exchange %v 's factory %v is not configed", ex.Exchange, factory.String())
		}
		log.Info("verify exchange token success", "exchange", ex.Exchange, "token0", ex.Token0, "token1", ex.Token1, "factory", factory.String())
	}
	return nil
}
