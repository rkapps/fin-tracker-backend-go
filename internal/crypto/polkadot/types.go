package polkadot

import (
	"time"
)

// coinbase.Provider only knows this:
type API interface {
	GetRewards(address string, page int, row int) (*PolkadotRewardData, error)
	GetTransfers(address string, page int, row int) (*PolkadotTransferData, error)
}

type polkadotRewardsCursor struct {
	Page int `json:"page"`
}
type polkadotTransfersCursor struct {
	Page int `json:"page"`
}

type PolkadotTransferInput struct {
	Address  string `json:"address"`
	After_Id []int  `json:"after_id"`
	Order    string `json:"order"`
	Page     int    `json:"page"`
	Row      int    `json:"row"`
}

type PolkadotTransferData struct {
	Message      string
	Generated_At int64
	Data         struct {
		Cound     int
		Transfers []PolkadotTransfer
	}
}

type PolkadotTransfer struct {
	UID                  string
	AccountId            string
	From                 string
	To                   string
	Success              bool
	Hash                 string
	Block_Num            int64
	Block_Timestamp      int64
	Date                 *time.Time
	Amount               string
	Asset_Symbol         string
	Asset_Unique_Id      string
	Asset_Type           string
	From_Account_Display struct {
		Address string
		Merkle  struct {
			Address_Type string
			Tag_Type     string
			Tag_Subtype  string
			Tag_Name     string
		}
	}
	To_Account_Display struct {
		Address string
	}
}

type PolkadotRewardInput struct {
	Address     string `json:"address"`
	Block_Range string `json:"block_range"`
	Category    string `json:"category"`
	Is_Stash    bool   `json:"is_stash"`
	Page        int    `json:"page"`
	Row         int    `json:"row"`
}

type PolkadotRewardData struct {
	Message      string
	Generated_At int64
	Data         struct {
		Cound int
		List  []PolkadotReward
	}
}

type PolkadotReward struct {
	UID             string
	AccountId       string
	Block_Num       int64
	Stash           string
	Account         string
	Module_Id       string
	Event_Idx       string
	Event_Method    string
	Event_Index     string
	Amount          string
	Block_Timestamp int64
	Date            *time.Time
}
