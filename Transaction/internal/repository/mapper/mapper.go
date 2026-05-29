package mapper

import (
	"study/Transaction/internal/model"
	repomodel "study/Transaction/internal/repository/model"
)

func RepoTransactionsToModel(repoTrans []repomodel.Transaction) []model.Transaction {
	transactions := make([]model.Transaction, 0, len(repoTrans))

	for _, rt := range repoTrans {
		addTran := RepoTranToModel(rt)
		transactions = append(transactions, addTran)
	}
	return transactions
}

func RepoTrEntriesToModel(rte []repomodel.TransactionEntry) []model.TransactionEntry {
	tEntries := make([]model.TransactionEntry, 0, len(rte))

	for _, te := range rte {
		addTran := RepoTransEntryToModel(te)
		tEntries = append(tEntries, addTran)
	}
	return tEntries
}

func RepoTransEntryToModel(repoTrEnt repomodel.TransactionEntry) model.TransactionEntry {
	return model.TransactionEntry{
		ID:            repoTrEnt.ID,
		TransactionID: repoTrEnt.TransactionID,
		AccountID:     repoTrEnt.AccountID,
		Direction:     model.TransactionEntryDirection(repoTrEnt.Direction),
		Amount:        repoTrEnt.Amount,
		CreatedAt:     repoTrEnt.CreatedAt,
		UpdatedAt:     repoTrEnt.UpdatedAt,
	}
}

func RepoTranToModel(rt repomodel.Transaction) model.Transaction {
	return model.Transaction{
		ID:        rt.ID,
		UserID:    rt.UserID,
		Amount:    rt.Amount,
		Status:    model.TransactionStatus(rt.Status),
		Type:      model.TransactionType(rt.Type),
		CreatedAt: rt.CreatedAt,
		UpdatedAt: rt.UpdatedAt,
	}
}

func TransactionToRepo(tr model.Transaction) repomodel.Transaction {
	return repomodel.Transaction{
		UserID: tr.UserID,
		Amount: tr.Amount,
		Status: string(tr.Status),
		Type:   string(tr.Type),
	}
}

func TransactionEntryToRepo(te model.TransactionEntry) repomodel.TransactionEntry {
	return repomodel.TransactionEntry{
		ID:            te.ID,
		TransactionID: te.TransactionID,
		AccountID:     te.AccountID,
		Direction:     string(te.Direction),
		Amount:        te.Amount,
		CreatedAt:     te.CreatedAt,
		UpdatedAt:     te.UpdatedAt,
	}
}

func UpdateTransactionToRepo(upTr model.UpdateTransaction) repomodel.UpdateTransaction {
	return repomodel.UpdateTransaction{
		Status: string(upTr.Status),
	}
}
