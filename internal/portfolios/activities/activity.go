package activities

import "time"

const (
	AVTIVITY_IMPORT_COL = "activity_import"
	AVTIVITY_COL        = "activity"
)

type Activity struct {
	ID      string     `json:"-"`
	Acct_ID string     `json:"acctId"`
	Hash    string     `json:"hash"`
	Date    *time.Time `json:"date"`
	TxnType string     `json:"txnType"`

	Rcv_Acct_ID string  `json:"rcvAcctId"`
	Rcv_Symbol  string  `json:"rcvSymbol"`
	Rcv_Amount  float64 `json:"rcvAmount"`
	Rcv_Price   float64 `json:"rcvPrice"`
	Rcv_Value   float64 `json:"rcvValue"`
	Rcv_Balance float64 `json:"rcvBalance"`
	Rcv_Address string  `json:"rcvAddress"`

	Sent_Acct_ID string  `json:"sentAcctId"`
	Sent_Symbol  string  `json:"sentSymbol"`
	Sent_Amount  float64 `json:"sentAmount"`
	Sent_Address string  `json:"sentAddress"`
	Sent_Price   float64 `json:"sentPrice"`
	Sent_Value   float64 `json:"sentValue"`
	Sent_Balance float64 `json:"sentBalance"`

	Value      float64 `json:"value"`
	Gl_Amount  float64 `json:"glAmount"`
	Error_Mesg string  `json:"errorMessage"`

	Transaction_Url string  `json:"transactionUrl"`
	Fee_Amount      float64 `json:"feeAmount"`
	Fee_Symbol      string  `json:"feeSymbol"`
	Notes           string  `json:"notes"`
	Tag             string  `json:"tag"`
}

type Activities []*Activity

type ActivityImport struct {
	ID            string     `json:"-"`
	UID           string     `json:"-"`
	Acct_ID       string     `json:"acctId"`
	Hash          string     `json:"hash"`
	TxnWallet     string     `json:"txnWallet"`
	TxnType       string     `json:"txnType"`
	Date          *time.Time `json:"date"`
	Rcv_Account   string     `json:"rcvAccount"`
	Rcv_Address   string     `json:"rcvAddress"`
	Rcv_Currency  string     `json:"rcvCurrency"`
	Rcv_Amount    float64    `json:"rcvAmount"`
	Sent_Account  string     `json:"sentAccount"`
	Sent_Address  string     `json:"sentAddress"`
	Sent_Currency string     `json:"sentCurrency"`
	Sent_Amount   float64    `json:"sentAmount"`
	Sent_Price    float64    `json:"sentPrice"`
	Sent_Balance  float64    `json:"sentBalance"`
	Gl_Amount     float64    `json:"glAmount"`
	Fee           float64    `json:"fee"`
	Fee_Currency  string     `json:"feeCurrency"`
	Notes         string     `json:"notes"`
}

func (a ActivityImport) Id() string {
	return a.ID
}

func (a ActivityImport) SetId() {

}

func (a ActivityImport) CollectionName() string {
	return AVTIVITY_IMPORT_COL
}

type ActivityImports []*ActivityImport
