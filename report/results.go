package report

import "math/big"

type results []accountResults

type accountResults struct {
	account         string
	durationResults []durationResult
}

type durationResult struct {
	duration   string
	validators []string
	netRewards []*big.Int
	netFees    []*big.Int
}

func (dr durationResult) grossRewards(validatorIdx int) *big.Int {
	vi := validatorIdx
	return (&big.Int{}).Sub(dr.netRewards[vi], dr.netFees[vi])
}
