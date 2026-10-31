package templates

type ToastData struct {
	ToastSuccess string
}

type ErrorTemplateData struct {
	Code       int
	Error      string
	SubErrors  []string
	InnerError error
}

type PageConfig struct {
	ContentOnly  bool
	Title        string
	ToastSuccess string
}
