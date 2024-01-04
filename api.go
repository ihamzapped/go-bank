package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)

const jwtSecret = "for_demo_purposes"

type APIServer struct {
	listenAddr string
	store      Storage
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

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

	tokenStr, err := createJWT(acc)

	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, &LoginResponse{Account: acc, TokenStr: tokenStr})
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

func (s *APIServer) handleTransferAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func UseJWT(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenStr)

		if err != nil || !token.Valid {
			// log.Println(err)
			writeJSON(w, http.StatusForbidden, ApiError{Error: "Invalid Token"})
			return
		}

		f(w, r)
	}
}

func createJWT(acc *Account) (string, error) {
	claims := &jwt.MapClaims{
		"id":            acc.ID,
		"accountNumber": acc.AccNumber,
		"iat":           time.Now().Unix(),
		"exp":           time.Now().Add(time.Hour * 24).Unix(), // 24 hour expiration
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func validateJWT(tokenStr string) (*jwt.Token, error) {
	hmacSampleSecret := []byte(jwtSecret)
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSampleSecret, nil
	})
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
