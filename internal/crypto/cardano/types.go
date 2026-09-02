package cardano

import (
	"context"
	"time"
)

// coinbase.Provider only knows this:
type API interface {
	GetEpochInformation(ctx context.Context, epoch int64, page int64) ([]EpochInformation, error)
	GetAccountRewards(ctx context.Context, saddress string, page int, count int) ([]AccountReward, error)
	GetAccountAddresses(ctx context.Context, saddress string) ([]AccountAddress, error)
	GetAccountTransactions(ctx context.Context, address string, bheight int64, page int) ([]AddressTransaction, error)
	GetTransactionInfo(txHash string) (TransactionInfo, error)
	GetTransactionUTXOs(txHash string) (TransactionUTXO, error)
	GetTransactionMetadata(txHash string) ([]TransactionMetadata, error)
	GetTransactionStakeCerticates(txHash string) ([]TransasctionStakeCertificate, error)
	GetTransactionWithdrawals(stakeaddress string) ([]TransactionWithdrawal, error)
	GetTransactionDelegations(txHash string) ([]TransactionDelegation, error)
}

type cardanoEpochCursor struct {
	Epoch int64 `json:"epoch"` // starting epoch to walk forward from
	Page  int64 `json:"page"`  // current page within that "next epochs" listing
}

type cardanoRewardsCursor struct {
	Page int `json:"page"`
}

type cardanoTransactionCursor struct {
	BlockHeight int64 `json:"block_height"`
}

type EpochInformation struct {
	Epoch          int64   `json:"epoch"`
	StartTime      int64   `json:"start_time"`
	EndTime        int64   `json:"end_time"`
	FirstBlockTime int64   `json:"first_block_time"`
	LastBlockTime  int64   `json:"last_block_time"`
	BlockCount     int64   `json:"block_count"`
	TxCount        int64   `json:"tx_count"`
	Output         string  `json:"output"`
	Fees           string  `json:"fees"`
	ActiveStake    *string `json:"active_stake"` // nullable — pointer to handle null cleanly
}

type AccountReward struct {
	Epoch  int64  `json:"epoch"`
	Amount string `json:"amount"`
	PoolId string `json:"pool_id"`
}

type AccountAddress struct {
	Address string
}

type AddressTransaction struct {
	TxHash      string `json:"tx_hash"`
	BlockHeight int64  `json:"block_height"`
	BlockTime   int64  `json:"block_time"`
}

// Cardano transaction
type Transaction struct {
	UID               string
	AccountID         string
	TxHash            string
	BlockTime         *time.Time
	BlockHeight       int64
	Fees              string
	UTXO              TransactionUTXO
	Metadata          []TransactionMetadata
	StakeCertificates []TransasctionStakeCertificate
	Delegations       []TransactionDelegation
	WithdrawalAmount  string
}

type TransactionUTXO struct {
	Inputs  []TransactionEntry
	Outputs []TransactionEntry
}

type TransactionEntry struct {
	Address      string
	Amount       []TransactionAmount
	TxHash       string `json:"tx_hash"`
	OutputIndex  int    `json:"output_index"`
	Collateral   bool
	Reference    bool
	ConsumedByTx string `json:"consumed_by_tx"`
}

type TransactionAmount struct {
	Unit     string
	Quantity string
}

type TransactionMetadata struct {
	Label        string
	JsonMetadata map[string]any `json:"json_metadata"`
}

type TransasctionStakeCertificate struct {
	CertIndex    int `json:"cert_index"`
	Address      string
	Registration bool
}

type TransactionWithdrawal struct {
	TxHash string `json:"tx_hash"`
	Amount string
}

type TransactionInfo struct {
	TxHash string `json:"hash"`
	Fees   string `json:"fees"`
}

type TransactionDelegation struct {
	Index     int
	CertIndex int `json:"cert_index"`
	Addresss  string
	PoolId    string `json:"pool_id"`
}
