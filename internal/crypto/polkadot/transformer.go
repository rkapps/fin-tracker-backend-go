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
	rewards, tsfrs := p.marshalData(rawsm)

	for _, reward := range rewards {
		p.logger.Info("Transform", "Reward", fmt.Sprintf("%s %v", reward.Event_Index, reward.Amount))
	}
	for _, tsfr := range tsfrs {
		p.logger.Info("Transform", "Transfer", tsfr.Hash)
	}
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
				pr.UID = raw.UID
				pr.AccountId = raw.AccountID
				err = json.Unmarshal(bytes, &pr)
				rewards = append(rewards, pr)
			case "tsfrs":

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
