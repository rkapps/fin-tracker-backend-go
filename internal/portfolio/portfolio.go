package portfolio

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/gl"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"golang.org/x/sync/errgroup"
)

type Portfolio struct {
	userStorage         storage.UserStorageService
	accountsStorage     storage.AccountStorageService
	tickersStorage      storage.TickerStorageService
	providerStorage     storage.ProviderStorageService
	cryptoStorage       storage.CryptoStorageService
	encryptionService   core.EncryptionService // Encrypt/Decrypt
	priceService        core.PriceService
	spamService         core.CryptoSpamService
	syncRegistry        core.SyncRegistry
	transformerRegistry core.TransformerRegistry
	logger              *logger.Logger
	logConfig           *logger.Config
}

func NewPortfolio(
	userStorage storage.UserStorageService,
	accountsStorage storage.AccountStorageService,
	tickersStorage storage.TickerStorageService,
	providerStorage storage.ProviderStorageService,
	cryptoStorage storage.CryptoStorageService,
	encryptionService core.EncryptionService, // Encrypt/Decrypt
	syncRegistry core.SyncRegistry,
	transformerRegistry core.TransformerRegistry,
	logConfig *logger.Config,
	logger *logger.Logger,
) Portfolio {
	plog := logConfig.For("portfolio")

	ps := core.NewPriceService(tickersStorage, cryptoStorage)
	spamService := core.NewCryptoSpamService(cryptoStorage)

	ps.LoadCryptoPrices()
	spamService.LoadCryptoSpams()

	return Portfolio{
		userStorage:         userStorage,
		accountsStorage:     accountsStorage,
		tickersStorage:      tickersStorage,
		providerStorage:     providerStorage,
		cryptoStorage:       cryptoStorage,
		priceService:        ps,
		spamService:         spamService,
		encryptionService:   encryptionService,
		syncRegistry:        syncRegistry,
		transformerRegistry: transformerRegistry,
		logConfig:           logConfig,
		logger:              plog,
	}
}

func (p Portfolio) RefreshUserAccounts(ctx context.Context, uid string, simulate bool) error {

	var err error
	user, err := p.userStorage.GetUser(uid)
	if err != nil {
		return fmt.Errorf("User record does not exist")
	}

	p.logger.Trace("RefreshUserAccounts", "storage", p.tickersStorage)
	p.logger.Info("RefreshUserAccounts", "UID", uid, "CurrencyCode", user.CurrencyCode)
	p.logger.Trace("RefreshUserAccounts", "UID", uid, "Simulate", simulate)
	accts, err := p.accountsStorage.GetAccounts(uid)
	if err != nil {
		return fmt.Errorf("error getting user accounts: %v", err)
	}
	p.logger.Info("RefreshUserAccounts", "UID", uid, "Accounts", len(accts))
	actvs, err := p.refreshUserActivities(ctx, accts)
	if err != nil {
		p.logger.Error("RefreshUserAccounts", "Error", err)
		return fmt.Errorf("error refreshing user activities")
	}

	// adjust actvities
	p.AdjustActivities(uid, actvs)

	// run gainloss
	p.logger.Info("RefreshUserAccounts", "Activities", len(actvs))
	gl := gl.NewGainLoss(*user, accts, simulate, p.logConfig)

	glResult, err := gl.Run(ctx, actvs)
	if err != nil {
		p.logger.Error("RefreshUserAccounts", "Run", err)
		return fmt.Errorf("error running gainloss")
	}

	cprices := p.priceService.GetCryptoPrices()
	p.logger.Info("RefresUserAccounts", "CryptoPrices", len(cprices))
	// store crypto prices
	p.cryptoStorage.SaveCryptoPrices(cprices)

	asumys, err := p.summarizeData(uid, accts, glResult.Actvs, glResult.Lots)
	if err != nil {
		p.logger.Error("RefreshUserAccounts", "SummarizeData", err)
		return fmt.Errorf("error summarizing data")
	}

	if !simulate {
		err = p.saveData(uid, asumys, glResult.Actvs, glResult.Lots, gl.GLEntries)
	}
	if err != nil {
		p.logger.Error("RefreshUserAccounts", "Error", err)
	}

	// gain loss here
	return err
}

