package config

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"outblock.io/go-server/demo/models"
)

var activeindexModel = new(models.ActiveindexModel)

func Flow(network string) string {
	switch network {
	case "testnet":
		return "access.devnet.nodes.onflow.org:9000"
	case "mainnet":
		return "access.mainnet.nodes.onflow.org:9000"
	default:
		return "access.devnet.nodes.onflow.org:9000"
	}
}

func FlowKey(network string) (string, string, int) {
	switch network {
	case "testnet":
		return os.Getenv("TESTNET_ADDRESS"), os.Getenv("TESTNET_KEY"), 0
	case "mainnet":
		return os.Getenv("MAINNET_ADDRESS"), os.Getenv("MAINNET_KEY"), 0

	default:
		return os.Getenv("TESTNET_ADDRESS"), os.Getenv("TESTNET_KEY"), 0
	}
}

func GetKey(network string, c *gin.Context) (string, string, int64) {
	activeIndex, activeIndexErr := activeindexModel.SelectCustom("network", network)

	if activeIndexErr != nil {
		log.Println("active index error: ", activeIndexErr)
	}

	currentIndex := activeIndex.Index
	address, privateKey, _ := FlowKey(network)
	//return value through firestore

	maxIndex := activeIndex.MaxIndex

	activeIndex.Index = currentIndex + 1

	if currentIndex == maxIndex {
		activeIndex.Index = 1
	}
	_, updateErr := activeindexModel.Update(activeIndex)

	if updateErr != nil {
		log.Println(updateErr)
	}

	return address, privateKey, currentIndex
}

func Gdrive() string {
	return "https://www.googleapis.com/drive/v3/files/"
}
