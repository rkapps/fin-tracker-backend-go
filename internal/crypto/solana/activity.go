package solana

import (
	"fmt"
	"strings"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

var (
	BASE_CURRENCY  = "SOL"
	TXN_DECIMALS   = decimal.NewFromInt(9)
	ATLAS_DECIMALS = decimal.NewFromInt(8)
	TOKEN_DECIMALS = decimal.NewFromInt(6)
)

var (
	TOKENADDRESS_SYMBOL = map[string]string{
		"So11111111111111111111111111111111111111112":  "SOL",
		"4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R": "RAY",
		"F5PPQHGcznZ2FxD9JaxJMXaf7XkaFFJ6zzTBcW8osQjw": "RAY-SOL",
		"7ngWUAjQBpUp5mgsLvGEReM5ec9SakysinFgsxCMZiDt": "RAY-SOL",
		"89ZKE4aoyfLBe2RuV6jM3JGNhaV18Nxh8eNtjRcndBip": "RAY-SOL",
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v": "USDC",
		"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB": "USDT",
		"ATLASXmbPQxBUYbxPsV97usA3fPQYEqzQBUHgiFCUsXx": "ATLAS",
		"Saber2gLauYim4Mvftnrasomsv6NvAuncvMEZwcLpD1":  "SBR",
		"SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt":  "SRM",
		"mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So":  "mSOL",
		"8Lg7TowFuMQoGiTsLE6qV9x3czRgDmVy8f8Vv8KS4uW":  "tuRAY",

		"9FC8xTFRbgTpuZZYAYnZLxgnQ8r7FwfSBM1SWvGwgF7s": "SBR-USDC",
		"418MFhkaYQtbn529wmjLLqL6uKxDz7j4eZBaV1cobkyd": "ATLAS-RAY",
	}
)

const (
	SOL_TOKEN         = "So11111111111111111111111111111111111111112"
	PGM_STAKE_ACCOUNT = "Stake11111111111111111111111111111111111111"
	PGM_STAKE_HISTORY = "SysvarStakeHistory1111111111111111111111111"
	PGM_DEF_ACCOUNT   = "11111111111111111111111111111111"

	PGM_SERUM_DEX_V3     = "9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"
	PGM_SERUM_SWAP       = "22Y43yTVxuUkoRKdm9thyRhQ3SdgQS7c7kB6UNCiaczD"
	PGM_RAYDIUM_POOL_V3  = "27haf8L6oxUeXrHrgEgsexjSY5hbVUWEmvv9Nyxg8vQv"
	PGM_RAYDIUM_POOL_V4  = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
	PGM_RAYDIUM_STAKE    = "EhhTKczWMGQt46ynNeRX1WfeagwwJd7ufHvCDjRxjo5Q"
	PGM_RAYDIUM_STAKE_V5 = "9KEPoZmtHUrBbhWN1v1KWLMkkvwY6WLtAVUCPRtRjP4z"
	PGM_RAYDIUM_ROUTING  = "routeUGWgWzqBWFcrCfv8tritsqukccJPu3q5GPP3xS"

	PGM_TOKEN_TRANSFER   = "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL"
	PGM_TOKEN_PROGRAM    = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	PGM_SOLFARM_VAULT    = "7vxeyaXGLqcp66fFShqUdHxdacp4k4kwUpRSSeoZLCZ4"
	PGM_BFF_LOADER       = "4bcFeLv4nydFrsZqV5CgwCVrPhkQKsXtzfy2KyMz7ozM"
	PGM_BFF_LOADER_1     = "noopb9bkMVfRPU8AsbpTUg8AQkHtKwMYZiFUjNRtMmV"
	PGM_MARINADE_FINANCE = "MarBmsSgKXdrN1egZf5sqe1TMai9K1rChYNDJgjq7aD"

	PGM_TULIP_VAULT        = "TLPv2tuSVvn3fSk8RgW3yPddkp5oFivzZV3rA9hQxtX"
	PGM_JUPITER_AGGREGATOR = "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4"
)

type SolanaActivity struct {
	saccts         []domain.Account
	ps             core.PriceService
	spamService    core.CryptoSpamService
	tokenAccountsm map[string]SolanaTokenAccount
	stakeAmountm   map[string]decimal.Decimal
	txn            SolanaParsedTransaction
	logger         *logger.Logger
	debug          bool
}

func NewSolanaActivity(saacts []domain.Account, ps core.PriceService,
	spamService core.CryptoSpamService,
	tokenAccountsm map[string]SolanaTokenAccount,
	stakeAmountm map[string]decimal.Decimal,
	txn SolanaParsedTransaction, logger *logger.Logger, debug bool) SolanaActivity {

	return SolanaActivity{saccts: saacts, ps: ps, spamService: spamService,
		tokenAccountsm: tokenAccountsm,
		stakeAmountm:   stakeAmountm, txn: txn, logger: logger, debug: debug,
	}
}

func (s SolanaActivity) ProcessTransaction() []*domain.Activity {

	var actvs []*domain.Activity
	innerInstructions := s.txn.Meta.InnerInstructions
	outerInstructions := s.txn.Transaction.Message.Instructions

	if s.debug {
		s.logger.Info("ProcessTransaction", "", fmt.Sprintf("Outer: %d Inner: %dd", len(outerInstructions), len(innerInstructions)))
	}

	instructionsm := make(map[int]SolanaParsedInnerInstruction)
	for _, inner := range innerInstructions {
		instructionsm[inner.Index] = inner
	}

	acctsm := make(map[string]SolanaParsedInstructionInfo)
	sacctsm := make(map[string]SolanaParsedInstructionInfo)
	actvCount := 1

	for z, instruction := range outerInstructions {

		var actv *domain.Activity
		parsed := instruction.Parsed
		if s.debug {
			s.logger.Info("ProcessTransaction", fmt.Sprintf("Outer #%d", z), fmt.Sprintf("Program:      %s", instruction.Program))
			s.logger.Info("ProcessTransaction", "", fmt.Sprintf("Type:         %s", parsed.Type))
			s.logger.Debug("ProcessTransaction", "", fmt.Sprintf("Owner         %s", parsed.Info.Owner))
			s.logger.Debug("ProcessTransaction", "", fmt.Sprintf("Account       %s", parsed.Info.Account))
			s.logger.Info("ProcessTransaction", "", fmt.Sprintf("Source        %s", parsed.Info.Source))
			s.logger.Debug("ProcessTransaction", "", fmt.Sprintf("Destination   %s", parsed.Info.Destination))
			s.logger.Debug("ProcessTransaction", "", fmt.Sprintf("Mint          %s", parsed.Info.Mint))
			s.logger.Debug("ProcessTransaction", "", fmt.Sprintf("Wallet        %s", parsed.Info.Wallet))
		}

		// //Check for spam sources
		if s.spamService.IsSpamSolanaSignature(parsed.Info.Source, crypto.BLOCKCHAIN_SOLANA) {
			if s.debug {
				s.logger.Info("ProcessTransaction", "Spam", fmt.Sprintf("Source: %s", parsed.Info.Source))
			}
			return nil
		}

		if strings.Compare(parsed.Type, "initializeAccount") == 0 {
			acctsm[parsed.Info.Account] = parsed.Info
			continue
		} else if strings.Compare(parsed.Type, "create") == 0 {
			sacctsm[parsed.Info.Account] = parsed.Info
			continue
		}

		actv = s.processOuterInstruction(instruction, sacctsm)
		if actv != nil {
			actvs = append(actvs, actv)
			if s.debug {
				s.logger.Info("ProcessTransaction", "", fmt.Sprintf("---%s---", actv.TxnType))
			}
			if actv.TxnType == domain.ActivityTypeUnStake {
				stakeAccount := strings.ReplaceAll(actv.Notes, "StakeAccount: ", "")
				stakeAmount := s.stakeAmountm[stakeAccount]
				// log.Println(stakeAmountm)
				if s.debug {
					s.logger.Info("ProcessTransaction", "", fmt.Sprintf("StakeAccount: %s-%v", stakeAccount, stakeAmount))
				}

				if stakeAmount.IsPositive() {
					rewardAmount := actv.RcvAmount.Sub(stakeAmount)
					if s.debug {
						s.logger.Info("ProcessTransaction", "", fmt.Sprintf("Rewards: %v", rewardAmount))
					}
					if rewardAmount.IsPositive() {
						ractv := &domain.Activity{}
						ractv.UID = actv.UID
						ractv.AccountID = actv.AccountID
						ractv.ID = fmt.Sprintf("%s-%d", actv.ID, 1)
						ractv.Date = actv.Date
						ractv.TxnType = domain.ActivityTypeReward
						ractv.RcvAccountID = actv.AccountID
						ractv.RcvSymbol = actv.RcvSymbol
						ractv.RcvAmount = rewardAmount
						price, _ := s.ps.GetCryptoPrice(ractv.RcvSymbol, ractv.Date)
						ractv.SentAmount = rewardAmount.Mul(price)
						actvs = append(actvs, ractv)
						if s.debug {
							s.logger.Info("ProcessTransaction", "", fmt.Sprintf("---%s---", ractv.TxnType))
						}
					}
				}
			}
			// core.PrintActivity(debug, actv)
		}

		iactvs := []*domain.Activity{}
		innerInstruction := instructionsm[z]
		stake := false
		for i, inner := range innerInstruction.Instructions {

			if s.debug {
				s.logger.Info("ProcessTransaction", "", fmt.Sprintf("Inner #%d program: %s", i, inner.ProgramId))
				// log.Println(inner)
				// s.logger.Debug("ProcessTransaction", "Account", parsed.Info.Account)
				// s.logger.Debug("ProcessTransaction", "Mint", parsed.Info.Mint)
			}
			if strings.Compare(inner.ProgramId, PGM_STAKE_ACCOUNT) == 0 {
				stake = true
			}
			iactv := s.processSingleInnerInstruction(inner, acctsm)
			if iactv != nil {

				// update txntype and notes
				s.updateActivityDetails(instruction, iactv, stake)
				if len(iactvs) > 0 {
					iactv.ID = fmt.Sprintf("%s-%d", iactv.ID, actvCount)
					// this fix is only so tradein go in first
					if iactv.TxnType == domain.ActivityTypeTradeIn {
						iactv.Date = iactv.Date.Add(time.Second * 1)
					}

					actvCount++
				}
				iactvs = append(iactvs, iactv)
				if s.debug {
					s.logger.Info("ProcessTransaction", "", iactv.ID)
					s.logger.Info("ProcessTransaction", "", fmt.Sprintf("---%s---", iactv.TxnType))
				}
				// s.logger.Info("ProcessTransaction", "", iactv.GetSendReceiveInfo())
			}
		}

		actvs = append(actvs, iactvs...)
	}

	if s.debug && len(actvs) > 0 {
		s.logger.Info("ProcessTransaction")
	}

	return actvs
}

func (s SolanaActivity) processOuterInstruction(
	instruction SolanaParsedInstruction,
	acctsTokenm map[string]SolanaParsedInstructionInfo,
) *domain.Activity {

	parsed := instruction.Parsed
	info := parsed.Info
	amount, _ := crypto.ConvertInt64ToBaseDecimal(info.Lamports, TXN_DECIMALS)

	if s.debug {
		s.logger.Debug("OuterInstruction", "Lamport", parsed.Info.Lamports, "Amount", amount)
		s.logger.Debug("OuterInstruction", "Source", info.Source)
		s.logger.Debug("OuterInstruction", "Destination", info.Destination)
		s.logger.Debug("OuterInstruction", "Mint", info.Mint)
		s.logger.Debug("OuterInstruction", "Authority", info.Authority)
		s.logger.Debug("OuterInstruction", "NewAccount", info.NewAccount)
	}

	// transfer
	// createAccountWithSeed
	// initialize
	// delegate
	// deactivate
	// withdraw
	// blank

	if strings.Compare(parsed.Type, "initialize") == 0 ||
		strings.Compare(parsed.Type, "initializeAccount") == 0 ||
		strings.Compare(parsed.Type, "closeAccount") == 0 ||
		strings.Compare(parsed.Type, "deactivate") == 0 ||
		strings.Compare(parsed.Type, "delegate") == 0 {
		return nil
	}

	if strings.Compare(parsed.Type, "transfer") == 0 {

		if amount.IsPositive() {

			fAddrAcct := crypto.GetAccountFromAddress(s.saccts, info.Source)
			tAddrAcct := crypto.GetAccountFromAddress(s.saccts, info.Destination)
			if s.debug {
				if fAddrAcct != nil {
					s.logger.Info("OuterInstruction", "", fmt.Sprintf("From: %s", fAddrAcct.ID))
				}
				if tAddrAcct != nil {
					s.logger.Info("OuterInstruction", "", fmt.Sprintf("To  : %s", tAddrAcct.ID))
				}
			}
			return s.createInstructionActivity(fAddrAcct, tAddrAcct, BASE_CURRENCY, amount)
		}

	} else if strings.Compare(parsed.Type, "createAccountWithSeed") == 0 ||
		strings.Compare(parsed.Type, "createAccount") == 0 {

		fAddrAcct := crypto.GetAccountFromAddress(s.saccts, info.Source)
		actv := s.createInstructionActivity(fAddrAcct, nil, BASE_CURRENCY, amount)
		actv.TxnType = domain.ActivityTypeStake
		actv.Notes = fmt.Sprintf("StakeAccount: %s", info.NewAccount)
		// add to stake account map
		s.stakeAmountm[info.NewAccount] = amount

		return actv

	} else if strings.Compare(parsed.Type, "withdraw") == 0 {

		if amount.IsPositive() {
			tAddrAcct := crypto.GetAccountFromAddress(s.saccts, info.Destination)
			actv := s.createInstructionActivity(nil, tAddrAcct, BASE_CURRENCY, amount)
			actv.TxnType = domain.ActivityTypeUnStake
			actv.Notes = fmt.Sprintf("StakeAccount: %s", info.StakeAccount)
			return actv
		}

	} else if strings.Compare(parsed.Type, "transferChecked") == 0 {

		// acct := s.AccountsService.GetAccount(ctx, result.Acct_Id)

		pinfo := acctsTokenm[info.Destination]
		// if debug {
		// 	log.Printf("            Source: %v", pinfo.Source)
		// 	log.Printf("            Source: %v", pinfo.Wallet)
		// }
		fAddrAcct := crypto.GetAccountFromAddress(s.saccts, pinfo.Source)
		tAddrAcct := crypto.GetAccountFromAddress(s.saccts, pinfo.Wallet)

		symbol := TOKENADDRESS_SYMBOL[info.Mint]
		amount = decimal.NewFromFloat(info.TokenAmount.UiAmount)
		if s.debug {
			s.logger.Info("OuterInstruction", "Mint-symbol", fmt.Sprintf("%s-%s", info.Mint, symbol), "Amount", amount)
		}
		if amount.IsPositive() {
			return s.createInstructionActivity(fAddrAcct, tAddrAcct, symbol, amount)
		}

	}
	return nil
}

func (s SolanaActivity) processSingleInnerInstruction(
	instruction SolanaParsedInstruction,
	acctsTokenm map[string]SolanaParsedInstructionInfo) *domain.Activity {

	parsed := instruction.Parsed
	info := instruction.Parsed.Info

	if s.debug {
		s.logger.Info("InnerInstruction", "", fmt.Sprintf("Type          %s", parsed.Type))
		s.logger.Info("InnerInstruction", "", fmt.Sprintf("Mint          %s", info.Mint))
		s.logger.Debug("InnerInstruction", "", fmt.Sprintf("Lamports:     %v", info.Lamports))
		s.logger.Debug("InnerInstruction", "", fmt.Sprintf("Amount:       %v", info.Amount))
		s.logger.Debug("InnerInstruction", "", fmt.Sprintf("NewAccount:   %s", info.NewAccount))
	}

	if strings.Compare(parsed.Type, "delegate") == 0 ||
		strings.Compare(parsed.Type, "allocate") == 0 ||
		strings.Compare(parsed.Type, "assign") == 0 ||
		strings.Compare(parsed.Type, "getAccountDataSize") == 0 ||
		strings.Compare(parsed.Type, "initializeAccount") == 0 ||
		strings.Compare(parsed.Type, "initializeAccount3") == 0 ||
		strings.Compare(parsed.Type, "initializeImmutableOwner") == 0 ||
		strings.Compare(parsed.Type, "createAccount") == 0 ||
		strings.Compare(parsed.Type, "closeAccount") == 0 ||
		strings.Compare(parsed.Type, "") == 0 ||
		strings.Compare(parsed.Type, "deactivate") == 0 {
		return nil
	}

	symbol := BASE_CURRENCY
	// acct := s.AccountsService.GetAccount(ctx, result.Acct_Id)
	var mergeBal int64

	fAddrAcct := crypto.GetAccountFromAddress(s.saccts, info.Source)
	tAddrAcct := crypto.GetAccountFromAddress(s.saccts, info.Destination)
	if s.debug {
		fAccount := ""
		tAccount := ""
		if fAddrAcct != nil {
			fAccount = fAddrAcct.ID
		}
		if tAddrAcct != nil {
			tAccount = tAddrAcct.ID
		}
		s.logger.Info("InnerInstruction", "", fmt.Sprintf("Source        %s", info.Source), "Account", fAccount)
		s.logger.Info("InnerInstruction", "", fmt.Sprintf("Destination   %s", info.Destination), "Account", tAccount)
	}

	if fAddrAcct == nil && tAddrAcct == nil {
		//spl token
		token := ""
		if strings.Compare(parsed.Type, "transfer") == 0 || strings.Compare(parsed.Type, "burn") == 0 {

			tokenAccount := s.tokenAccountsm[info.Source]
			if len(tokenAccount.Pubkey) > 0 {
				token = tokenAccount.Account.Data.Parsed.Info.Mint
			}
			if len(token) == 0 {
				pinfo := acctsTokenm[info.Source]
				token = pinfo.Mint
				// find the burn token
				if len(token) == 0 {
					tokenAccount := s.tokenAccountsm[info.Mint]
					if len(tokenAccount.Pubkey) > 0 {
						token = tokenAccount.Account.Data.Parsed.Info.Mint
					}
				}
			}

			symbol = TOKENADDRESS_SYMBOL[token]
			if s.debug {
				s.logger.Info("InnerInstruction", "", fmt.Sprintf("Source - Token: %s", token), "Symbol", symbol)
			}

			if len(symbol) > 0 {
				fAddrAcct = crypto.GetAccountFromAddress(s.saccts, s.txn.Address)
			} else {

				tokenAccount := s.tokenAccountsm[info.Destination]
				if len(tokenAccount.Pubkey) > 0 {
					token = tokenAccount.Account.Data.Parsed.Info.Mint
				}
				if len(token) == 0 {
					pinfo := acctsTokenm[info.Destination]
					token = pinfo.Mint
				}
				if len(token) > 0 {
					symbol = TOKENADDRESS_SYMBOL[token]
					tAddrAcct = crypto.GetAccountFromAddress(s.saccts, s.txn.Address)
				}
				if s.debug {
					s.logger.Info("InnerInstruction", "", fmt.Sprintf("Destination - Token: %s", token), "Symbol", symbol)
				}
			}
		} else if strings.Compare(parsed.Type, "mintTo") == 0 {

			if len(info.Mint) > 0 {
				token = info.Mint
				if len(token) > 0 {
					symbol = TOKENADDRESS_SYMBOL[token]
					if len(symbol) > 0 {
						tAddrAcct = crypto.GetAccountFromAddress(s.saccts, s.txn.Address)
					}
				}
			}
		} else if strings.Compare(parsed.Type, "transferChecked") == 0 {

			token = info.Mint
			symbol = TOKENADDRESS_SYMBOL[token]
			if len(symbol) > 0 {
				if strings.Compare(s.txn.Address, info.Authority) == 0 {
					fAddrAcct = crypto.GetAccountFromAddress(s.saccts, s.txn.Address)
				} else {
					tAddrAcct = crypto.GetAccountFromAddress(s.saccts, s.txn.Address)
				}
			}

		} else {
			// log.Println(parsed.Type)
		}
	}
	decimals := TXN_DECIMALS
	if strings.Compare(symbol, BASE_CURRENCY) == 0 || strings.Compare(symbol, "mSOL") == 0 {
	} else if strings.Compare(symbol, "ATLAS") == 0 {
		decimals = ATLAS_DECIMALS
	} else {
		decimals = TOKEN_DECIMALS
	}

	amount := decimal.Zero
	if mergeBal > 0 {
		// amount = core.GetDecimalValue(float64(mergeBal), decimals)
	} else if info.Lamports > 0 {
		amount, _ = crypto.ConvertInt64ToBaseDecimal(info.Lamports, decimals)
	} else {
		amount, _ = crypto.ConvertStringToBaseDecimal(info.Amount, decimals)
	}

	if strings.Compare(parsed.Type, "transferChecked") == 0 {
		amount = decimal.NewFromFloat(info.TokenAmount.UiAmount)
	}
	if s.debug {
		s.logger.Info("InnerInstruction", "", fmt.Sprintf("Symbol: %s", symbol), "Amount", amount)
	}

	if amount.IsZero() {
		return nil
	}

	return s.createInstructionActivity(fAddrAcct, tAddrAcct, symbol, amount)
	// return nil
}

func (s SolanaActivity) createInstructionActivity(
	fAddrAcct *domain.Account,
	tAddrAcct *domain.Account,
	symbol string,
	amount decimal.Decimal,
) *domain.Activity {

	actv := &domain.Activity{}
	actv.UID = s.txn.UID
	// actv.AccountID = s.txn.Acct_Id
	actv.ID = s.txn.Signature
	actv.Hash = s.txn.Signature
	actv.Date = *s.txn.Date

	if fAddrAcct != nil && tAddrAcct != nil {

		actv.AccountID = fAddrAcct.ID
		actv.TxnType = domain.ActivityTypeTransfer
		actv.SentAccountID = fAddrAcct.ID
		actv.SentSymbol = symbol
		actv.SentAmount = amount
		price, _ := s.ps.GetCryptoPrice(actv.SentSymbol, actv.Date)
		actv.SentPrice = price

		actv.RcvAccountID = tAddrAcct.ID
		actv.RcvSymbol = symbol
		actv.RcvAmount = amount

	} else if fAddrAcct != nil {

		actv.AccountID = fAddrAcct.ID
		actv.TxnType = domain.ActivityTypeSend
		actv.SentAccountID = fAddrAcct.ID
		actv.SentSymbol = symbol
		actv.SentAmount = amount

		price, _ := s.ps.GetCryptoPrice(actv.SentSymbol, actv.Date)
		actv.SentPrice = price
		actv.RcvSymbol = "USD"
		actv.RcvPrice = decimal.NewFromInt(1)
		actv.RcvAmount = actv.SentAmount.Mul(price)

	} else if tAddrAcct != nil {

		actv.AccountID = tAddrAcct.ID
		actv.TxnType = domain.ActivityTypeReceive
		actv.RcvAccountID = tAddrAcct.ID
		actv.RcvSymbol = symbol
		actv.RcvAmount = amount

		price, _ := s.ps.GetCryptoPrice(actv.RcvSymbol, actv.Date)
		if s.debug {
			s.logger.Info("createActivity", "", actv.Date, "Symbol", fmt.Sprintf("%s-%v", actv.RcvSymbol, price))
		}
		actv.RcvPrice = price
		actv.SentSymbol = "USD"
		actv.SentPrice = decimal.NewFromInt(1)
		actv.SentAmount = actv.RcvAmount.Mul(price)
	} else {
		return nil
	}

	return actv
}

func (s SolanaActivity) updateActivityDetails(instruction SolanaParsedInstruction, actv *domain.Activity, stake bool) {

	if stake {
		if s.debug {
			s.logger.Info("updateActivityDetails", "Actv", actv.ID, "TxnType", actv.TxnType)
		}
		switch actv.TxnType {
		case domain.ActivityTypeSend:
			actv.TxnType = domain.ActivityTypeStake
		case domain.ActivityTypeReceive:
			actv.TxnType = domain.ActivityTypeUnStake
		}
	}

	if strings.Compare(instruction.ProgramId, "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL") == 0 {
		// actv.TxnType = activities.TXN_TYPE_OTHER_FEE
		actv.Notes = "ACCOUNT_CREATE"

	} else if strings.Compare(instruction.ProgramId, PGM_SERUM_DEX_V3) == 0 ||
		strings.Compare(instruction.ProgramId, PGM_SERUM_SWAP) == 0 ||
		strings.Compare(instruction.ProgramId, PGM_BFF_LOADER) == 0 ||
		strings.Compare(instruction.ProgramId, PGM_RAYDIUM_ROUTING) == 0 {

		switch actv.TxnType {
		case domain.ActivityTypeSend:
			actv.TxnType = domain.ActivityTypeTradeIn
		case domain.ActivityTypeReceive:
			actv.TxnType = domain.ActivityTypeTradeOut
		}
		actv.Notes = "SERUM_DEX"
		if strings.Compare(instruction.ProgramId, PGM_RAYDIUM_ROUTING) == 0 {
			actv.Notes = "RADIUM_ROUTING"
		}

	} else if strings.Compare(instruction.ProgramId, PGM_RAYDIUM_POOL_V3) == 0 ||

		strings.Compare(instruction.ProgramId, PGM_RAYDIUM_POOL_V4) == 0 {
		switch actv.TxnType {
		case domain.ActivityTypeSend:
			actv.TxnType = domain.ActivityTypeAddLiquidity
		case domain.ActivityTypeReceive:
			actv.TxnType = domain.ActivityTypeExitLiquidity
		}
		actv.Notes = "RAYDIUM_POOL"

	} else if strings.Compare(instruction.ProgramId, PGM_RAYDIUM_STAKE) == 0 ||
		strings.Compare(instruction.ProgramId, PGM_RAYDIUM_STAKE_V5) == 0 ||
		strings.Compare(instruction.ProgramId, PGM_SOLFARM_VAULT) == 0 ||
		strings.Compare(instruction.ProgramId, PGM_TULIP_VAULT) == 0 ||
		strings.Compare(instruction.ProgramId, PGM_JUPITER_AGGREGATOR) == 0 ||
		strings.Compare(instruction.ProgramId, PGM_MARINADE_FINANCE) == 0 {

		switch actv.TxnType {
		case domain.ActivityTypeSend:
			actv.TxnType = domain.ActivityTypeAddLiquidity
		case domain.ActivityTypeReceive:
			actv.TxnType = domain.ActivityTypeExitLiquidity
		}

		actv.Notes = "RAYDIUM_STAKE"
		if strings.Compare(instruction.ProgramId, PGM_SOLFARM_VAULT) == 0 {
			actv.Notes = "SOLFARE_VAULT"
		} else if strings.Compare(instruction.ProgramId, PGM_MARINADE_FINANCE) == 0 {
			actv.Notes = "MARINADE_STAKE"
		} else if strings.Compare(instruction.ProgramId, PGM_TULIP_VAULT) == 0 {
			actv.Notes = "TULIP_VAULT"
		} else if strings.Compare(instruction.ProgramId, PGM_JUPITER_AGGREGATOR) == 0 {
			actv.Notes = "JUPITER_AGGREGATOR"
		}

	}
}
