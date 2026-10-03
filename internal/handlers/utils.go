package handlers

import (
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
)

func DoSendContentOnly(r *http.Request) bool {
	isHtmxRequest := r.Header.Get(constants.HxRequestHeaderRequest) == "true"
	isHistoryRestore := r.Header.Get(constants.HxRequestHeaderHistoryRestore) == "true"
	return isHtmxRequest && !isHistoryRestore
}
