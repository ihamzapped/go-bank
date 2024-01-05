package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strconv"
	"time"
)

const jwtSecret = "for_demo_purposes"

type UsrClaims struct {
	ID        int    `json:"id"`
	AccNumber uint64 `json:"accNumber"`
	jwt.RegisteredClaims
}

type APIServer struct {
	listenAddr string
	store      Storage
}

type apiFunc func(r http.ResponseWriter, w *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func NewApiServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/login", makeHttpHandleFunc(s.handleLogin))
	router.HandleFunc("/register", makeHttpHandleFunc(s.handleCreateAccount))
	router.HandleFunc("/transfer", UseJWT(makeHttpHandleFunc(s.handleTransfer)))
	router.HandleFunc("/account/{id}", UseJWT(makeHttpHandleFunc(s.handleAccount)))

	log.Printf("Api server running on http://localhost%s", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)

}

func (s *APIServer) notAllowed(r *http.Request) error {
	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccountByID(w, r)
	case "DELETE":
		return s.handleDeleteAccount(w, r)
	default:
		return s.notAllowed(r)
	}
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)

	if err != nil {
		return fmt.Errorf("Invalid id given %s", idStr)
	}

	acc, err := s.store.GetAccountByID(id)

	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, acc)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return s.notAllowed(r)
	}

	req := &LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	acc, err := s.store.GetAccountByNumber(req.AccNumber)

	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(req.Password))

	if err != nil {
		return err
	}

	tokenStr, err := createJWT(acc)

	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"account": acc,
		"token":   tokenStr,
	})
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return s.notAllowed(r)
	}

	createAccountReq := &CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}

	account, err := NewAccount(createAccountReq.FirstName, createAccountReq.LastName)

	if err != nil {
		return err
	}

	res, err := s.store.CreateAccount(account)

	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, res)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)

	if err != nil {
		return fmt.Errorf("Invalid id given %s", idStr)
	}

	return s.store.DeleteAccount(id)
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return s.notAllowed(r)
	}

	claims, ok := r.Context().Value("user").(*UsrClaims)

	if !ok {
		log.Print("Smth went wrong while getting claims from context")
	}

	req := &TransferRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	sender, err := s.store.GetAccountByNumber(claims.AccNumber)

	if err != nil {
		return err
	}

	recipient, err := s.store.GetAccountByNumber(req.Recipient)

	if err != nil {
		return err
	}

	if sender.Balance <= req.Amount {
		return writeJSON(w, http.StatusBadRequest, "Insufficient Balance")
	}

	newSenderBal := sender.Balance - req.Amount
	err = s.store.UpdateAccountBal(sender.ID, newSenderBal)

	if err != nil {
		return err
	}

	newRecipientBal := recipient.Balance + req.Amount
	err = s.store.UpdateAccountBal(recipient.ID, newRecipientBal)

	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, true)
}

func UseJWT(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("x-jwt-token")
		claims, err := validateJWT(tokenStr)

		if err != nil {
			log.Println(err)
			writeJSON(w, http.StatusForbidden, ApiError{Error: "Invalid Token"})
			return
		}

		ctx := context.WithValue(r.Context(), "user", claims)
		r = r.WithContext(ctx)

		f(w, r)
	}
}

func createJWT(acc *Account) (string, error) {
	claims := &UsrClaims{
		acc.ID,
		acc.AccNumber,
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hour expiration
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func validateJWT(tokenStr string) (*UsrClaims, error) {
	hmacSampleSecret := []byte(jwtSecret)
	token, err := jwt.ParseWithClaims(tokenStr, &UsrClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSampleSecret, nil
	})

	if err != nil {
		return &UsrClaims{}, err
	}

	claims, ok := token.Claims.(*UsrClaims)

	if ok && token.Valid {
		return claims, nil
	}

	return &UsrClaims{}, errors.New("Failed to parse claims")

}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}
