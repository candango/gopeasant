package peasant

import "net/http"

// NonceService defines methods for managing nonces in HTTP requests.
// It provides functionality for blocking, clearing, consuming, getting,
// and checking the provision of nonces.
type NonceService interface {

	// Block blocks the provided HTTP request if the nonce is not valid.
	Block(http.ResponseWriter, *http.Request) error

	// Clear clears a nonce associated with the specified key. If the key
	// doesn't exists no error will be returned.
	//
	// Return errors only if an actual error occours.
	Clear(string) error

	// Consume processes the nonce associated with the specified key and
	// returns a boolean indicating whether the nonce was successfully
	// consumed, along with any error encountered.
	// If the specified key is not present in the map, the method will return
	// false without performing any action.
	Consume(http.ResponseWriter, *http.Request) (bool, error)

	// GetNonce generates a new nonce, and stores it for a future validation.
	// It returns the nonce as a string and an error if any occurred during
	// the nonce generation or header update.
	GetNonce(*http.Request) (string, error)

	// Provided verifies the presence of a valid nonce in the specified HTTP
	// request.
	//
	// If the nonce is not provided or is invalid, it sets the response HTTP
	// status to "Unauthorized", "Forbidden", or another appropriate status
	// based on the specific conditions and checks performed within the method.
	// It returns a boolean indicating whether the nonce was provided and
	// valid, along with any error encountered.
	Provided(http.ResponseWriter, *http.Request) (bool, error)
}

func NoncedHandlerFunc(s NonceService,
	f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ok, err := s.Provided(res, req)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			return
		}
		ok, err = s.Consume(res, req)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			return
		}
		nonce, err := s.GetNonce(req)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.Header().Add("nonce", nonce)
		f(res, req)
	}
}

// Nonced processes the provided HTTP request to verify and consume a
// valid nonce. It returns an error if any occurred during the process.
func Nonced(res http.ResponseWriter, req *http.Request,
	service NonceService) (err error) {
	ok, err := service.Provided(res, req)
	if err != nil {
		return err
	}
	if !ok {
		err = service.Block(res, req)
		if err != nil {
			return err
		}
		return nil
	}
	nonce, err := service.GetNonce(req)
	if err != nil {
		return err
	}

	ok, err = service.Consume(res, req)
	if err != nil {
		return err
	}
	if ok {
		err = service.Clear(nonce)
		if err != nil {
			return err
		}
	}

	return nil
}
