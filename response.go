package compass

type Response struct {
	ContentType *string
	Body        []byte
	StatusCode  int
}

func Raw(contentType *string, body []byte, code int) Response {
	return Response{
		ContentType: contentType,
		Body:        body,
		StatusCode:  code,
	}
}

func Text(text string) Response {
	return TextWithCode(text, 200)
}

func TextWithCode(text string, code int) Response {
	return Raw(nil, []byte(text), code)
}
