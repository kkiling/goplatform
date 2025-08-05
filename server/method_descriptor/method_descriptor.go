package methoddescriptor

import (
	"errors"
	"reflect"
	"runtime"
	"strings"
)

var (
	ErrMethodDescriptorNotFound = errors.New("method method_descriptor not found")
)

type Descriptor interface {
	Method() interface{}
}

func getName(m Descriptor) (string, error) {
	methodPointer := reflect.ValueOf(m.Method()).Pointer()
	fullName := runtime.FuncForPC(methodPointer).Name()
	methodNameParts := strings.Split(fullName, ".")
	return methodNameParts[len(methodNameParts)-1], nil
}

type DescriptorsMap map[string]Descriptor

func (m DescriptorsMap) GetByFullName(fullName string) (Descriptor, bool) {
	methodNameParts := strings.Split(fullName, "/")
	methodName := methodNameParts[len(methodNameParts)-1]
	methodDescriptor, ok := m[methodName]
	return methodDescriptor, ok
}

func NewMethodDescriptorMap(
	methodDescriptors []Descriptor,
) (DescriptorsMap, error) {
	m := make(DescriptorsMap)
	for _, methodDescriptor := range methodDescriptors {
		methodName, err := getName(methodDescriptor)
		if err != nil {
			return nil, err
		}
		m[methodName] = methodDescriptor
	}
	return m, nil
}
