package pluginjson

import "encoding/json"

type Client interface {
	Marshal(data interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type DefaultClient struct{}

func (c *DefaultClient) Marshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (c *DefaultClient) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
