package handlers

import (
	"hw/pkg/ethindexa"
	"hw/pkg/logger"
)

// var chainlinkETHUSDAddress = common.HexToAddress("0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419")

func HandleTransfer(idx *ethindexa.IndexerService, event ethindexa.Event) {
	logger.Infof("#%s:%s:%s %+v %v", event.NetworkName, event.ContractName, event.EventName, event.ContractAddress, event.Args)

	// trt to call chainlink contract to get eth price
	// block, err := idx.GetBlockByHash(event.BlockHash)
	// if err != nil {
	// 	logger.Infof("GetBlockByHash error: %+v", err)
	// 	return
	// }
	// abi, err := utils.LoadABI("chainlink")
	// if err != nil {
	// 	logger.Infof("LoadABI error: %+v", err)
	// 	return
	// }
	// answer, err := utils.ReadContract(idx.Client, chainlinkETHUSDAddress, abi, big.NewInt(block.Number().Int64()), "latestAnswer")
	// if err != nil {
	// 	logger.Infof("ReadContract error: %+v", err)
	// 	return
	// }
	// logger.Infof("latestAnswer: %+v", answer)
}

func HandleApproval(idx *ethindexa.IndexerService, event ethindexa.Event) {
	logger.Infof("#%s:%s:%s %+v %v", event.NetworkName, event.ContractName, event.EventName, event.ContractAddress, event.Args)
}
