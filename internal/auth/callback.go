package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

// CallbackServer handles the OAuth callback on localhost.
type CallbackServer struct {
	listener net.Listener
	codeChan chan string
	errChan  chan error
}

// NewCallbackServer creates a callback server listening on a random available port.
func NewCallbackServer() (*CallbackServer, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}

	return &CallbackServer{
		listener: listener,
		codeChan: make(chan string, 1),
		errChan:  make(chan error, 1),
	}, nil
}

// Port returns the port the server is listening on.
func (s *CallbackServer) Port() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

// RedirectURL returns the redirect URL for OAuth.
func (s *CallbackServer) RedirectURL() string {
	return fmt.Sprintf("http://localhost:%d/callback", s.Port())
}

// Start begins serving and handling the OAuth callback.
func (s *CallbackServer) Start(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)

	server := &http.Server{Handler: mux}

	go func() {
		<-ctx.Done()
		server.Close()
	}()

	go func() {
		if err := server.Serve(s.listener); err != http.ErrServerClosed {
			s.errChan <- err
		}
	}()
}

func (s *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		errMsg := r.URL.Query().Get("error")
		if errMsg == "" {
			errMsg = "no authorization code received"
		}
		s.errChan <- fmt.Errorf("OAuth error: %s", errMsg)
		http.Error(w, "Authorization failed", http.StatusBadRequest)
		return
	}

	s.codeChan <- code

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>gmail-cli</title></head>
<body>
<h1>Authorization successful!</h1>
<p>You can close this window and return to the terminal.</p>
</body>
</html>`)
}

// WaitForCode blocks until an authorization code is received or an error occurs.
func (s *CallbackServer) WaitForCode(ctx context.Context) (string, error) {
	select {
	case code := <-s.codeChan:
		return code, nil
	case err := <-s.errChan:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Close shuts down the callback server.
func (s *CallbackServer) Close() error {
	return s.listener.Close()
}
