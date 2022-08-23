package main

import (
	"errors"
	"fmt"
)

type StateManager struct {
	WarmUp    func(*Payout) error
	Validator func(*Payout) error
	Update    func(*Payout, *Payout) error
	CleanUp   func(*Payout) error
}

var (
	ErrInvalid = errors.New("invalid status")
)

// STATE MACHINE CONFIG - Should use builder methods in the future
var (
	movements = map[PayoutType]map[PayoutStatus]map[PayoutEventType]map[PayoutEventStatus]map[PayoutEventType]map[PayoutEventStatus]*StateManager{
		PayoutPix: {
			PayoutStatusProcessing: {
				EventCounterparty: {
					EventStatusProcessing: {
						EventCounterparty: {
							EventStatusSuccess: &StateManager{
								Validator: func(p *Payout) error {
									if p.Counterparty == nil {
										return errors.New("empty counterparty")
									}
									return nil
								},
								Update: func(p1, p2 *Payout) error {
									p1.Counterparty = p2.Counterparty
									return nil
								},
							},
						},
					},
					EventStatusSuccess: {
						EventWithdraw: {
							EventStatusSuccess: &StateManager{
								Validator: func(p *Payout) error {
									if p.Withdraw == nil {
										return errors.New("withdraw is empty")
									}
									return nil
								},
								Update: func(p1, p2 *Payout) error {
									p1.Withdraw = p2.Withdraw
									return nil
								},
								CleanUp: func(p *Payout) error {
									p.Status = PayoutStatusCompleted
									return nil
								},
							},
						},
					},
				},
			},
		},
	}
)

func Warmup(payout *Payout, event PayoutEventType, eventStatus ...PayoutEventStatus) error {
	payout, err := GetPayout(payout.Id)

	if err != nil {
		return err
	}

	states, err := possibleStates(payout, event)
	fmt.Println("possibleStates ", states)
	if err != nil {
		return fmt.Errorf("state machine warmup failed: %w", err)
	}

	for _, status := range eventStatus {
		if state, ok := states[status]; ok {
			if state.WarmUp != nil {
				if err := state.WarmUp(payout); err != nil {
					return err
				}
			}
			continue
		}
		return ErrInvalid
	}

	return nil
}

func Set(payout *Payout) error {
	payoutDb, err := GetPayout(payout.Id)

	if err != nil {
		return err
	}

	newEvents, err := newEvents(payoutDb, payout)

	if err != nil {
		return err
	}

	for _, ev := range newEvents {
		state, err := next(payoutDb, ev.Type, ev.Status)

		if err != nil {
			return err
		}

		if state.Update != nil {
			if err := state.Update(payoutDb, payout); err != nil {
				return err
			}
		}

		if err := state.Validator(payoutDb); err != nil {
			return err
		}

		if state.CleanUp != nil {
			if err := state.CleanUp(payoutDb); err != nil {
				return err
			}
		}

		if len(payoutDb.Events) > 0 && payoutDb.Events[len(payoutDb.Events)-1].Id == ev.Id {
			payoutDb.Events[len(payoutDb.Events)-1] = ev
		} else {
			payoutDb.Events = append(payoutDb.Events, ev)
		}
	}

	return SavePayout(payoutDb)
}

func newEvents(payoutDb *Payout, payout *Payout) ([]PayoutEvent, error) {
	if payoutDb.Events == nil {
		return payout.Events, nil
	}

	if len(payout.Events) == len(payoutDb.Events) {
		return []PayoutEvent{payout.Events[len(payout.Events)-1]}, nil
	}

	return payout.Events[len(payoutDb.Events):], nil
}

func next(payout *Payout, eventType PayoutEventType, status PayoutEventStatus) (*StateManager, error) {
	states, err := possibleStates(payout, eventType)

	if err != nil {
		return nil, err
	}

	if state, ok := states[status]; ok {
		return state, nil
	}

	return nil, ErrInvalid
}

func possibleStates(payout *Payout, eventType PayoutEventType) (map[PayoutEventStatus]*StateManager, error) {

	if len(payout.Events) == 0 {
		return nil, ErrInvalid
	}

	lastEvent := payout.Events[len(payout.Events)-1]
	if next, ok := movements[payout.Type]; ok {
		if next, ok := next[payout.Status]; ok {
			if next, ok := next[lastEvent.Type]; ok {
				if next, ok := next[lastEvent.Status]; ok {
					if states, ok := next[eventType]; ok {
						return states, nil
					}
				}
			}
		}
	}

	return nil, ErrInvalid
}
