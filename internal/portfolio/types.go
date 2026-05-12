package portfolio

import (
	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
)

type Portfolio struct {
	storage       storage.StorageService
	logger        *logger.Logger
	acctLotSeqMap map[string]int // scoped to one GL run

}

func NewPortfolio(storage storage.StorageService, logger *logger.Logger) Portfolio {
	acctLotSeqm := make(map[string]int)
	return Portfolio{storage: storage, logger: logger, acctLotSeqMap: acctLotSeqm}
}
