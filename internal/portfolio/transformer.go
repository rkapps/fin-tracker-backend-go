package portfolio

import (
	"context"
	"log/slog"

	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
)

const CurrentVersion = 1

func ResolveTransformer(transformerRegistry core.TransformerRegistry, provider string) (core.TransformerProvider, error) {
	return transformerRegistry.Get(provider)
}

// Refresh drives a transformer over unprocessed rows. Deterministic IDs +
// upsert make it safe to re-run; bumping CurrentVersion re-processes history
// without re-fetching.
func Refresh(ctx context.Context,
	ps core.PriceService,
	storage storage.ProviderStorageService, gaccts []*domain.Account, acreds []domain.AccountWithCredential, tf core.TransformerProvider,
) ([]*domain.Activity, error) {

	var actvs []*domain.Activity
	var err error
	rawsm := make(map[string][]domain.RawItem)

	globalRaws, err := storage.UnprocessedRaw(ctx, tf.Name(), tf.Name(), CurrentVersion)
	slog.Debug("Refresh", "Global", tf.Name(), "Items", len(globalRaws))
	if err != nil {
		return actvs, err
	}

	for _, acred := range acreds {
		araws, err := storage.UnprocessedRaw(ctx, acred.Account.UID, acred.Account.ID, CurrentVersion)
		slog.Debug("Refresh", "Account", acred.Account.ID, "Items", len(araws))
		if err != nil {
			return actvs, err
		}
		rawsm[acred.Account.ID] = araws
	}
	actvs, err = tf.Transform(ctx, ps, gaccts, acreds, globalRaws, rawsm)
	return actvs, err

	// if err := st.UpsertActivities(ctx, acts); err != nil {
	// 	return err
	// }
	// ids := make([]int64, len(raws))
	// for i, r := range raws {
	// 	ids[i] = r.ID
	// }
	// return st.MarkProcessed(ctx, ids, CurrentVersion)
}
