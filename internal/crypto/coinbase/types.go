package coinbase

import (
	"context"
	"encoding/json"
	"time"
)

// Config are the per-account credentials (stored in core.Account.Config).
type Config struct {
	KeyName    string `json:"key_name"`    // CDP key name
	PrivateKey string `json:"private_key"` // EC private key (PEM)
}

// coinbase.Provider only knows this:
type API interface {
	ListPaymentMethods(ctx context.Context, cfg Config, pageToken string) (CnbV3PaymentData, error)
	ListFills(ctx context.Context, cfg Config, pageToken string) error
	ListOrders(ctx context.Context, cfg Config, pageToken string) error
	ListAccounts(ctx context.Context, cfg Config, pageToken string, limit int) (CnbPaginatedData, error)
	ListTransactions(ctx context.Context, cfg Config, accountId string, endpoint string, pageToken string) (CnbPaginatedData, error)
}

type CnbPaginatedData struct {
	Pagination CnbPagination     `json:"pagination"`
	Data       []json.RawMessage `json:"data"`
}

type CnbPeek struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CnbV3PaymentData struct {
	Payment_methods []CnbPaymentMethod
}

type CnbPaymentMethod struct {
	UID        string
	Acct_Id    string
	Id         string
	Name       string
	Type       string
	Created_At string
	Updated_At string
}

type CnbPagination struct {
	Ending_before  string
	Starting_after string
	Limit          int
	Order          string
	Previous_uri   string
	Next_uri       string
}

type AccountsCursor struct {
	Txns_next_url map[string]string
}

type CnbAccount struct {
	UID     string
	Acct_Id string
	Id      string
	Name    string
	Primary bool
	Type    string
	Balance struct {
		Currency string
	}
	Created_At   string
	Updated_At   string
	Resource     string
	ResourcePath string
}

type CnbTransaction struct {
	UID        string
	Acct_Id    string
	Id         string
	Type       string
	Status     string
	Created_At string
	// Updated_At          string
	Resource       string
	Resource_Path  string
	Amount         CnbAmount
	Native_Amount  CnbAmount
	Buy            CnbTransactionType
	Sell           CnbTransactionType
	Payment_Method struct {
		Id            string
		Resource      string
		Resource_Path string
	}
	Network struct {
		Network_Name    string
		Status          string
		Transaction_Url string
		Transaction_Fee CnbAmount
	}
	From struct {
		Id       string
		Name     string
		Resource string
		// Resource_Path string
	}
	To struct {
		Resource string
		Address  string
	}
	Advanced_trade_fill struct {
		Fill_Price       string
		Product_Id       string
		Commission       string
		Order_Side       string
		Commission_Total struct {
			Total_Commission  string
			Client_Commission string
		}
	}
}

type CnbTransactionType struct {
	Id                  string
	Payment_Method_Name string
	Fee                 CnbAmount
	// Amount              CnbAmount
	Total    CnbAmount
	Subtotal CnbAmount
}

type CnbAmount struct {
	Amount   string
	Currency string
}
