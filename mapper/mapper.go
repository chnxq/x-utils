package mapper

import (
	"reflect"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type CopierMapper[DTO any, ENTITY any] struct {
	copierOption copier.Option
}

func NewCopierMapper[DTO any, ENTITY any]() *CopierMapper[DTO, ENTITY] {
	return &CopierMapper[DTO, ENTITY]{
		copierOption: copier.Option{
			Converters: []copier.TypeConverter{},
		},
	}
}

func (m *CopierMapper[DTO, ENTITY]) AppendConverter(converter copier.TypeConverter) {
	m.copierOption.Converters = append(m.copierOption.Converters, converter)
}

func (m *CopierMapper[DTO, ENTITY]) AppendConverters(converters []copier.TypeConverter) {
	m.copierOption.Converters = append(m.copierOption.Converters, converters...)
}

func (m *CopierMapper[DTO, ENTITY]) ToEntity(dto *DTO) *ENTITY {
	if dto == nil {
		return nil
	}

	var entity ENTITY
	if err := copier.CopyWithOption(&entity, dto, m.copierOption); err != nil {
		panic(err) // Handle error appropriately in production code
	}
	applyDefaultFieldConversions(&entity, dto)

	return &entity
}

func (m *CopierMapper[DTO, ENTITY]) ToDTO(entity *ENTITY) *DTO {
	if entity == nil {
		return nil
	}

	var dto DTO
	if err := copier.CopyWithOption(&dto, entity, m.copierOption); err != nil {
		panic(err) // Handle error appropriately in production code
	}
	applyDefaultFieldConversions(&dto, entity)

	return &dto
}

type reflectedProtoEnum interface {
	Descriptor() protoreflect.EnumDescriptor
	Number() protoreflect.EnumNumber
}

func applyDefaultFieldConversions(dst any, src any) {
	dstValue := indirectValue(reflect.ValueOf(dst))
	srcValue := indirectValue(reflect.ValueOf(src))
	if !dstValue.IsValid() || !srcValue.IsValid() {
		return
	}
	if dstValue.Kind() != reflect.Struct || srcValue.Kind() != reflect.Struct {
		return
	}

	dstType := dstValue.Type()
	for i := 0; i < dstValue.NumField(); i++ {
		dstField := dstValue.Field(i)
		if !dstField.CanSet() {
			continue
		}

		srcField := srcValue.FieldByName(dstType.Field(i).Name)
		if !srcField.IsValid() {
			continue
		}

		if setProtoEnumFieldFromString(dstField, srcField) {
			continue
		}
		setStringFieldFromProtoEnum(dstField, srcField)
	}
}

func setProtoEnumFieldFromString(dstField reflect.Value, srcField reflect.Value) bool {
	dstType, dstIsPtr := indirectType(dstField.Type())
	enumDescriptor, ok := protoEnumDescriptor(dstType)
	if !ok {
		return false
	}

	srcValue := indirectValue(srcField)
	if !srcValue.IsValid() || srcValue.Kind() != reflect.String {
		return false
	}

	enumValue := enumDescriptor.Values().ByName(protoreflect.Name(srcValue.String()))
	if enumValue == nil {
		return false
	}

	value, ok := newIntegerValue(dstType, int64(enumValue.Number()))
	if !ok {
		return false
	}
	setFieldValue(dstField, value, dstIsPtr)
	return true
}

func setStringFieldFromProtoEnum(dstField reflect.Value, srcField reflect.Value) bool {
	dstType, dstIsPtr := indirectType(dstField.Type())
	if dstType.Kind() != reflect.String {
		return false
	}

	srcType, _ := indirectType(srcField.Type())
	enumDescriptor, ok := protoEnumDescriptor(srcType)
	if !ok {
		return false
	}

	srcValue := indirectValue(srcField)
	if !srcValue.IsValid() || !isIntegerKind(srcValue.Kind()) {
		return false
	}

	enumValue := enumDescriptor.Values().ByNumber(protoreflect.EnumNumber(srcValue.Int()))
	if enumValue == nil {
		return false
	}

	value := reflect.New(dstType).Elem()
	value.SetString(string(enumValue.Name()))
	setFieldValue(dstField, value, dstIsPtr)
	return true
}

func protoEnumDescriptor(valueType reflect.Type) (protoreflect.EnumDescriptor, bool) {
	if valueType == nil || !isIntegerKind(valueType.Kind()) {
		return nil, false
	}

	zero := reflect.Zero(valueType)
	if !zero.CanInterface() {
		return nil, false
	}

	enumValue, ok := zero.Interface().(reflectedProtoEnum)
	if !ok {
		return nil, false
	}
	return enumValue.Descriptor(), true
}

func indirectType(valueType reflect.Type) (reflect.Type, bool) {
	if valueType.Kind() == reflect.Pointer {
		return valueType.Elem(), true
	}
	return valueType, false
}

func indirectValue(value reflect.Value) reflect.Value {
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func newIntegerValue(valueType reflect.Type, value int64) (reflect.Value, bool) {
	out := reflect.New(valueType).Elem()
	switch valueType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		out.SetUint(uint64(value))
	default:
		return reflect.Value{}, false
	}
	return out, true
}

func setFieldValue(field reflect.Value, value reflect.Value, fieldIsPtr bool) {
	if fieldIsPtr {
		ptr := reflect.New(value.Type())
		ptr.Elem().Set(value)
		field.Set(ptr)
		return
	}
	field.Set(value)
}

func isIntegerKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	default:
		return false
	}
}
