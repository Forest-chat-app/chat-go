package service

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	ChatService
	UserService
	MongoToEsSync
	TokenService
}
