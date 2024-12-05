package validatorutils

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

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
