package solana

import (
	"time"
)

// coinbase.Provider only knows this:
type API interface {
	GetSolanaTokenAccounts(address string) (SolanaTokenAccountResult, error)
	GetSolanaSignaturesForAddress(addr string, untilSig string) (SolanaSignatureResult, error)
	GetSolanaTransaction(sig string) (SolanaParsedTransactionResult, error)
}

type SolanaTransactionCursor struct {
	UntilSig string `json:"until_sign,omitempty"`
}

// Input
type solanaInput struct {
	Id      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type ProgramId struct {
	ProgramId string `json:"programId"`
}
type MyEncoding struct {
	Encoding                       string `json:"encoding"`
	MaxSupportedTransactionVersion int    `json:"maxSupportedTransactionVersion"`
}

type solanaInputConfig struct {
	Limit int         `json:"limit"`
	Until interface{} `json:"until"`
}

// Outputs
type SolanaTokenAccountResult struct {
	Result struct {
		Value []SolanaTokenAccount
	}
}

type SolanaTokenAccount struct {
	Account struct {
		Data struct {
			Parsed struct {
				Info struct {
					Mint  string
					Owner string
				}
			}
		}
	}
	Pubkey string
}

type SolanaSignatureResult struct {
	Result []*SolanaSignature
}

type SolanaSignature struct {
	Address   string
	Stake     bool
	BlockTime uint64
	Date      *time.Time
	Signature string
	Slot      uint64
}

type SolanaParsedTransactionResult struct {
	Result SolanaParsedTransaction
}

type SolanaParsedTransaction struct {
	UID       string
	Acct_Id   string
	Signature string
	Address   string
	Stake     bool
	BlockTime uint64
	Date      *time.Time
	Meta      struct {
		Err struct {
			InstructionError []SolanaInstructionError
		}
		Fee               int
		InnerInstructions []SolanaParsedInnerInstruction
		PostBalances      []int64
		PreBalances       []int64
		PostTokenBalances []SolanaTokenBalance
		PreTokenBalances  []SolanaTokenBalance
	}
	Transaction struct {
		Message struct {
			AccountKeys  []SolanaParsedAccountKey
			Instructions []SolanaParsedInstruction
		}
	}
}

type SolanaInstructionError struct {
	Custom int
}

type SolanaParsedInnerInstruction struct {
	Index        int
	Instructions []SolanaParsedInstruction
}

type SolanaParsedInstruction struct {
	Parsed struct {
		Info SolanaParsedInstructionInfo
		Type string
	}
	Program   string
	ProgramId string
}

type SolanaParsedInstructionInfo struct {
	Account      string
	Amount       string
	Authority    string
	Destination  string
	Lamports     int64
	Mint         string
	NewAccount   string
	StakeAccount string
	Owner        string
	Source       string
	Wallet       string
	TokenAmount  struct {
		Amount         string
		Decimals       int
		UiAmount       float64
		UiAmountString string
	}
}

type SolanaTokenBalance struct {
	AccountIndex  int
	Mint          string
	UiTokenAmount struct {
		Amount         string
		Decimals       int
		UiAmount       float64
		UiAmountString string
	}
}

type SolanaParsedAccountKey struct {
	Pubkey   string
	Signer   bool
	Source   string
	Writable bool
}
