package hostserve

import "github.com/google/uuid"

func NewUUID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		id = uuid.New()
	}
	return id
}
