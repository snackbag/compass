package compass

import (
	"encoding/json"
	"fmt"
)

type Response struct {
	internalError bool

	ContentType *string
	Body        []byte
	StatusCode  int
	Headers     map[string]string
}

func Raw(contentType *string, body []byte, code int) Response {
	return Response{
		internalError: false,

		ContentType: contentType,
		Body:        body,
		StatusCode:  code,
		Headers:     make(map[string]string),
	}
}

func InternalError(message string, code int) Response {
	return Response{
		internalError: true,

		ContentType: nil,
		Body:        []byte(message),
		StatusCode:  code,
		Headers:     make(map[string]string),
	}
}

func Text(text string) Response {
	return TextWithCode(text, 200)
}

func TextWithCode(text string, code int) Response {
	return Raw(nil, []byte(text), code)
}

func JsonString(content string) Response {
	return JsonStringWithCode(content, 200)
}

func JsonStringWithCode(content string, code int) Response {
	typ := "application/json"
	return Raw(&typ, []byte(content), code)
}

func JsonMarshal(obj any) Response {
	return JsonMarshalWithCode(obj, 200)
}

func JsonMarshalWithCode(obj any, code int) Response {
	content, err := json.Marshal(obj)
	if err != nil {
		return InternalError(fmt.Sprintf("failed to marshal json object: %s", err), 500)
	}

	return JsonStringWithCode(string(content), code)
}
