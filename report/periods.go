package report

import (
	"context"
	"time"

	"github.com/figment-networks/cosmos-extract/client"
)

type period struct {
	startTime     time.Time
	nextStartTime time.Time
	endHeight     uint64
}

func (r *runner) buildOrderedPeriods(
	ctx context.Context,
	startTime,
	endTime time.Time,
) (periods []period, err error) {

	// We want to get the last height for each of the months that fall within the provided time range,
	// regardless of the specific times. Start by calculating the number of months we need heights for.
	startYear := startTime.Year()
	startMonth := int(startTime.Month())
	endYear := endTime.Year()
	endMonth := int(endTime.Month())
	numMonths := (endYear-startYear)*12 + endMonth - startMonth + 1

	periods = make([]period, numMonths)
	for i := 0; i < numMonths; i++ {
		// Use the current month to get the first time of the next month. We will pass this "before time"
		// as an argument to indicate we want the last height that occurred before this time.
		currMonthStartTime := time.Date(startYear, time.Month(startMonth+i), 1, 0, 0, 0, 0, time.UTC)
		nextMonthStartTime := time.Date(startYear, time.Month(startMonth+i+1), 1, 0, 0, 0, 0, time.UTC)
		req := client.LastHeightBeforeReq{
			Network:    network,
			ChainID:    chainID,
			BeforeTime: nextMonthStartTime,
		}

		var height uint64
		height, err = r.client.GetLastHeightBefore(ctx, req)
		if err != nil {
			return
		}

		periods[i] = period{
			startTime:     currMonthStartTime,
			nextStartTime: nextMonthStartTime,
			endHeight:     height,
		}
	}

	return
}
