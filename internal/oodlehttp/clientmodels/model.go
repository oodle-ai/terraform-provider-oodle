package clientmodels

type ClientModel interface {
	GetID() string
	//MarshalJSON() ([]byte, error)
	//UnmarshalJSON([]byte) error
}
