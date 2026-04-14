package core

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// ListenForShutdown blocks until SIGINT or SIGTERM, then gracefully
// drains active connections within ShutdownTimeout. This is the
// recommended way to handle server lifecycle in production.
func (e *Engine) ListenForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("rudra: shutdown signal received, draining connections...")

	ctx, cancel := context.WithTimeout(context.Background(), e.opts.ShutdownTimeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Printf("rudra: forced shutdown after timeout: %v", err)
	} else {
		log.Println("rudra: clean shutdown complete")
	}
}
