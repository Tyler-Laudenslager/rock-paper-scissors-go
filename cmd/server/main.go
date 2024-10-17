// cmd/server/main.go
// **************************************************************
// Author: Tyler Laudenslager
// Purpose: Server implementation for Rock, Paper, Scissors
//          Handles multiple concurrent games with multiple players.
// **************************************************************

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"rps_game/internal/rps"
)

// NewGameManager creates a new GameManager with specified rounds.
func NewGameManager(rounds int) *rps.GameManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &rps.GameManager{
		WaitingPlayers: make(chan net.Conn, 100), // buffer for 100 waiting players
		Rounds:         rounds,
		wg:             sync.WaitGroup{},
		ctx:            ctx,
		cancel:         cancel,
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <number_of_rounds> <port>\n", os.Args[0])
		os.Exit(1)
	}

	rounds, err := strconv.Atoi(os.Args[1])
	if err != nil || rounds <= 0 {
		log.Fatalf("Invalid number_of_rounds: %v\n", err)
	}
	port := os.Args[2]

	// Initialize the server
	server := rps.NewServer(rounds, port)

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Server encountered an error: %v\n", err)
	}
}
