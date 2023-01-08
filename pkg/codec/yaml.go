package codec

import "gopkg.in/yaml.v3"

type yamlCodec struct{}

func (y yamlCodec) Marshal(in any) ([]byte, error) {
	return yaml.Marshal(in)
}

func (y yamlCodec) Unmarshal(in []byte, out any) error {
	return yaml.Unmarshal(in, out)
}

func (y yamlCodec) Extension() string {
	return "yml"
}

func NewYAMLCodec() Codec {
	return &yamlCodec{}
}
