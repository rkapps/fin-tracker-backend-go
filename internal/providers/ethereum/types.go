package ethereum

import "github.com/nanmu42/etherscan-api"

type API interface {
	GetNormalTransactions(addr string, blockNumber *int, page int, offset int) ([]etherscan.NormalTx, error)
	GetInternalTransactions(addr string, blockNumber *int, page int, offset int) ([]etherscan.InternalTx, error)
	GetERC20Transfers(addr string, blockNumber *int, page int, offset int) ([]etherscan.ERC20Transfer, error)
	GetERC721Transfers(addr string, blockNumber *int, page int, offset int) ([]etherscan.ERC721Transfer, error)
}

type ethereumCursor struct {
	BlockNumber int `json:"block_number"`
}
