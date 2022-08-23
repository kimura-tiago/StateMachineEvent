package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	EventCounterparty PayoutEventType = "COUNTERPARTY"
	EventWithdraw     PayoutEventType = "WITHDRAW"
	EventExchange     PayoutEventType = "EXCHANGE"

	EventStatusSuccess    PayoutEventStatus = "SUCCESS"
	EventStatusProcessing PayoutEventStatus = "PROCESSING"
	EventStatusError      PayoutEventStatus = "ERROR"

	PayoutTed PayoutType = "TED"
	PayoutPix PayoutType = "PIX"

	PayoutStatusProcessing PayoutStatus = "PROCESSING"
	PayoutStatusCompleted  PayoutStatus = "COMPLETED"
)

type PayoutType string

type Payout struct {
	Id           string
	Amount       float64
	Status       PayoutStatus
	Type         PayoutType
	Events       []PayoutEvent
	Counterparty *Counterparty
	Withdraw     *Withdraw
}

type PayoutEvent struct {
	Id     string
	Type   PayoutEventType
	Status PayoutEventStatus
}

type PayoutEventStatus string

type PayoutEventType string

type PayoutStatus string

type Counterparty struct {
	Name string
	Age  int
}

type Withdraw struct {
	Amount    float64
	CreatedAt time.Time
}

func main() {
	fmt.Println("begin")
	payout := Payout{
		Id:     uuid.NewString(),
		Amount: 10,
		Status: PayoutStatusProcessing,
		Type:   PayoutPix,
		Events: []PayoutEvent{
			{
				Id:     uuid.NewString(),
				Type:   EventCounterparty,
				Status: EventStatusProcessing,
			},
		},
	}

	if err := SavePayout(&payout); err != nil {
		panic(err)
	}

	if err := Warmup(&payout, EventCounterparty); err != nil {
		panic(err)
	}

	CounterpartyBusinessLogic(&payout)

	if err := Set(&payout); err != nil {
		panic(err)
	}

	if err := Warmup(&payout, EventWithdraw); err != nil {
		panic(err)
	}

	WithdrawBusinessLogic(&payout)

	if err := Set(&payout); err != nil {
		panic(err)
	}
}

// BUSINESS

func WithdrawBusinessLogic(payout *Payout) {
	payout.Withdraw = &Withdraw{
		Amount:    payout.Amount,
		CreatedAt: time.Now(),
	}
	payout.Events = append(payout.Events, PayoutEvent{
		Id:     uuid.NewString(),
		Type:   EventWithdraw,
		Status: EventStatusSuccess,
	})
}

func CounterpartyBusinessLogic(payout *Payout) {
	payout.Counterparty = &Counterparty{
		Name: "Murillo",
		Age:  18,
	}

	payout.Events[len(payout.Events)-1].Status = EventStatusSuccess
}
