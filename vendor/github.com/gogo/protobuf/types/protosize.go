// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

func (m *Any) ProtoSize() (n int)               { return m.Size() }
func (m *Api) ProtoSize() (n int)               { return m.Size() }
func (m *Method) ProtoSize() (n int)            { return m.Size() }
func (m *Mixin) ProtoSize() (n int)             { return m.Size() }
func (m *Duration) ProtoSize() (n int)          { return m.Size() }
func (m *Empty) ProtoSize() (n int)             { return m.Size() }
func (m *FieldMask) ProtoSize() (n int)         { return m.Size() }
func (m *SourceContext) ProtoSize() (n int)     { return m.Size() }
func (m *Struct) ProtoSize() (n int)            { return m.Size() }
func (m *Value) ProtoSize() (n int)             { return m.Size() }
func (m *Value_NullValue) ProtoSize() (n int)   { return m.Size() }
func (m *Value_NumberValue) ProtoSize() (n int) { return m.Size() }
func (m *Value_StringValue) ProtoSize() (n int) { return m.Size() }
func (m *Value_BoolValue) ProtoSize() (n int)   { return m.Size() }
func (m *Value_StructValue) ProtoSize() (n int) { return m.Size() }
func (m *Value_ListValue) ProtoSize() (n int)   { return m.Size() }
func (m *ListValue) ProtoSize() (n int)         { return m.Size() }
func (m *Timestamp) ProtoSize() (n int)         { return m.Size() }
func (m *Type) ProtoSize() (n int)              { return m.Size() }
func (m *Field) ProtoSize() (n int)             { return m.Size() }
func (m *Enum) ProtoSize() (n int)              { return m.Size() }
func (m *EnumValue) ProtoSize() (n int)         { return m.Size() }
func (m *Option) ProtoSize() (n int)            { return m.Size() }
func (m *DoubleValue) ProtoSize() (n int)       { return m.Size() }
func (m *FloatValue) ProtoSize() (n int)        { return m.Size() }
func (m *Int64Value) ProtoSize() (n int)        { return m.Size() }
func (m *UInt64Value) ProtoSize() (n int)       { return m.Size() }
func (m *Int32Value) ProtoSize() (n int)        { return m.Size() }
func (m *UInt32Value) ProtoSize() (n int)       { return m.Size() }
func (m *BoolValue) ProtoSize() (n int)         { return m.Size() }
func (m *StringValue) ProtoSize() (n int)       { return m.Size() }
func (m *BytesValue) ProtoSize() (n int)        { return m.Size() }
