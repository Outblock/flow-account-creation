package helper

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/onflow/flow-go-sdk"
	"outblock.io/go-server/demo/config"
	"outblock.io/go-server/demo/forms"
	"outblock.io/go-server/demo/models"
)

type AccountHelper struct{}

var walletModelPreview = new(models.WalletModelPreview)

type CreateAddressRequest struct {
	Network    string            `json:"network"`
	AccountKey forms.AccountForm `json:"account_key"`
}

func (ctrl AccountHelper) CreatePreviewAccount(c *gin.Context, Network string, AccountKey forms.AccountForm) (string, error) {

	walletId, createErr := walletModelPreview.CreateWallet(AccountKey)
	if createErr != nil {
		log.Printf("Error while creating a previewnet wallet, Reason: %v\n", createErr)
		return "", createErr
	}

	wallet, selectWalletErr := walletModelPreview.Select(walletId)
	if selectWalletErr != nil {
		log.Printf("Error while selecting a previewnet wallet, Reason: %v\n", selectWalletErr)
		return "", selectWalletErr
	}

	result := createPreviewAccount(wallet, c)
	return result, nil
}

// Create flow account
func createPreviewAccount(wallet *models.WalletPreview, c *gin.Context) string {

	tx := config.CreateFlowKey(wallet.HashAlgoString, wallet.SignAlgoString, wallet.PublicKey, wallet.Weight, "previewnet", c)
	txid := tx.ID()

	go generateWalletPreview(txid, wallet)

	return txid.String()
}

// Receive and save flow account
func generateWalletPreview(id flow.Identifier, wallet *models.WalletPreview) error {
	result := config.WaitForSeal(id, "previewnet")

	err := saveWalletPreview(wallet, result)
	if err != nil {
		log.Printf("Error while creating a wallet, Reason: %v\n", err)
		return err
	}

	return nil
}

// Save the flow account to database
func saveWalletPreview(wallet *models.WalletPreview, result string) error {

	savedAddress := "0x" + result
	wallet.Address = savedAddress
	_, walletErr := walletModelPreview.Update(wallet)

	if walletErr != nil {
		log.Printf("Error while saving, Reason: %v\n", walletErr)
		return walletErr
	}

	return nil

}
