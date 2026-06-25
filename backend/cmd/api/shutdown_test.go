package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestGracefulShutdown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(200 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	srv := &http.Server{Handler: r}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()

	go srv.Serve(ln) //nolint:errcheck

	// Fire a slow request
	done := make(chan int, 1)
	go func() {
		resp, err := http.Get("http://" + addr + "/slow") //nolint:noctx
		if err != nil {
			done <- 0
			return
		}
		done <- resp.StatusCode
	}()

	time.Sleep(50 * time.Millisecond) // let request land

	// Shutdown — should wait for the slow request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	select {
	case code := <-done:
		if code != http.StatusOK {
			t.Errorf("in-flight request got status %d, want 200", code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("in-flight request never completed")
	}
}
