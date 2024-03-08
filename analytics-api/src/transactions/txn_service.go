package transactions

import (
	"errors"
	"fmt"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

type TxnService struct {
	db *gorm.DB
}

func (handler *TxnService) HandleFinalTxnStep(txnContext TxnContext, finalTxnStep *db.TransactionStep) {
	err := handler.db.Transaction(func(tx *gorm.DB) error {
		var txn *db.Transaction

		r := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Steps").
			Where(&db.Transaction{TxnId: txnContext.TxnId}).
			Find(&txn)
		err := r.Error

		if err != nil {
			return err
		}

		// Tx not found -> error
		if r.RowsAffected == 0 {
			return errors.New(fmt.Sprintf("Step with id %s is already registered!", finalTxnStep.StepId))
		}

		// Check if final step entry already exists
		for _, existingStep := range txn.Steps {
			if existingStep.StepId == finalTxnStep.StepId {
				return errors.New(fmt.Sprintf("Step with id %s is already registered!", existingStep.StepId))
			}
		}

		// Add new tx newTxnStep
		txn.Steps = append(txn.Steps, *finalTxnStep)
		err = tx.Save(finalTxnStep).Error
		if err != nil {
			return err
		}

		// Determine status of whole transaction
		txn.Status = txnContext.TxnStatus
		err = tx.Save(txn).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Printf("Error trying to save newTxnStep %s for txn %s",
			finalTxnStep.StepName,
			txnContext.TxnId.String())
	}
}

func (handler *TxnService) HandleTxnStep(txnContext TxnContext, newTxnStep *db.TransactionStep) {
	err := handler.db.Transaction(func(tx *gorm.DB) error {
		var txn *db.Transaction

		r := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Steps").
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
				Status:    "PENDING",
			}
			err = tx.Save(txn).Error
			if err != nil {
				return err
			}
		}

		// Check if newTxnStep already exists
		for _, existingStep := range txn.Steps {
			if existingStep.StepId == newTxnStep.StepId {
				return errors.New(fmt.Sprintf("Step with id %s is already registered!", existingStep.StepId))
			}
		}

		// Add new tx newTxnStep
		txn.Steps = append(txn.Steps, *newTxnStep)
		err = tx.Save(newTxnStep).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Printf("Error trying to save newTxnStep %s for txn %s",
			newTxnStep.StepName,
			txnContext.TxnId.String())
	}
}

func NewTxnService(db *gorm.DB) *TxnService {
	return &TxnService{db: db}
}
