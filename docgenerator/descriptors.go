package docgenerator

import (
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/topfreegames/pitaya/v2/constants"
)

// ProtoDescriptors returns the descriptor for a given message name or .proto file
// 返回给定消息名称或.proto文件的描述符
func ProtoDescriptors(protoName string) ([]byte, error) {
	if strings.HasSuffix(protoName, ".proto") {
		descriptor := proto.FileDescriptor(protoName)
		if descriptor == nil {
			return nil, constants.ErrProtodescriptor
		}
		return descriptor, nil
	}

	if strings.HasPrefix(protoName, "types.") {
		protoName = strings.Replace(protoName, "types.", "google.protobuf.", 1)
	}
	protoReflectTypePointer := proto.MessageType(protoName)
	if protoReflectTypePointer == nil {
		return nil, constants.ErrProtodescriptor
	}

	protoReflectType := protoReflectTypePointer.Elem()
	protoValue := reflect.New(protoReflectType)
	descriptorMethod, ok := protoReflectTypePointer.MethodByName("Descriptor")
	if !ok {
		return nil, constants.ErrProtodescriptor
	}

	descriptorValue := descriptorMethod.Func.Call([]reflect.Value{protoValue})
	protoDescriptor := descriptorValue[0].Bytes()

	return protoDescriptor, nil
}
