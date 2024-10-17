// internal/rps/library.go
// **************************************************************
// Author: Tyler Laudenslager
// Purpose: Common functions for both the client and server
//          for Rock, Paper, Scissors
// **************************************************************

package rps

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// InvalidHeaderError is returned when the message does not start with the correct header.
type InvalidHeaderError struct {
	Received string
}

func (e *InvalidHeaderError) Error() string {
	return fmt.Sprintf("invalid header: expected ⚠, got %s", e.Received)
}

// InvalidFooterError is returned when the message does not end with the correct footer.
type InvalidFooterError struct {
	Received string
}

func (e *InvalidFooterError) Error() string {
	return fmt.Sprintf("invalid footer: expected ☠, got %s", e.Received)
}

// Encrypt shifts each character in the message by 3 positions.
// This simple encryption ensures basic message security.
func Encrypt(msg string) string {
	var encrypted strings.Builder
	for _, char := range msg {
		encrypted.WriteRune(char + 3)
	}
	return encrypted.String()
}

// Decrypt shifts each character in the message by -3 positions.
// It reverses the encryption applied to the message.
func Decrypt(msg string) string {
	var decrypted strings.Builder
	for _, char := range msg {
		decrypted.WriteRune(char - 3)
	}
	return decrypted.String()
}

// SendMessage sends a message with a header and footer over the connection.
// It encrypts the message before sending to ensure basic security.
func SendMessage(conn net.Conn, msg string) error {
	header := "\u26A0" // ⚠ Warning Sign
	footer := "\u2620" // ☠ Skull and Crossbones
	encryptedMsg := Encrypt(msg)
	fullMsg := fmt.Sprintf("%s%s%s\n", header, encryptedMsg, footer)

	writer := bufio.NewWriter(conn)
	_, err := writer.WriteString(fullMsg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush message: %w", err)
	}

	return nil
}

// ReceiveMessage reads, validates, and decrypts a message from the connection.
// It ensures that the message contains the correct header and footer.
func ReceiveMessage(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read message: %w", err)
	}

	msg = strings.TrimSpace(msg)

	if len(msg) < 2 {
		return "", errors.New("message too short to contain header and footer")
	}

	header := msg[:1]
	footer := msg[len(msg)-1:]
	content := msg[1 : len(msg)-1]

	if header != "\u26A0" { // ⚠
		return "", &InvalidHeaderError{Received: header}
	}

	if footer != "\u2620" { // ☠
		return "", &InvalidFooterError{Received: footer}
	}

	decryptedMsg := Decrypt(content)
	return decryptedMsg, nil
}

// Player represents a connected client.
type Player struct {
	Conn     net.Conn
	Nickname string
	Choice   string
	Score    int
}

// Game represents a game between two players.
type Game struct {
	Player1 *Player
	Player2 *Player
	Rounds  int
}

// NewServer initializes a new RPS server.
func NewServer(rounds int, port string) *Server {
	return &Server{
		Rounds:         rounds,
		Port:           port,
		WaitingPlayers: make(chan net.Conn, 100), // buffer for 100 waiting players
		Quit:           make(chan os.Signal, 1),
	}
}

// Server represents the RPS server.
type Server struct {
	Rounds         int
	Port           string
	WaitingPlayers chan net.Conn
	Quit           chan os.Signal
	wg             sync.WaitGroup
}

// determineOutcome calculates the outcome for each player based on their choices.
// Returns the score increments for Player1 and Player2 respectively.
func determineOutcome(p1Choice, p2Choice string) (int, int) {
	// Rules: rock beats scissors, scissors beats paper, paper beats rock
	if p1Choice == p2Choice {
		return 0, 0 // draw
	}

	var outcome1, outcome2 int

	switch p1Choice {
	case "rock":
		if p2Choice == "scissors" {
			outcome1 = 1 // Player1 wins
			outcome2 = -1
		} else { // paper
			outcome1 = -1 // Player1 loses
			outcome2 = 1
		}
	case "paper":
		if p2Choice == "rock" {
			outcome1 = 1
			outcome2 = -1
		} else { // scissors
			outcome1 = -1
			outcome2 = 1
		}
	case "scissors":
		if p2Choice == "paper" {
			outcome1 = 1
			outcome2 = -1
		} else { // rock
			outcome1 = -1
			outcome2 = 1
		}
	}

	return outcome1, outcome2
}

// outcomeString converts the outcome integer to a string representation.
func outcomeString(score int) string {
	switch score {
	case 1:
		return "win"
	case -1:
		return "lose"
	default:
		return "draw"
	}
}
