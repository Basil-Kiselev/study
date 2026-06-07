package mapper

import (
	"strconv"
	"study/Transaction/internal/model"
	transpb "study/contracts/transaction"
)

func PbToGetTransRequest(req *transpb.GetTransactionsRequest) model.GetTransactionsParams {
	dateFrom := req.DateFrom.AsTime()
	dateTo := req.DateTo.AsTime()

	return model.GetTransactionsParams{
		UserID:   req.UserId,
		Type:     req.Type,
		Status:   req.Status,
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
		Limit:    int(req.Pagination.Limit),
		Offset:   int(req.Pagination.Offset),
	}
}

func TransactionDetailsToPb(td model.TransactionDetails) transpb.TransactionDetails {
	strID := strconv.Itoa(int(td.Transaction.ID))
	return transpb.TransactionDetails{
		Transaction: &transpb.Transaction{
			Id:     strID,
			Type:   string(td.Transaction.Type),
			Status: string(td.Transaction.Status),
		},
		Entries: changeEntriesToPb(td.Entries),
	}
}

func changeEntriesToPb(entries []model.TransactionEntry) []*transpb.TransactionEntry {
	res := make([]*transpb.TransactionEntry, len(entries))

	for i, entry := range entries {
		strID := strconv.Itoa(int(entry.ID))
		transStrID := strconv.Itoa(int(entry.TransactionID))
		accID := strconv.Itoa(int(entry.AccountID))
		res[i] = &transpb.TransactionEntry{
			Id:            strID,
			TransactionId: transStrID,
			AccountId:     accID,
			Direction:     string(entry.Direction),
			Amount:        entry.Amount,
		}
	}

	return res
}

func TransDetailsToPb(tds []model.TransactionDetails) []*transpb.TransactionDetails {
	res := make([]*transpb.TransactionDetails, 0, len(tds))

	for _, td := range tds {
		newTd := TransactionDetailsToPb(td)
		res = append(res, &newTd)
	}

	return res
}
