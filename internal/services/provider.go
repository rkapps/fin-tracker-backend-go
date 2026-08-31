package services

// import (
// 	"context"

// 	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
// 	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
// )

// type ProviderService struct {
// 	storage storage.ProviderStorageService
// }

// func NewProviderService(storage storage.ProviderStorageService) ProviderService {
// 	return ProviderService{storage: storage}
// }

// func (c ProviderService) DeleteAllCursors(ctx context.Context, uid string, acctId string) error {
// 	return c.storage.DeleteAllCursors(ctx, uid, acctId)
// }

// func (c ProviderService) DeleteAllRawItems(ctx context.Context, uid string, acctId string) error {
// 	return c.storage.DeleteAllRawItems(ctx, uid, acctId)
// }

// func (c ProviderService) LoadAllCursors(ctx context.Context, uid string, acctId string) ([]*domain.SyncCursor, error) {
// 	return c.storage.LoadAllCursors(ctx, uid, acctId)
// }

// func (c ProviderService) LoadAllRawItems(ctx context.Context, uid string, acctId string) ([]*domain.RawItem, error) {
// 	return c.storage.LoadAllRawItems(ctx, uid, acctId)
// }

// // LoadCursor implements [storage.ProviderStorageService].
// func (c ProviderService) LoadCursor(ctx context.Context, uid, acctID, provider, stream string) (*domain.SyncCursor, error) {
// 	return c.storage.LoadCursor(ctx, uid, acctID, provider, stream)
// }

// // MarkProcessed implements [storage.ProviderStorageService].
// func (c ProviderService) MarkProcessed(ctx context.Context, rawIDs []string, transformVersion int) error {
// 	return c.storage.MarkProcessed(ctx, rawIDs, transformVersion)
// }

// // SaveCursor implements [storage.ProviderStorageService].
// func (c ProviderService) SaveCursor(ctx context.Context, cursor *domain.SyncCursor) error {
// 	return c.storage.SaveCursor(ctx, cursor)
// }

// // UnprocessedRaw implements [storage.ProviderStorageService].
// func (c ProviderService) UnprocessedRaw(ctx context.Context, uid, acctID string, transformVersion int) ([]domain.RawItem, error) {
// 	return c.storage.UnprocessedRaw(ctx, uid, acctID, transformVersion)
// }

// // UpsertRaw implements [storage.ProviderStorageService].
// func (c ProviderService) UpsertRaw(ctx context.Context, uid, acctID, provider string, items []domain.RawItem) error {
// 	return c.storage.UpsertRaw(ctx, uid, acctID, provider, items)
// }
