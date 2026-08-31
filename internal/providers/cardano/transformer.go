package cardano

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/providers"
)

type CardanoAccountTransformer struct {
	logger *logger.Logger
}

func NewCardanoAccountTransformer(logConfig *logger.Config) CardanoAccountTransformer {
	plog := logConfig.For("refresher.cardano")
	return CardanoAccountTransformer{plog}
}

func (c CardanoAccountTransformer) Name() string {
	return "cardano"
}

func (c CardanoAccountTransformer) Transform(ctx context.Context, ps core.PriceService,
	gaccts []*domain.Account,
	acreds []domain.AccountWithCredential,
	globalRaws []domain.RawItem,
	rawsm map[string][]domain.RawItem,
) ([]*domain.Activity, error) {

	actvs := []*domain.Activity{}
	// c.logger.Info("Transform", "Provider", c.Name(), "Items", fmt.Sprintf("%d-%d", len(globalRaws), len(raws)))
	addrAcctm, addrsm, rewardsm, txnsm := c.marshalData(rawsm)

	c.logger.Info("Transform", "Provider", c.Name(), "Rewards", len(rewardsm))
	c.logger.Info("Transform", "Provider", c.Name(), "Txns", len(txnsm))

	// add rewards activities
	ractvs := c.handleRewards(acreds, ps, globalRaws, rewardsm)
	actvs = append(actvs, ractvs...)

	withdrawalsm := make(map[string]string)
	txns := []Transaction{}

	for _, txn := range txnsm {
		txns = append(txns, txn...)
		for _, entry := range txns {
			if len(entry.WithdrawalAmount) > 0 {
				withdrawalsm[entry.TxHash] = entry.WithdrawalAmount
			}
		}
	}

	sort.Slice(txns, func(i, j int) bool {
		date1 := txns[i].BlockTime
		date2 := txns[j].BlockTime
		return date1.Before(*date2)
	})
	c.logger.Info("Transform", "Provider", c.Name(), "Txns", len(txns))

	// map tx hash
	atxnsm := make(map[string]string)

	for i, txn := range txns {

		if i > 500 {
			// break
		}
		// withdrawal := withdrawalsm[txn.TxHash]
		if _, ok := atxnsm[txn.TxHash]; ok {
			continue
		}
		debug := false
		atxnsm[txn.TxHash] = txn.TxHash
		addresses := addrsm[txn.AccountID]
		if strings.Compare(txn.TxHash, "eec7a867a3a1bee054aebf93a5fcc83da5c05a7a1e5ff6b6354037805135a9d1--1") == 0 ||
			strings.Compare(txn.TxHash, "3b6ad048929d4163036a1a124b81e0736c735f719f33396eab4386d05df0fa4b") == 0 {
			// debug = true
		}

		if debug {
			c.logger.Info("Transform", "AccountId", txn.AccountID, "Addresses", len(addresses))
			c.logger.Info("Transform", "TxHash", txn.TxHash, "Date", txn.BlockTime)
		}

		// log.Println(addresses)
		aActivity := NewAccountActivity(addrAcctm, withdrawalsm, c.logger, debug)
		aActivity.add_transaction(txn)
		uactvs, err := aActivity.getActivity()

		if err != nil {
			actv := &domain.Activity{}
			actv.UID = txn.UID
			actv.AccountID = txn.AccountID
			actv.ID = txn.TxHash
			actv.Hash = txn.TxHash
			actv.Date = *txn.BlockTime
			actv.Notes = err.Error()
			actvs = append(actvs, actv)
			continue
		}

		if debug {
			c.logger.Debug("Transform", "TxHash", txn.TxHash, "UTXO", len(uactvs))
		}
		for x, uactv := range uactvs {
			id := txn.TxHash
			if x > 0 {
				id = fmt.Sprintf("%s-%d", txn.TxHash, x)
			}
			// uactv.UID = txn.UID
			// uactv.AccountID = txn.AccountID
			uactv.ID = id
			uactv.Hash = txn.TxHash
			uactv.Date = *txn.BlockTime
			actvs = append(actvs, uactv)
		}
	}

	c.logger.Info("Transform", "Provider", c.Name(), "Actvs", len(actvs))
	return actvs, nil
}

