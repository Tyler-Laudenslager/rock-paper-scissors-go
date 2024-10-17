// cmd/client/main.go
// **************************************************************
// Author: Tyler Laudenslager
// Purpose: Client implementation for Rock, Paper, Scissors game
//          Connects to the server, sends nickname, makes choices,
//          and displays game outcomes.
// **************************************************************

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	"rps_game/internal/rps"
)

// Client represents the RPS client.
type Client struct {
	Conn     net.Conn
	Hostname string
	Port     string
}

// NewClient initializes a new RPS client.
func NewClient(hostname, port string) *Client {
	return &Client{
		Hostname: hostname,
		Port:     port,
	}
}

// Start connects to the server and starts the game.
func (c *Client) Start() error {
	address := net.JoinHostPort(c.Hostname, c.Port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to server at %s: %w", address, err)
	}
	defer conn.Close()
	c.Conn = conn

	fmt.Println("Connected to Rock, Paper, Scissors server.")

	// Handle graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	go func() {
		<-done
		fmt.Println("\nDisconnecting from server...")
		conn.Close()
		os.Exit(0)
	}()

	reader := bufio.NewReader(os.Stdin)

	// Receive initial READY message from server
	initialMsg, err := rps.ReceiveMessage(c.Conn)
	if err != nil {
		return fmt.Errorf("failed to receive initial message: %w", err)
	}
	if initialMsg != "READY" {
		return fmt.Errorf("unexpected initial message: %s", initialMsg)
	}

	// Prompt for nickname
	nickname, err := c.promptNickname(reader)
	if err != nil {
		return err
	}

	// Start the game loop
	if err := c.gameLoop(reader); err != nil {
		return err
	}

	fmt.Println("Game over. Thanks for playing!")
	return nil
}

// promptNickname prompts the user to enter a unique nickname and sends it to the server.
func (c *Client) promptNickname(reader *bufio.Reader) (string, error) {
	for {
		fmt.Print("Enter a unique nickname: ")
		nickname, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read nickname: %w", err)
		}
		nickname = strings.TrimSpace(nickname)
		if nickname == "" {
			fmt.Println("Nickname cannot be empty. Please try again.")
			continue
		}

		// Send nickname to server
		if err := rps.SendMessage(c.Conn, nickname); err != nil {
			return "", fmt.Errorf("failed to send nickname: %w", err)
		}

		// Wait for server response
		response, err := rps.ReceiveMessage(c.Conn)
		if err != nil {
			return "", fmt.Errorf("failed to receive nickname response: %w", err)
		}

		if response == "RETRY" {
			fmt.Println("Nickname is not unique. Please choose another one.")
			continue
		} else if response == "START" {
			fmt.Println("Nickname accepted. Game is starting...")
			return nickname, nil
		} else {
			fmt.Printf("Unexpected server response: %s\n", response)
			continue
		}
	}
}

// gameLoop handles the main game interaction with the server.
func (c *Client) gameLoop(reader *bufio.Reader) error {
	for {
		// Prompt for choice
		choice, err := c.promptChoice(reader)
		if err != nil {
			return err
		}

		// Send choice to server
		if err := rps.SendMessage(c.Conn, choice); err != nil {
			return fmt.Errorf("failed to send choice: %w", err)
		}

		// Receive response from server
		response, err := rps.ReceiveMessage(c.Conn)
		if err != nil {
			return fmt.Errorf("failed to receive game response: %w", err)
		}

		if strings.HasPrefix(response, "SCORE") {
			// Final score received
			fmt.Println("Final Score:", response)
			break
		}

		// Parse round result
		parts := strings.Split(response, " ")
		if len(parts) != 4 {
			fmt.Printf("Malformed round result: %s\n", response)
			continue
		}

		enemyChoice, outcome, roundsLeft, enemyNickname := parts[0], parts[1], parts[2], parts[3]
		fmt.Printf("\n%s's choice: %s\nOutcome: %s\nRounds left: %s\n\n", enemyNickname, enemyChoice, outcome, roundsLeft)

		if roundsLeft == "0" {
			break
		}
	}
	return nil
}

// promptChoice prompts the user to choose rock, paper, or scissors.
func (c *Client) promptChoice(reader *bufio.Reader) (string, error) {
	for {
		fmt.Print("Choose ('rock', 'paper', 'scissors'): ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read choice: %w", err)
		}
		choice = strings.TrimSpace(strings.ToLower(choice))
		if choice != "rock" && choice != "paper" && choice != "scissors" {
			fmt.Println("Invalid choice. Please enter 'rock', 'paper', or 'scissors'.")
			continue
		}
		return choice, nil
	}
}
