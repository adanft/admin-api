package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"admin.com/admin-api/internal/domain"
	appLogger "admin.com/admin-api/pkg/logger"
)

func WriteSuccess(w http.ResponseWriter, status int, body any) {
	writeJSON(w, status, SuccessResponse{
		Success: true,
		Data:    body,
		Status:  status,
	})
}

func WriteError(w http.ResponseWriter, status int, errMsg string) {
	WriteErrorWithCode(w, status, "", errMsg)
}

func WriteErrorWithCode(w http.ResponseWriter, status int, code string, errMsg string) {
	writeJSON(w, status, ErrorResponse{
		Success: false,
		Code:    code,
		Error:   errMsg,
		Status:  status,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error(appLogger.MsgResponseMarshalError, "status", status, "error", err)
		writeInternalFallback(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		slog.Error(appLogger.MsgResponseWriteError, "status", status, "error", err)
		return
	}

	if _, err := w.Write([]byte("\n")); err != nil {
		slog.Error(appLogger.MsgResponseWriteError, "status", status, "error", err)
	}
}

func writeInternalFallback(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	fallback := `{"success":false,"error":"` + domain.InternalServerErrorMessage + `","status":` + strconv.Itoa(http.StatusInternalServerError) + `}` + "\n"
	if _, err := w.Write([]byte(fallback)); err != nil {
		slog.Error(appLogger.MsgResponseFallbackWriteErr, "status", http.StatusInternalServerError, "error", err)
	}
}
