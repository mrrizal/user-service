package handler

import (
	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/SawitProRecruitment/UserService/utils"
)

type Server struct {
	Repository repository.RepositoryInterface
	Service    Service
	Utils      utils.Utils
}

type NewServerOptions struct {
	Repository repository.RepositoryInterface
	Service    Service
}

func NewServer(opts NewServerOptions) *Server {
	optsValidator := NewValidatorOptions{
		Repository: opts.Repository,
	}

	optsService := NewServiceOptions{
		Repository: opts.Repository,
		Validator:  NewValidator(optsValidator),
		Utils:      utils.NewUtils(),
	}

	service := NewService(optsService)
	return &Server{opts.Repository, service, optsService.Utils}
}
