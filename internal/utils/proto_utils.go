package utils

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

// MarshalProtoMessage serializes a protobuf message into a byte slice.
func MarshalProtoMessage(msg interface{}) ([]byte, error) {
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		log.Error().Msg("invalid request type")

		return nil, errors.New("request does not implement proto.Message")
	}

	data, err := proto.Marshal(protoMsg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal proto message")
	}

	return data, nil
}

// UnmarshalProtoMessage deserializes a byte slice into the given protobuf message.
func UnmarshalProtoMessage(data []byte, msg proto.Message) error {
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal proto message")
	}

	return nil
}
