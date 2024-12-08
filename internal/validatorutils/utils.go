package validatorutils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

type StringValue interface {
	ValueString() string
}

func ToAttrMap(values map[string]string, diagnostics *diag.Diagnostics) types.Map {
	attrMap := map[string]attr.Value{}
	for k, v := range values {
		attrMap[k] = types.StringValue(v)
	}

	convertedMap, mapDiag := types.MapValue(basetypes.StringType{}, attrMap)
	diagnostics.Append(mapDiag...)
	return convertedMap
}

func ToAttrList(values []string, diagnostics *diag.Diagnostics) types.List {
	attrList := make([]attr.Value, len(values))
	for i, v := range values {
		attrList[i] = types.StringValue(v)
	}

	convertedList, listDiag := types.ListValue(basetypes.StringType{}, attrList)
	diagnostics.Append(listDiag...)
	return convertedList
}

func IDsToAttrList(values []clientmodels.ID, diagnostics *diag.Diagnostics) types.List {
	attrList := make([]attr.Value, len(values))
	for i, v := range values {
		attrList[i] = types.StringValue(v.UUID.String())
	}

	convertedList, listDiag := types.ListValue(basetypes.StringType{}, attrList)
	diagnostics.Append(listDiag...)
	return convertedList
}

func AttrListToIDs(values types.List) ([]clientmodels.ID, error) {
	var ids []clientmodels.ID
	for _, id := range values.Elements() {
		strVal, ok := id.(StringValue)
		if !ok {
			return nil, fmt.Errorf("expected string for IDs, got %T", id)
		}

		uid, err := uuid.Parse(strVal.ValueString())
		if err != nil {
			return nil, err
		}

		ids = append(ids, clientmodels.ID{UUID: uid})
	}

	return ids, nil
}
