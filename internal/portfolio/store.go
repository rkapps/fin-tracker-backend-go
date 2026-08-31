package portfolio

import (
	"log"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

func (p Portfolio) saveData(uid string, asumys []*domain.AccountSummary, actvs []*domain.Activity,
	lots []*domain.ActivityLot, glEntries []*domain.GLEntry,
) error {

	p.logger.Info("SaveData", "Activities", len(actvs))
	p.logger.Info("SaveData", "ActivityLots", len(lots))
	p.logger.Info("SaveData", "GlEntries", len(glEntries))

	ids := []string{}

	// clear and delete activitysummaries
	clear(ids)
	oasumys, err := p.accountsStorage.GetAccountSummaries(uid)
	if err != nil {
		return err
	}
	for _, oasum := range oasumys {
		ids = append(ids, oasum.ID)
	}
	p.accountsStorage.DeleteAccountSummaries(ids)

	clear(ids)
	// get and delete activities
	oactvs, _ := p.accountsStorage.GetActivities(uid)
	for _, oactv := range oactvs {
		ids = append(ids, oactv.ID)
	}
	p.accountsStorage.DeleteActivities(ids)
	err = p.accountsStorage.SaveActivities(actvs)
	if err != nil {
		log.Println(err)
		for _, actv := range actvs {
			if actv.Value.GreaterThan(decimal.NewFromInt(1833)) && actv.Value.LessThan(decimal.NewFromInt(1834)) {
				log.Println(actv.ID)
			}
		}
		p.logger.Error("SaveData", "SaveActivities", err)
		return err
	}

	// clear and delete activitylogs
	clear(ids)
	olots, _ := p.accountsStorage.GetActivityLots(uid)
	for _, olot := range olots {
		ids = append(ids, olot.ID)
	}

	err = p.accountsStorage.DeleteActivityLots(ids)
	if err != nil {
		return err
	}

	err = p.accountsStorage.SaveAccountSummaries(asumys)
	if err != nil {
		return err
	}

	err = p.accountsStorage.SaveActivityLots(lots)

	//Gl entries
	oglEntries, _ := p.accountsStorage.GetGlEntries(uid)
	clear(ids)
	for _, entry := range oglEntries {
		ids = append(ids, entry.ID)
	}
	err = p.accountsStorage.DeleteGlEntries(ids)
	if err != nil {
		return err
	}
	p.accountsStorage.SaveGlEntries(glEntries)

	return err
}
