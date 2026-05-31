package services

import (
	"fmt"
	"time"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/google/uuid"
)

const (
	forecastDefaultMonths = 3
	forecastMinMonths     = 1
	forecastMaxMonths     = 12
)

// ForecastService projects the org's cash flow forward based on the current
// balance, unpaid payables/receivables and recurring movements.
type ForecastService struct {
	txs      *repositories.TransactionRepository
	accounts *repositories.AccountRepository
}

func NewForecastService(
	txs *repositories.TransactionRepository,
	accounts *repositories.AccountRepository,
) *ForecastService {
	return &ForecastService{txs: txs, accounts: accounts}
}

// monthName maps a month number to its pt-BR name for alert messages.
var monthNamePTBR = [...]string{
	"", "janeiro", "fevereiro", "março", "abril", "maio", "junho",
	"julho", "agosto", "setembro", "outubro", "novembro", "dezembro",
}

// Forecast builds an N-month cash-flow projection (months clamped to 1..12).
func (s *ForecastService) Forecast(orgID uuid.UUID, months int) (*dto.ForecastResponse, error) {
	if months < forecastMinMonths {
		months = forecastDefaultMonths
	}
	if months > forecastMaxMonths {
		months = forecastMaxMonths
	}

	// current_balance = sum of every account's computed (paid-only) balance.
	balances, err := s.accounts.Balances(orgID)
	if err != nil {
		return nil, err
	}
	var current int64
	for _, b := range balances {
		current += b
	}

	now := time.Now().UTC()
	// Buckets start at the 1st of next month and run forward N months.
	firstBucket := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)

	// Unpaid receivables/payables grouped by the YYYY-MM of their effective due
	// date, from the first bucket onward.
	inByMonth, err := s.txs.UnpaidByMonth(orgID, entities.TxIncome, firstBucket)
	if err != nil {
		return nil, err
	}
	outByMonth, err := s.txs.UnpaidByMonth(orgID, entities.TxExpense, firstBucket)
	if err != nil {
		return nil, err
	}

	// Recurring movements are NOT projected as a flat per-month amount here: the
	// RecurrenceRule engine materialises real future (unpaid) transactions, which
	// UnpaidByMonth already counts. Adding a flat projection on top would
	// double-count those rows, so the forecast relies solely on the materialised
	// rule-generated rows plus any other unpaid bills/receivables.

	resp := &dto.ForecastResponse{
		CurrentBalance: current,
		Months:         make([]dto.ForecastMonth, 0, months),
	}

	running := current
	lowestBal := current
	lowestMonth := ""
	for i := 0; i < months; i++ {
		bucket := firstBucket.AddDate(0, i, 0)
		key := bucket.Format("2006-01")

		inflow := inByMonth[key]
		outflow := outByMonth[key]
		net := inflow - outflow
		running += net

		resp.Months = append(resp.Months, dto.ForecastMonth{
			Month:            key,
			Inflow:           inflow,
			Outflow:          outflow,
			Net:              net,
			ProjectedBalance: running,
		})

		if i == 0 || running < lowestBal {
			lowestBal = running
			lowestMonth = key
		}
		if running < 0 {
			resp.Alerts = append(resp.Alerts, dto.ForecastAlert{
				Month:   key,
				Message: fmt.Sprintf("Saldo projetado negativo em %s", forecastMonthLabel(bucket)),
			})
		}
	}

	resp.EndBalance = running
	resp.Lowest = dto.ForecastLowest{Month: lowestMonth, Balance: lowestBal}
	return resp, nil
}

// forecastMonthLabel renders a "mês/ano" label in pt-BR for alert messages.
func forecastMonthLabel(t time.Time) string {
	return fmt.Sprintf("%s/%d", monthNamePTBR[int(t.Month())], t.Year())
}