func (c CardanoAccountTransformer) handleRewards(
	acred []domain.AccountWithCredential,
	ps core.PriceService,
	globalRaws []domain.RawItem,
	rewardsm map[string][]AccountReward,
) []*domain.Activity {

	actvs := []*domain.Activity{}
	epochsm := c.marshalEpochData(globalRaws)

	// add rewards activiites
	for _, acred := range acred {
		rewards := rewardsm[acred.Account.ID]
		for _, reward := range rewards {
			epoch := epochsm[reward.Epoch]
			if epoch.Epoch == 0 {
				continue
			}
			actv := &domain.Activity{}
			actv.UID = acred.Account.UID
			actv.AccountID = acred.Account.ID
			actv.ID = fmt.Sprintf("%s-%v", acred.Account.ID, epoch.Epoch)
			tm := time.Unix(epoch.StartTime, 10)
			actv.Date = tm
			actv.TxnType = domain.ActivityTypeReward
			actv.RcvAccountID = acred.Account.ID
			actv.RcvSymbol = "ADA"
			amount, _ := providers.ConvertStringToBaseDecimal(reward.Amount, TXN_DECIMALS)
			actv.RcvAmount = amount
			price, err := ps.GetCryptoPrice(actv.RcvSymbol, actv.Date)
			if err != nil {
				// k.logger.Error("Transform", "Ledger", ledger.Debug(), "Error", err)
				// continue
			}
			actv.SentAmount = amount.Mul(price)
			actv.SentSymbol = "USD"

			// if epoch.Epoch >= 215 && epoch.Epoch <= 220 {
			// 	c.logger.Info("Transform", "Date", actv.Date, "Reward", fmt.Sprintf("%s-%v-%v-%v", actv.RcvSymbol, actv.RcvAmount, price, actv.SentAmount))
			// }
			actv.Notes = fmt.Sprintf("Epoch: %v", epoch.Epoch)
			actvs = append(actvs, actv)
		}

	}
	return actvs

}

func (c CardanoAccountTransformer) marshalData(rawsm map[string][]domain.RawItem,
) (
	map[string]string,
	map[string][]string,
	map[string][]AccountReward,
	map[string][]Transaction,
) {

	addrAcctm := make(map[string]string)
	addrsm := make(map[string][]string)
	rewardsm := make(map[string][]AccountReward)
	txnsm := make(map[string][]Transaction)
	txhasm := make(map[string]string)

	for accountId, raws := range rawsm {

		var addresses []string
		var rewards []AccountReward
		var txns []Transaction

		for _, raw := range raws {

			c.logger.Debug("Refresh", "Stream", raw.Stream, "Id", raw.ExternalID)
			bytes, err := json.Marshal(raw.Payload)
			if err != nil {
				c.logger.Error("marshalData", "Stream", raw.Stream, "Id", raw.ExternalID)
				continue
			}

			switch raw.Stream {
			case "addresses":
				err = json.Unmarshal(bytes, &addresses)
				for _, addr := range addresses {
					addrAcctm[addr] = accountId
				}
			case "rewards":
				ar := AccountReward{}
				err = json.Unmarshal(bytes, &ar)
				rewards = append(rewards, ar)
			case "transactions":

				if
				// strings.Compare(raw.ExternalID, "24c04a83a836641c1f8fab5c0414fd818cf2401a07ab3f52eef42ee4a9517a1d") == 0 ||
				// strings.Compare(raw.ExternalID, "1b0d6e6307c2fc81106a56f26a2a89ae2571468175f104a0c5980acd96438399") == 0 ||
				// strings.Compare(raw.ExternalID, "f7318a6e05b5f349fb6474808305e6dd4237306095c75e67f714cf4c77b1d2fc") == 0 ||
				strings.Compare(raw.ExternalID, "48db8f286cd33c304a401deb61236ddacc6346dbdb3fc4611cdd7f8ab59416f4") == 0 ||
					false {
					// log.Println(string(bytes))
				}
				txn := Transaction{}
				err = json.Unmarshal(bytes, &txn)
				txn.UID = raw.UID
				txn.AccountID = raw.AccountID
				c.logger.Debug("marshalData", "Id", raw.ExternalID, "Date", txn.BlockTime)

				//check for duplicates
				if _, ok := txhasm[txn.TxHash]; !ok {
					txns = append(txns, txn)
				}
			}
		}

		addrsm[accountId] = addresses
		rewardsm[accountId] = rewards
		txnsm[accountId] = txns

	}

	return addrAcctm, addrsm, rewardsm, txnsm
}

func (c CardanoAccountTransformer) marshalEpochData(raws []domain.RawItem) map[int64]EpochInformation {
	epochsm := make(map[int64]EpochInformation)
	for _, raw := range raws {
		bytes, err := json.Marshal(raw.Payload)
		if err != nil {
			c.logger.Error("marshalData", "Stream", raw.Stream, "Id", raw.ExternalID)
			continue
		}
		epoch := EpochInformation{}
		err = json.Unmarshal(bytes, &epoch)
		epochsm[epoch.Epoch] = epoch
	}
	return epochsm
}
