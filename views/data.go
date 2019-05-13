package views

import (
	"net/http"
	"time"
)

func persistAlert(w http.ResponseWriter, alert Alert) {
	expiresAt := time.Now().Add(5 * time.Minute)
	lvl := http.Cookie{
		Name:     "alert_level",
		Value:    alert.Level,
		Expires:  expiresAt,
		HttpOnly: true,
	}

	msg := http.Cookie{
		Name:     "alert_msg",
		Value:    alert.Message,
		Expires:  expiresAt,
		HttpOnly: true,
	}

	http.SetCookie(w, &lvl)
	http.SetCookie(w, &msg)
}

func clearAlert(w http.ResponseWriter) {
	lvl := http.Cookie{
		Name:     "alert_level",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}

	msg := http.Cookie{
		Name:     "alert_msg",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}

	http.SetCookie(w, &lvl)
	http.SetCookie(w, &msg)
}

func getAlert(r *http.Request) *Alert {
	lvl, err := r.Cookie("alert_level")
	if err != nil {
		return nil
	}
	msg, err := r.Cookie("alert_msg")
	if err != nil {
		return nil
	}

	alert := &Alert{
		Level:   lvl.String(),
		Message: msg.String(),
	}

	return alert
}

func RedirectAlert(w http.ResponseWriter, r *http.Request, dstUrl string, code int, alert Alert) {
	persistAlert(w, alert)
	http.Redirect(w, r, dstUrl, code)
}
