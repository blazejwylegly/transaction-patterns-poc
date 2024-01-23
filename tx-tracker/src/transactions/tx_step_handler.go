package transactions

import (
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

type TxnStepHandler struct {
	db *gorm.DB
}

func (handler TxnStepHandler) HandleTxnStep(txnContext TxnContext, step *db.TransactionStep) {
	err := handler.db.Transaction(func(tx *gorm.DB) error {
		var txn *db.Transaction

		r := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(&db.Transaction{TxnId: txnContext.TxnId}).
			Find(&txn)
		err := r.Error

		if err != nil {
			return err
		}

		// Tx not found -> create new one
		if r.RowsAffected == 0 {
			txn = &db.Transaction{
				TxnId:     txnContext.TxnId,
				TxnName:   txnContext.TxnName,
				StartedAt: txnContext.TxnStartedAt,
				Steps:     []db.TransactionStep{},
				Status:    step.StepStatus,
			}
		}
		err = tx.Save(txn).Error
		if err != nil {
			return err
		}

		// Add new tx step
		txn.Steps = append(txn.Steps, *step)
		err = tx.Save(step).Error
		if err != nil {
			return err
		}

		txn.Status = determineTransactionStatus(txn.Steps)
		err = tx.Save(txn).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Printf("Error trying to save step %s for txn %s",
			step.StepName,
			txnContext.TxnId.String())
	}
}

func determineTransactionStatus(steps []db.TransactionStep) string {
	for _, step := range steps {
		if step.StepStatus == "FAILED" {
			return step.StepStatus
		}
	}
	return "SUCCESS"
}

func NewTxnStepHandler(db *gorm.DB) *TxnStepHandler {
	return &TxnStepHandler{db: db}
}
