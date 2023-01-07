package codec

type Marshaler interface {
	Marshal(in any) ([]byte, error)
}

type Unmarshaler interface {
	Unmarshal(in []byte, out any) error
}

type Codec interface {
	Marshaler
	Unmarshaler
	Extension() string
}