func (p Portfolio) refreshUserActivities(ctx context.Context, accts []*domain.Account) ([]*domain.Activity, error) {

	acredsm := p.groupAccountsWithCredentialsByProvider(accts)
	p.logger.Info("refreshUserActivities", "Grouped Accounts", len(acredsm))

	var (
		mu         sync.Mutex
		activities []*domain.Activity
	)

	g, ctx := errgroup.WithContext(ctx)

	for provider, acreds := range acredsm {
		// resync := false
		// if strings.Compare(provider, "ethereum") == 0 || strings.Compare(provider, "binance") == 0 {
		// 	resync = true
		// }

		// if !resync {
		// 	for _, acred := range acreds {
		// 		aactvs, _ := p.accountsStorage.GetActivitiesForAccount(acred.Account.UID, acred.Account.ID)
		// 		// p.logger.Info("refreshUserActivities", "Provider", acred.Account.ID, "Refreshed", len(aactvs))
		// 		activities = append(activities, aactvs...)
		// 	}
		// 	continue
		// }

		g.Go(func() error {

			importRefresher := NewImportAccountRefresher(p.accountsStorage, p.logConfig)
			result, err := importRefresher.Refresh(ctx, p.priceService, accts, acreds, p.logConfig)
			if err != nil {
				p.logger.Error("refreshUserActivities", "NewImportAccountRefresher", err)
			}
			p.logger.Info("refreshUserActivities", "Provider", provider, "Imported", len(result))
			if len(result) > 0 {
				mu.Lock()
				activities = append(activities, result...)
				mu.Unlock()
			}

			transformer, err := ResolveTransformer(p.transformerRegistry, provider)
			if transformer != nil {
				actvs, err := Refresh(ctx, p.priceService, p.spamService, p.providerStorage, accts, acreds, transformer)
				if err != nil {
					p.logger.Error("refreshUserActivities", "provider", err)
				}
				if len(actvs) > 0 {
					mu.Lock()
					activities = append(activities, actvs...)
					mu.Unlock()
				}
			}

			// slog.Debug("refreshUserActivities", "refresher", refresher, "error", err)
			// result, err := refresher.Refresh(ctx, account, p.logConfig)

			// if err != nil {
			// 	return err
			// }
			// mu.Lock()
			// activities = append(activities, result...)
			// mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return activities, nil
}

func (p Portfolio) AdjustActivities(uid string, actvs []*domain.Activity) {

	adjs, _ := p.accountsStorage.GetActivityAdjs(uid)
	adjm := make(map[string]domain.ActivityAdj)
	for _, actv := range adjs {
		adjm[actv.ID] = *actv
	}

	for _, actv := range actvs {
		adjActv := adjm[actv.ID]
		if len(adjActv.ID) == 0 {
			continue
		}
		if len(adjActv.TxnType) > 0 {
			actv.TxnType = domain.ActivityType(adjActv.TxnType)
		}
		p.logger.Info("AdjustActivities", "Actv", actv.Hash, "seonds", adjActv.AdjustSeconds)

		adjSeconds, err := strconv.Atoi(adjActv.AdjustSeconds)
		if err == nil {
			actv.Date = actv.Date.Add(time.Second * time.Duration(adjSeconds))
		}
	}
}

// group accounts by provider, paired with their matching credential
func (p Portfolio) groupAccountsWithCredentialsByProvider(accts domain.Accounts) map[string][]domain.AccountWithCredential {

	grouped := make(map[string][]domain.AccountWithCredential)

	for _, account := range accts {

		config := json.RawMessage{}
		cred, _ := p.accountsStorage.GetAccountCredential(account.UID, account.ID)
		if cred != nil {
			config, _ = p.encryptionService.Decrypt(cred.EncryptedConfig)
		}

		acred := domain.AccountWithCredential{
			Account: *account,
			Config:  config,
		}

		provider := account.ProviderName()
		acreds := grouped[provider]
		if len(acreds) == 0 {
			acreds = []domain.AccountWithCredential{}
		}
		acreds = append(acreds, acred)
		grouped[provider] = acreds
	}
	return grouped
}
