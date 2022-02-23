package request

import (
	"encoding/json"
	"fmt"

	"github.com/gojek/turing/api/turing/models"
)

// CreateOrUpdateEnsemblerRequest is the request to
// update or create an ensembler
type CreateOrUpdateEnsemblerRequest struct {
	models.EnsemblerLike
}

// UnmarshalJSON is a function to unmarshal the json into a go object
func (r *CreateOrUpdateEnsemblerRequest) UnmarshalJSON(data []byte) error {
	typeCheck := struct {
		Type models.EnsemblerType
	}{}

	if err := json.Unmarshal(data, &typeCheck); err != nil {
		return err
	}

	var ensembler models.EnsemblerLike
	switch typeCheck.Type {
	case models.EnsemblerPyFuncType:
		ensembler = &models.PyFuncEnsembler{}
	default:
		return fmt.Errorf("unsupported ensembler type: %s", typeCheck.Type)
	}

	if err := json.Unmarshal(data, ensembler); err != nil {
		return err
	}

	r.EnsemblerLike = ensembler
	return nil
}
