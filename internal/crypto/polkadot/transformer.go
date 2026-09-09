package polkadot

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type PolkadotTransformer struct {
	logger *logger.Logger
}

func NewPolkadotTransformer(logConfig *logger.Config) PolkadotTransformer {
	plog := logConfig.For("refresher.polkadot")
	return PolkadotTransformer{plog}
}

func (p PolkadotTransformer) Name() string {
	return "polkadot"
}

func (p PolkadotTransformer) Transform(ctx context.Context, ps core.PriceService,
	spamService core.CryptoSpamService,
	gaccts []*domain.Account,
	acreds []domain.AccountWithCredential,
	globalRaws []domain.RawItem,
	rawsm map[string][]domain.RawItem,
) ([]*domain.Activity, error) {

	actvs := []*domain.Activity{}

	// gather the polkadot accounts
	paccts := []domain.Account{}
	for _, acred := range acreds {
		paccts = append(paccts, acred.Account)
	}

	rewards, tsfrs := p.marshalData(rawsm)

	for _, reward := range rewards {
		p.logger.Debug("Transform", "Reward", fmt.Sprintf("%s %v", reward.Event_Index, reward.Amount))
		actv := p.createRewardActivity(ps, reward)
		if actv != nil {
			actvs = append(actvs, actv)
		}
	}

	tsfrm := make(map[string]PolkadotTransfer)
	for _, tsfr := range tsfrs {
		if _, ok := tsfrm[tsfr.Hash]; ok {
			continue
		}
		tsfrm[tsfr.Hash] = tsfr
		actv := p.createTransferActivity(ps, paccts, tsfr)
		if actv != nil {
			actvs = append(actvs, actv)
		}
		p.logger.Debug("Transform", "Transfer", tsfr.Hash)
	}

	p.logger.Info("Transform", "Provider", p.Name(), "Activities", len(actvs))

	return actvs, nil
}

func (p PolkadotTransformer) marshalData(rawsm map[string][]domain.RawItem,
) (
	[]PolkadotReward,
	[]PolkadotTransfer,
) {

	var rewards []PolkadotReward
	var tsfrs []PolkadotTransfer

	for _, raws := range rawsm {
		for _, raw := range raws {

			p.logger.Debug("Refresh", "Stream", raw.Stream, "Id", raw.ExternalID)
			bytes, err := json.Marshal(raw.Payload)
			if err != nil {
				p.logger.Error("marshalData", "Stream", raw.Stream, "Id", raw.ExternalID)
				continue
			}

			switch raw.Stream {
			case "rewards":
				pr := PolkadotReward{}
				err = json.Unmarshal(bytes, &pr)
				pr.UID = raw.UID
				pr.AccountId = raw.AccountID
				rewards = append(rewards, pr)
			case "transfers":

				tsfr := PolkadotTransfer{}
				err = json.Unmarshal(bytes, &tsfr)
				tsfr.UID = raw.UID
				tsfr.AccountId = raw.AccountID
				tsfrs = append(tsfrs, tsfr)
			}
		}
	}

	return rewards, tsfrs
}
