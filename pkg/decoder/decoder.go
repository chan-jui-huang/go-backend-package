package decoder

import (
	"time"

	"github.com/mitchellh/mapstructure"
)

func Decode(input any, output any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339),
		),
		Result: output,
	})
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
