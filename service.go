package peasant

import "net/http"

type NonceService interface {
	Block(http.ResponseWriter, *http.Request) error
	Clear(string) error
	Consume(http.ResponseWriter, *http.Request, string) (bool, error)
	GetNonce(*http.Request) (string, error)
	Provided(http.ResponseWriter, *http.Request) (bool, error)
}

func Nonced(resp http.ResponseWriter, req *http.Request,
	service NonceService) (err error) {
	ok, err := service.Provided(resp, req)
	if err != nil {
		return err
	}
	if !ok {
		err = service.Block(resp, req)
		if err != nil {
			return err
		}
		return nil
	}
	nonce, err := service.GetNonce(req)
	if err != nil {
		return err
	}

	ok, err = service.Consume(resp, req, nonce)
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
