package report

import (
	"encoding/csv"
	"math/big"
	"os"
	"time"
)

type accountResults map[string][]durationResult

type durationResult struct {
	duration time.Time
	// Tracks unique validators from all of the fields below.
	validators map[string]bool
	// Validator address as the key.
	delegations map[string]*big.Int
	rewards     map[string]*big.Int
	fees        map[string]*big.Int
}

func (dr durationResult) netRewards(validator string) *big.Int {
	v := validator
	return (&big.Int{}).Sub(dr.rewards[v], dr.fees[v])
}

func initAccountResults(accounts []string) accountResults {

	accountResults := map[string][]durationResult{}
	for _, a := range accounts {
		accountResults[a] = []durationResult{}
	}

	return accountResults
}

func initDurationResult(duration time.Time) durationResult {
	return durationResult{
		duration:    duration,
		validators:  map[string]bool{},
		delegations: map[string]*big.Int{},
		rewards:     map[string]*big.Int{},
		fees:        map[string]*big.Int{},
	}
}

func (ar accountResults) writeToDisk(path string) error {

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	cw := csv.NewWriter(f)
	headers := []string{"account", "date", "validator", "delegation", "gross_rewards", "fees", "net_rewards"}
	if err := cw.Write(headers); err != nil {
		return err
	}
	defer cw.Flush()

	// Write all the data and skip nil big ints.
	for acc, durationResults := range ar {
		for _, result := range durationResults {
			date := result.duration.Format("2006-01")
			for v := range result.validators {

				values := []string{acc, date, v}
				var delegation, rewards, fees, netRewards string

				if value := result.delegations[v]; value != nil {
					delegation = value.String()
				}
				if value := result.rewards[v]; value != nil {
					rewards = value.String()
				}
				if value := result.fees[v]; value != nil {
					fees = value.String()
				}

				// The latter two cases should never happen but it felt more complete
				// to include them.
				if rewards != "" && fees != "" {
					diff := big.NewInt(0)
					diff = diff.Sub(result.rewards[v], result.fees[v])
					netRewards = diff.String()
				} else if rewards != "" {
					netRewards = rewards
				} else if fees != "" {
					netRewards = fees
				}

				values = append(values, delegation, rewards, fees, netRewards)
				if err = cw.Write(values); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
