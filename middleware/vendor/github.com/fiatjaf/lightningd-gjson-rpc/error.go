package lightning

import "fmt"

type ErrorConnect struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type ErrorCommand struct {
	Message string      `json:"msg"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

type ErrorTimeout struct {
	Seconds int `json:"seconds"`
}

type ErrorJSONDecode struct {
	Message string `json:"message"`
}

type ErrorConnectionBroken struct{}

func (c ErrorConnect) Error() string {
	return fmt.Sprintf("unable to dial socket %s:%s", c.Path, c.Message)
}
func (l ErrorCommand) Error() string {
	return fmt.Sprintf("lightningd replied with error: %s (%d)", l.Message, l.Code)
}
func (t ErrorTimeout) Error() string {
	return fmt.Sprintf("call timed out after %ds", t.Seconds)
}
func (j ErrorJSONDecode) Error() string {
	return "error decoding JSON response from lightningd: " + j.Message
}
func (c ErrorConnectionBroken) Error() string {
	return "got an EOF while reading response, it seems the connection is broken"
}
