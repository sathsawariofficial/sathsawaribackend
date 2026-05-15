package passenger

import (
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
)

func mapBookSeatData(request BookSeatDemandRequest) postgress.BookSeatDemand {
	return postgress.BookSeatDemand{
		ID:           utils.GenerateUUID(),
		RideId:       request.RideId,
		Name:         request.Name,
		Mobilenumber: request.Mobilenumber,
		Message:      request.Message,
		IsHandled:    false,
	}
}
