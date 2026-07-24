package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/LisLisich/RESTAPI/internal/cli"
)

const defaultAPIURL = "http://localhost:5050"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Getenv))
}

func run(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	getenv func(string) string,
) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "login":
		return runLogin(args[1:], stdout, stderr, getenv)
	case "refresh":
		return runRefresh(args[1:], stdout, stderr, getenv)
	case "wallet":
		return runWallet(args[1:], stdout, stderr, getenv)
	case "keygen":
		return runKeygen(stdout, stderr)
	default:
		fmt.Fprintf(stderr, "неизвестная команда %q\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runLogin(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	getenv func(string) string,
) int {
	flags := flag.NewFlagSet("login", flag.ContinueOnError)
	flags.SetOutput(stderr)
	apiURL := flags.String("api-url", apiURLFromEnvironment(getenv), "адрес FinTask API")
	email := flags.String("email", "", "email аккаунта")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	password := getenv("FINTASK_PASSWORD")
	if strings.TrimSpace(*email) == "" || password == "" {
		fmt.Fprintln(stderr, "нужны --email и переменная окружения FINTASK_PASSWORD")
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pair, err := cli.NewClient(*apiURL, http.DefaultClient).Login(ctx, *email, password)
	return writeResult(stdout, stderr, pair, err)
}

func runRefresh(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	getenv func(string) string,
) int {
	flags := flag.NewFlagSet("refresh", flag.ContinueOnError)
	flags.SetOutput(stderr)
	apiURL := flags.String("api-url", apiURLFromEnvironment(getenv), "адрес FinTask API")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	refreshToken := getenv("FINTASK_REFRESH_TOKEN")
	if refreshToken == "" {
		fmt.Fprintln(stderr, "нужна переменная окружения FINTASK_REFRESH_TOKEN")
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pair, err := cli.NewClient(*apiURL, http.DefaultClient).Refresh(ctx, refreshToken)
	return writeResult(stdout, stderr, pair, err)
}

func runWallet(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	getenv func(string) string,
) int {
	flags := flag.NewFlagSet("wallet", flag.ContinueOnError)
	flags.SetOutput(stderr)
	apiURL := flags.String("api-url", apiURLFromEnvironment(getenv), "адрес FinTask API")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	accessToken := getenv("FINTASK_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Fprintln(stderr, "нужна переменная окружения FINTASK_ACCESS_TOKEN")
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	wallet, err := cli.NewClient(*apiURL, http.DefaultClient).GetWallet(ctx, accessToken)
	return writeResult(stdout, stderr, wallet, err)
}

func runKeygen(stdout io.Writer, stderr io.Writer) int {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Fprintf(stderr, "не удалось создать ключ: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, base64.StdEncoding.EncodeToString(privateKey))
	return 0
}

func writeResult(stdout io.Writer, stderr io.Writer, value any, err error) int {
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintf(stderr, "не удалось вывести результат: %v\n", err)
		return 1
	}
	return 0
}

func apiURLFromEnvironment(getenv func(string) string) string {
	if value := strings.TrimSpace(getenv("FINTASK_API_URL")); value != "" {
		return value
	}
	return defaultAPIURL
}

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "использование: fintaskctl <login|refresh|wallet|keygen> [параметры]")
}
