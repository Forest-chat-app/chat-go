package common

type ServiceErr struct {
	ResponseCode ResponseCode
}

func (e ServiceErr) Error() string {
	return e.ResponseCode.Msg
}

func (e ServiceErr) GetResponseCode() ResponseCode {
	return e.ResponseCode
}

func NewServiceError(code ResponseCode) ServiceErr {
	return ServiceErr{
		ResponseCode: code,
	}
}
