package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

var (
	vaultAddr      string
	roleID         string
	secretID       string
	vaultPath      string
	cronExpr       string
	certPath       string
	privateKeyPath string
	token          string
	tokenMu        sync.Mutex
)

type AppRoleLoginRequest struct {
	RoleID   string `json:"role_id"`
	SecretID string `json:"secret_id"`
}

type AppRoleLoginResponse struct {
	Auth struct {
		ClientToken string `json:"client_token"`
	} `json:"auth"`
}

type VaultResponse struct {
	Data struct {
		Data map[string]string `json:"data"`
	} `json:"data"`
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, loading from OS environment")
	}
	vaultAddr = os.Getenv("VAULT_ADDR")
	roleID = os.Getenv("VAULT_ROLE_ID")
	secretID = os.Getenv("VAULT_SECRET_ID")
	vaultPath = os.Getenv("VAULT_PATH")
	cronExpr = os.Getenv("CRON_SCHEDULE")
	certPath = os.Getenv("CERT_PATH")
	privateKeyPath = os.Getenv("PRIVATE_KEY_PATH")

	if vaultAddr == "" || roleID == "" || secretID == "" || vaultPath == "" || cronExpr == "" || certPath == "" {
		log.Fatal("Missing required environment variables")
	}
}

func loginToVault() error {
	url := fmt.Sprintf("%s/v1/auth/approle/login", vaultAddr)
	body, _ := json.Marshal(AppRoleLoginRequest{RoleID: roleID, SecretID: secretID})
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to authenticate with Vault, status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading login response: %v", err)
	}

	var loginResp AppRoleLoginResponse
	if err := json.Unmarshal(bodyBytes, &loginResp); err != nil {
		return fmt.Errorf("failed to parse Vault login response: %v", err)
	}

	tokenMu.Lock()
	token = loginResp.Auth.ClientToken
	tokenMu.Unlock()
	return nil
}

func fetchSSLCerts() error {
	tokenMu.Lock()
	currentToken := token
	tokenMu.Unlock()

	if currentToken == "" {
		if err := loginToVault(); err != nil {
			return fmt.Errorf("failed to login to Vault: %v", err)
		}
		tokenMu.Lock()
		currentToken = token // Update after login
		tokenMu.Unlock()
	}

	fmt.Print("current token:" + currentToken)

	url := fmt.Sprintf("%s/v1/%s", vaultAddr, vaultPath)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Vault-Token", currentToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch SSL certs, status: %d", resp.StatusCode)
	}

	var vaultResp VaultResponse
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return err
	}

	certPEM, ok1 := vaultResp.Data.Data["tls.crt"]
	keyPEM, ok2 := vaultResp.Data.Data["tls.key"]
	if !ok1 || !ok2 {
		return fmt.Errorf("missing certificate or private key in Vault response")
	}

	certPath := filepath.Join(certPath)
	keyPath := filepath.Join(privateKeyPath)

	if err := os.WriteFile(certPath, []byte(certPEM), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(keyPath, []byte(keyPEM), 0600); err != nil {
		return err
	}

	log.Println("SSL certificate and key updated successfully in", certPath+" and "+privateKeyPath)
	return nil
}

func main() {
	loadEnv()
	log.Println("Starting SSL cert fetcher...")

	if err := fetchSSLCerts(); err != nil {
		log.Println("Error fetching SSL certs on startup:", err)
	}

	c := cron.New()
	c.AddFunc(cronExpr, func() {
		if err := fetchSSLCerts(); err != nil {
			log.Println("Error fetching SSL certs:", err)
		}
	})
	c.Start()
	select {} // Keeps the main routine alive
}
