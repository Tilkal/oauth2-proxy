package app

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/logger"
)

var errorMessages = map[int]string{
	http.StatusInternalServerError: "Oops! Something went wrong. For more information contact your server administrator.",
	http.StatusNotFound:            "We could not find the resource you were looking for.",
	http.StatusForbidden:           "You do not have permission to access this resource.",
	http.StatusUnauthorized:        "You need to be logged in to access this resource.",
}

// ErrorPage is used to render error pages
type ErrorPage struct {
	// Template is the error page HTML template.
	Template *template.Template

	// ProxyPrefix is the prefix under which OAuth2 Proxy pages are served.
	ProxyPrefix string

	// Footer is the footer to be displayed at the bottom of the page.
	// If not set, a default footer will be used.
	Footer string

	// Version is the OAuth2 Proxy version to be used in the default footer.
	Version string

	// Debug determines whether errors pages should be rendered with detailed
	// errors.
	Debug bool
}

func (e *ErrorPage) Render(rw http.ResponseWriter, status int, redirectURL string, appError string, messages ...interface{}) {
	rw.WriteHeader(status)

	// We allow unescaped template.HTML since it is user configured options
	/* #nosec G203 */
	data := struct {
		Title       string
		Message     string
		ProxyPrefix string
		StatusCode  int
		Redirect    string
		Footer      template.HTML
		Version     string
	}{
		Title:       http.StatusText(status),
		Message:     e.getMessage(status, appError, messages...),
		ProxyPrefix: e.ProxyPrefix,
		StatusCode:  status,
		Redirect:    redirectURL,
		Footer:      template.HTML(e.Footer),
		Version:     e.Version,
	}

	if err := e.Template.Execute(rw, data); err != nil {
		logger.Printf("Error rendering error template: %v", err)
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (e *ErrorPage) ProxyErrorHandler(rw http.ResponseWriter, req *http.Request, proxyErr error) {
	logger.Errorf("Error proxying to upstream server: %v", proxyErr)
	e.Render(rw, http.StatusBadGateway, "", proxyErr.Error(), "There was a problem connecting to the upstream server.")
}

func (e *ErrorPage) getMessage(status int, appError string, messages ...interface{}) string {
	if e.Debug {
		return appError
	}
	if len(messages) > 0 {
		format := fmt.Sprintf("%v", messages[0])
		return fmt.Sprintf(format, messages[1:]...)
	}
	if msg, ok := errorMessages[status]; ok {
		return msg
	}
	return "Unknown error"
}
