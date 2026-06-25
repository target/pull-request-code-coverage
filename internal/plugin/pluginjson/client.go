package pluginjson

import "encoding/json"

type Client interface {
	Marshal(data interface{}) ([]byte, error)
}

type DefaultClient struct{}

func (c *DefaultClient) Marshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}
