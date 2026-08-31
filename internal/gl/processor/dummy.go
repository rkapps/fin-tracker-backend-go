package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type DummyActivityProcessor struct {
	logger *logger.Logger
}

func NewDummyActivityProcessor(logConfig *logger.Config) DummyActivityProcessor {
	plog := logConfig.For("processor.dummy")
	return DummyActivityProcessor{logger: plog}
}

// ensures DummyActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*DummyActivityProcessor)(nil)

func (p DummyActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	return nil, nil
}
