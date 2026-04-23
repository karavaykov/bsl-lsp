package lsp

func Run() error {
	logf := NewLogFunc()
	handler := NewHandler(logf)
	serve(logf, handler)
	return nil
}
