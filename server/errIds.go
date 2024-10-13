package server

const (
	ErrIdBadRequest = iota
	ErrIdDbErr
	ErrIdDataNotFound
	ErrIdJsonMarshal
	ErrIdNotApproved
	ErrIdCantExtendIntoPast
	ErrIdAlreadyExists
)
